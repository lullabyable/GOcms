package collect

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-resty/resty/v2"
	"gorm.io/gorm"
)

// ConcurrentEngine 高并发采集引擎
type ConcurrentEngine struct {
	db        *gorm.DB
	client    *resty.Client
	config    CollectConfig
	jobs      sync.Map // map[string]*CollectJob
	progress  sync.Map // map[string]*CollectProgress
	mu        sync.RWMutex
}

func NewConcurrentEngine(db *gorm.DB, cfg CollectConfig) *ConcurrentEngine {
	if cfg.Workers <= 0 {
		cfg.Workers = 5
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 50
	}
	if cfg.RateLimit <= 0 {
		cfg.RateLimit = 10
	}
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 500
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30
	}

	return &ConcurrentEngine{
		db:     db,
		client: resty.New().SetTimeout(time.Duration(cfg.Timeout) * time.Second),
		config: cfg,
	}
}

// UpdateConfig 热更新配置
func (e *ConcurrentEngine) UpdateConfig(cfg CollectConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config = cfg
}

// GetConfig 获取当前配置
func (e *ConcurrentEngine) GetConfig() CollectConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.config
}

// GetJob 获取任务状态
func (e *ConcurrentEngine) GetJob(jobID string) *CollectJob {
	if v, ok := e.jobs.Load(jobID); ok {
		return v.(*CollectJob)
	}
	return nil
}

// ListJobs 列出所有任务
func (e *ConcurrentEngine) ListJobs() []*CollectJob {
	var list []*CollectJob
	e.jobs.Range(func(_, v interface{}) bool {
		list = append(list, v.(*CollectJob))
		return true
	})
	return list
}

// StopJob 停止任务
func (e *ConcurrentEngine) StopJob(jobID string) {
	if v, ok := e.jobs.Load(jobID); ok {
		job := v.(*CollectJob)
		job.Status = "paused"
	}
}

// ============ 核心采集流程 ============

// CollectFromSource 从资源站采集入口（异步）
func (e *ConcurrentEngine) CollectFromSource(ctx context.Context, source CollectSource, opts CollectOptions) *CollectJob {
	jobID := fmt.Sprintf("collect_%d_%d", source.ID, time.Now().UnixMilli())
	job := &CollectJob{
		ID:         jobID,
		SourceID:   source.ID,
		SourceName: source.Name,
		Mid:        source.Mid,
		Status:     "running",
		StartedAt:  time.Now(),
	}
	e.jobs.Store(jobID, job)

	// 异步执行
	go e.runCollect(ctx, source, opts, job)

	return job
}

// runCollect 执行采集主流程
func (e *ConcurrentEngine) runCollect(ctx context.Context, source CollectSource, opts CollectOptions, job *CollectJob) {
	defer func() {
		if r := recover(); r != nil {
			job.Status = "error"
			job.Message = fmt.Sprintf("panic: %v", r)
			job.EndedAt = time.Now()
		}
	}()

	// 第一步：获取远程分类和数据列表
	categories, videos, pageInfo, err := e.fetchRemoteData(ctx, source, opts)
	if err != nil {
		job.Status = "error"
		job.Message = err.Error()
		job.EndedAt = time.Now()
		return
	}

	job.Total = pageInfo.Total
	job.TotalPages = pageInfo.PageCount
	_ = categories // 分类绑定信息，后续可返回给前端

	// 第二步：获取本地分类映射
	bindMap := e.getBindMap(source.ID)

	// 第三步：并发处理视频数据
	e.processItems(ctx, source, videos, bindMap, job)

	// 多页采集：翻页继续
	if opts.Page == 0 {
		opts.Page = 1
	}
	for page := opts.Page + 1; page <= pageInfo.PageCount; page++ {
		// 检查暂停/取消
		if job.Status == "paused" || job.Status == "error" {
			break
		}

		opts.Page = page
		job.Page = page

		_, videos, _, err = e.fetchRemoteData(ctx, source, opts)
		if err != nil {
			job.Errors++
			continue
		}
		e.processItems(ctx, source, videos, bindMap, job)
	}

	if job.Status == "running" {
		job.Status = "done"
	}
	job.EndedAt = time.Now()

	// 保存断点
	e.saveProgress(source.ID, source.Mid, job.Page, 0)
}

// fetchRemoteData 从远程API获取数据
func (e *ConcurrentEngine) fetchRemoteData(ctx context.Context, source CollectSource, opts CollectOptions) (
	categories []RemoteCategory, videos []RemoteVideo, pageInfo PageResult, err error) {

	e.mu.RLock()
	timeout := e.config.Timeout
	e.mu.RUnlock()

	client := resty.New().SetTimeout(time.Duration(timeout) * time.Second)

	// 构建请求参数
	params := map[string]string{}
	if opts.Page > 0 {
		params["pg"] = fmt.Sprintf("%d", opts.Page)
	}
	if opts.TypeID > 0 {
		params["t"] = fmt.Sprintf("%d", opts.TypeID)
	}
	if opts.Hours > 0 {
		params["h"] = fmt.Sprintf("%d", opts.Hours)
	}
	if opts.Keyword != "" {
		params["wd"] = opts.Keyword
	}
	if opts.IDs != "" {
		params["ids"] = opts.IDs
	}

	// 附加参数
	if source.Param != "" {
		for _, p := range strings.Split(source.Param, "&") {
			if kv := strings.SplitN(p, "=", 2); len(kv) == 2 {
				params[kv[0]] = kv[1]
			}
		}
	}

	var resp *resty.Response
	url := strings.TrimRight(source.APIURL, "/")

	// 根据 collect_type 决定请求方式
	if source.Type == 1 {
		// XML 协议
		params["ac"] = "list"
		resp, err = client.R().SetQueryParams(params).SetHeader("User-Agent", "MacCMS/10.0").Get(url)
		if err != nil {
			return nil, nil, pageInfo, fmt.Errorf("请求失败: %w", err)
		}
		categories, videos, pageInfo, err = e.parseXMLResponse(resp.String())
	} else {
		// JSON 协议（默认）
		params["ac"] = "detail"
		resp, err = client.R().SetQueryParams(params).SetHeader("User-Agent", "MacCMS/10.0").Get(url)
		if err != nil {
			return nil, nil, pageInfo, fmt.Errorf("请求失败: %w", err)
		}
		categories, videos, pageInfo, err = e.parseJSONResponse(resp.String())
	}

	if err != nil {
		return nil, nil, pageInfo, fmt.Errorf("解析失败: %w", err)
	}

	// 补充分类信息（JSON协议 ac=list 时才有分类）
	if len(categories) == 0 && source.Type != 1 {
		// 用 ac=list 再请求一次获取分类
		params["ac"] = "list"
		delete(params, "pg")
		resp2, err := client.R().SetQueryParams(params).SetHeader("User-Agent", "MacCMS/10.0").Get(url)
		if err == nil {
			var apiResp APIResponse
			if json.Unmarshal([]byte(resp2.String()), &apiResp) == nil {
				categories = apiResp.Class
			}
		}
	}

	return categories, videos, pageInfo, nil
}

// PageResult 分页信息
type PageResult struct {
	Page       int
	PageCount  int
	PageSize   int
	Total      int
}

// parseJSONResponse 解析JSON响应
func (e *ConcurrentEngine) parseJSONResponse(body string) ([]RemoteCategory, []RemoteVideo, PageResult, error) {
	var resp APIResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return nil, nil, PageResult{}, err
	}

	page := PageResult{
		Page:      resp.Page,
		PageCount: resp.PageCount,
		PageSize:  resp.Limit,
		Total:     resp.Total,
	}

	var videos []RemoteVideo
	for _, item := range resp.List {
		v := RemoteVideo{
			VodID:        getInt(item, "vod_id"),
			TypeID:       getInt(item, "type_id"),
			TypeName:     getString(item, "type_name"),
			VodName:      getString(item, "vod_name"),
			VodSub:       getString(item, "vod_sub"),
			VodEn:        getString(item, "vod_en"),
			VodPic:       getString(item, "vod_pic"),
			VodActor:     getString(item, "vod_actor"),
			VodDirector:  getString(item, "vod_director"),
			VodWriter:    getString(item, "vod_writer"),
			VodArea:      getString(item, "vod_area"),
			VodLang:      getString(item, "vod_lang"),
			VodYear:      getString(item, "vod_year"),
			VodContent:   getString(item, "vod_content"),
			VodBlurb:     getString(item, "vod_blurb"),
			VodRemarks:   getString(item, "vod_remarks"),
			VodClass:     getString(item, "vod_class"),
			VodTag:       getString(item, "vod_tag"),
			VodState:     getString(item, "vod_state"),
			VodVersion:   getString(item, "vod_version"),
			VodPlayFrom:  getString(item, "vod_play_from"),
			VodPlayURL:   getString(item, "vod_play_url"),
			VodDownFrom:  getString(item, "vod_down_from"),
			VodDownURL:   getString(item, "vod_down_url"),
			VodScore:     getString(item, "vod_score"),
			VodScoreAll:  getInt(item, "vod_score_all"),
			VodScoreNum:  getInt(item, "vod_score_num"),
			VodHits:      getInt(item, "vod_hits"),
			VodHitsDay:   getInt(item, "vod_hits_day"),
			VodHitsWeek:  getInt(item, "vod_hits_week"),
			VodHitsMonth: getInt(item, "vod_hits_month"),
			VodTime:      getString(item, "vod_time"),
			VodPlot:      getInt(item, "vod_plot"),
			VodPlotName:  getString(item, "vod_plot_name"),
			VodPlotDetail: getString(item, "vod_plot_detail"),
		}
		videos = append(videos, v)
	}

	return resp.Class, videos, page, nil
}

// parseXMLResponse 解析XML响应（兼容maccms XML协议）
func (e *ConcurrentEngine) parseXMLResponse(body string) ([]RemoteCategory, []RemoteVideo, PageResult, error) {
	type XMLDD struct {
		Flag string `xml:"flag,attr"`
		Text string `xml:",chardata"`
	}
	type XMLVideo struct {
		ID       int    `xml:"id"`
		TID      int    `xml:"tid"`
		Name     string `xml:"name"`
		Pic      string `xml:"pic"`
		Lang     string `xml:"lang"`
		Area     string `xml:"area"`
		Year     string `xml:"year"`
		State    string `xml:"state"`
		Remarks  string `xml:"remarks"`
		Des      string `xml:"des"`
		Actor    string `xml:"actor"`
		Director string `xml:"director"`
		Last     string `xml:"last"`
		DL       struct {
			DD []XMLDD `xml:"dd"`
		} `xml:"dl"`
	}
	type XMLType struct {
		ID   int    `xml:"id"`
		Name string `xml:"name"`
	}
	type XMLRSS struct {
		List struct {
			Video []XMLVideo `xml:"video"`
			Type  []XMLType  `xml:"ty"`
		} `xml:"list"`
		PageCount   int `xml:"pagecount"`
		RecordCount int `xml:"recordcount"`
	}

	var rss XMLRSS
	if err := xml.Unmarshal([]byte(body), &rss); err != nil {
		return nil, nil, PageResult{}, err
	}

	page := PageResult{
		Page:      1,
		PageCount: rss.PageCount,
		Total:     rss.RecordCount,
	}

	var categories []RemoteCategory
	for _, t := range rss.List.Type {
		categories = append(categories, RemoteCategory{
			TypeID:   t.ID,
			TypeName: t.Name,
		})
	}

	var videos []RemoteVideo
	for _, v := range rss.List.Video {
		video := RemoteVideo{
			VodID:      v.ID,
			TypeID:     v.TID,
			VodName:    v.Name,
			VodPic:     v.Pic,
			VodLang:    v.Lang,
			VodArea:    v.Area,
			VodYear:    v.Year,
			VodState:   v.State,
			VodRemarks: v.Remarks,
			VodContent: v.Des,
			VodActor:   v.Actor,
			VodDirector: v.Director,
			VodTime:    v.Last,
		}
		// 解析播放地址
		var froms, urls []string
		for _, dd := range v.DL.DD {
			froms = append(froms, dd.Flag)
			urls = append(urls, e.formatPlayURL(dd.Text))
		}
		video.VodPlayFrom = strings.Join(froms, "$$$")
		video.VodPlayURL = strings.Join(urls, "$$$")
		videos = append(videos, video)
	}

	return categories, videos, page, nil
}

// formatPlayURL 格式化播放地址
func (e *ConcurrentEngine) formatPlayURL(raw string) string {
	raw = strings.ReplaceAll(raw, "||", "#")
	raw = strings.ReplaceAll(raw, "$$", "#")
	return raw
}

// getBindMap 获取分类绑定映射
func (e *ConcurrentEngine) getBindMap(sourceID int) map[int]int {
	var binds []CollectBind
	e.db.Where("collect_flag LIKE ?", fmt.Sprintf("%d_%%", sourceID)).Find(&binds)

	m := make(map[int]int)
	for _, b := range binds {
		m[b.RemoteTypeID] = b.LocalTypeID
	}
	return m
}

// processItems 并发处理视频列表
func (e *ConcurrentEngine) processItems(ctx context.Context, source CollectSource, videos []RemoteVideo, bindMap map[int]int, job *CollectJob) {
	if len(videos) == 0 {
		return
	}

	e.mu.RLock()
	workers := e.config.Workers
	batchSize := e.config.BatchSize
	rateLimit := e.config.RateLimit
	e.mu.RUnlock()

	// 限速器：每秒最多 rateLimit 个请求
	rateLimiter := time.NewTicker(time.Second / time.Duration(rateLimit))
	defer rateLimiter.Stop()

	// batchWrite 用于未来批量写入优化
	_ = batchSize

	// 任务通道（有界缓冲，防 OOM）
	taskCh := make(chan RemoteVideo, e.config.BufferSize)

	// 结果通道
	type processResult struct {
		action  string // imported/updated/skipped/error
		errMsg  string
	}
	resultCh := make(chan processResult, e.config.BufferSize)

	// 启动 worker pool
	var wg sync.WaitGroup
	var activeWorkers int32

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			atomic.AddInt32(&activeWorkers, 1)
			defer atomic.AddInt32(&activeWorkers, -1)

			for video := range taskCh {
				// 检查上下文取消
				select {
				case <-ctx.Done():
					resultCh <- processResult{action: "error", errMsg: "cancelled"}
					continue
				default:
				}

				// 检查任务暂停
				if job.Status == "paused" || job.Status == "error" {
					resultCh <- processResult{action: "skipped"}
					continue
				}

				// 限速
				<-rateLimiter.C

				// 映射分类
				if localID, ok := bindMap[video.TypeID]; ok {
					video.TypeID = localID
				} else if video.TypeID > 0 {
					// 未绑定的分类，跳过
					resultCh <- processResult{action: "skipped"}
					continue
				}

				// 处理单条视频
				action, err := e.processOneVideo(video, source)
				if err != nil {
					resultCh <- processResult{action: "error", errMsg: err.Error()}
				} else {
					resultCh <- processResult{action: action}
				}
			}
		}(i)
	}

	// 分发任务到通道（生产者）
	go func() {
		for _, v := range videos {
			select {
			case <-ctx.Done():
				break
			case taskCh <- v:
			}
		}
		close(taskCh)
	}()

	// 收集结果（消费者）
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// 批量统计
	for result := range resultCh {
		job.Current++
		switch result.action {
		case "imported":
			job.Imported++
		case "updated":
			job.Updated++
		case "skipped":
			job.Skipped++
		case "error":
			job.Errors++
		}
	}
}

// processOneVideo 处理单条视频（去重+入库）
func (e *ConcurrentEngine) processOneVideo(video RemoteVideo, source CollectSource) (string, error) {
	e.mu.RLock()
	cfg := e.config
	e.mu.RUnlock()

	// 过滤检查
	if cfg.Filter != "" {
		keywords := strings.Split(cfg.Filter, ",")
		for _, kw := range keywords {
			if kw = strings.TrimSpace(kw); kw != "" && strings.Contains(video.VodName, kw) {
				return "skipped", nil
			}
		}
	}

	// 年份过滤
	if source.FilterYear != "" {
		years := strings.Split(source.FilterYear, ",")
		matched := false
		for _, y := range years {
			if strings.TrimSpace(y) == video.VodYear {
				matched = true
				break
			}
		}
		if !matched {
			return "skipped", nil
		}
	}

	// 去重检查（根据 InRule）
	where := e.buildDedupWhere(video, cfg.InRule)
	if len(where) == 0 {
		where["vod_name"] = video.VodName
	}

	var existing struct {
		VodID int `gorm:"column:vod_id"`
	}
	query := e.db.Table("mac_vod")
	for k, v := range where {
		query = query.Where(k+" = ?", v)
	}
	result := query.Select("vod_id").First(&existing)

	if result.Error == gorm.ErrRecordNotFound {
		// 新增
		if source.Opt == 2 { // 仅更新模式，跳过新增
			return "skipped", nil
		}
		return "imported", e.insertVideo(video, cfg, source)
	}
	if result.Error != nil {
		return "error", result.Error
	}

	// 已存在
	if source.Opt == 1 { // 仅新增模式，跳过更新
		return "skipped", nil
	}

	// 更新
	return "updated", e.updateVideo(existing.VodID, video, cfg, source)
}

// buildDedupWhere 构建去重查询条件
func (e *ConcurrentEngine) buildDedupWhere(video RemoteVideo, inRule string) map[string]interface{} {
	where := make(map[string]interface{})

	if strings.Contains(inRule, "a") {
		where["vod_name"] = video.VodName
	}
	if strings.Contains(inRule, "b") && video.TypeID > 0 {
		where["type_id"] = video.TypeID
	}
	if strings.Contains(inRule, "c") && video.VodYear != "" {
		where["vod_year"] = video.VodYear
	}
	if strings.Contains(inRule, "d") && video.VodArea != "" {
		where["vod_area"] = video.VodArea
	}
	if strings.Contains(inRule, "e") && video.VodLang != "" {
		where["vod_lang"] = video.VodLang
	}
	if strings.Contains(inRule, "f") && video.VodActor != "" {
		where["vod_actor"] = video.VodActor
	}
	if strings.Contains(inRule, "g") && video.VodDirector != "" {
		where["vod_director"] = video.VodDirector
	}

	return where
}

// insertVideo 新增视频
func (e *ConcurrentEngine) insertVideo(video RemoteVideo, cfg CollectConfig, source CollectSource) error {
	now := time.Now().Unix()

	vod := map[string]interface{}{
		"type_id":        video.TypeID,
		"vod_name":       video.VodName,
		"vod_sub":        video.VodSub,
		"vod_en":         video.VodEn,
		"vod_pic":        video.VodPic,
		"vod_actor":      video.VodActor,
		"vod_director":   video.VodDirector,
		"vod_writer":     video.VodWriter,
		"vod_area":       video.VodArea,
		"vod_lang":       video.VodLang,
		"vod_year":       video.VodYear,
		"vod_content":    video.VodContent,
		"vod_blurb":      video.VodBlurb,
		"vod_remarks":    video.VodRemarks,
		"vod_class":      video.VodClass,
		"vod_tag":        video.VodTag,
		"vod_state":      video.VodState,
		"vod_version":    video.VodVersion,
		"vod_play_from":  video.VodPlayFrom,
		"vod_play_url":   video.VodPlayURL,
		"vod_down_from":  video.VodDownFrom,
		"vod_down_url":   video.VodDownURL,
		"vod_plot":       video.VodPlot,
		"vod_plot_name":  video.VodPlotName,
		"vod_plot_detail": video.VodPlotDetail,
		"vod_status":     cfg.Status,
		"vod_time_add":   now,
		"vod_time":       video.VodTime,
	}

	// 随机数据
	if cfg.HitsStart > 0 && cfg.HitsEnd > cfg.HitsStart {
		hits := rand.Intn(cfg.HitsEnd-cfg.HitsStart) + cfg.HitsStart
		vod["vod_hits"] = hits
		vod["vod_hits_day"] = hits
		vod["vod_hits_week"] = hits
		vod["vod_hits_month"] = hits
	}
	if cfg.UpdownStart > 0 && cfg.UpdownEnd > cfg.UpdownStart {
		vod["vod_up"] = rand.Intn(cfg.UpdownEnd-cfg.UpdownStart) + cfg.UpdownStart
		vod["vod_down"] = rand.Intn(cfg.UpdownEnd-cfg.UpdownStart) + cfg.UpdownStart
	}
	if cfg.ScoreRandom == 1 {
		num := rand.Intn(1000) + 1
		all := num * (rand.Intn(10) + 1)
		vod["vod_score_num"] = num
		vod["vod_score_all"] = all
		vod["vod_score"] = fmt.Sprintf("%.1f", float64(all)/float64(num))
	}

	return e.db.Table("mac_vod").Create(vod).Error
}

// updateVideo 更新视频
func (e *ConcurrentEngine) updateVideo(vodID int, video RemoteVideo, cfg CollectConfig, source CollectSource) error {
	updates := map[string]interface{}{}

	// 根据 UpRule 决定更新哪些字段
	if strings.Contains(cfg.UpRule, "a") { // 播放地址
		updates["vod_play_from"] = video.VodPlayFrom
		updates["vod_play_url"] = video.VodPlayURL
	}
	if strings.Contains(cfg.UpRule, "b") { // 下载地址
		updates["vod_down_from"] = video.VodDownFrom
		updates["vod_down_url"] = video.VodDownURL
	}
	if strings.Contains(cfg.UpRule, "d") { // 备注
		updates["vod_remarks"] = video.VodRemarks
	}
	if strings.Contains(cfg.UpRule, "e") { // 导演
		updates["vod_director"] = video.VodDirector
	}
	if strings.Contains(cfg.UpRule, "f") { // 演员
		updates["vod_actor"] = video.VodActor
	}
	if strings.Contains(cfg.UpRule, "g") { // 年份
		updates["vod_year"] = video.VodYear
	}
	if strings.Contains(cfg.UpRule, "h") { // 地区
		updates["vod_area"] = video.VodArea
	}
	if strings.Contains(cfg.UpRule, "i") { // 语言
		updates["vod_lang"] = video.VodLang
	}
	if strings.Contains(cfg.UpRule, "j") { // 图片
		updates["vod_pic"] = video.VodPic
	}
	if strings.Contains(cfg.UpRule, "k") { // 内容
		updates["vod_content"] = video.VodContent
	}
	if strings.Contains(cfg.UpRule, "l") { // 标签
		updates["vod_tag"] = video.VodTag
	}

	if len(updates) == 0 {
		// 默认更新播放地址
		updates["vod_play_from"] = video.VodPlayFrom
		updates["vod_play_url"] = video.VodPlayURL
		updates["vod_remarks"] = video.VodRemarks
	}

	updates["vod_time"] = video.VodTime

	return e.db.Table("mac_vod").Where("vod_id = ?", vodID).Updates(updates).Error
}

// saveProgress 保存断点
func (e *ConcurrentEngine) saveProgress(sourceID, mid, lastPage, lastID int) {
	key := fmt.Sprintf("%d_%d", sourceID, mid)
	e.progress.Store(key, CollectProgress{
		SourceID:  sourceID,
		Mid:       mid,
		LastPage:  lastPage,
		LastID:    lastID,
		Timestamp: time.Now().Unix(),
	})
}

// GetProgress 获取断点
func (e *ConcurrentEngine) GetProgress(sourceID, mid int) *CollectProgress {
	key := fmt.Sprintf("%d_%d", sourceID, mid)
	if v, ok := e.progress.Load(key); ok {
		p := v.(CollectProgress)
		return &p
	}
	return nil
}

// ClearProgress 清除断点
func (e *ConcurrentEngine) ClearProgress(sourceID, mid int) {
	key := fmt.Sprintf("%d_%d", sourceID, mid)
	e.progress.Delete(key)
}

// FetchRemoteData 公开方法：获取远程数据（供 handler 直接调用）
func (e *ConcurrentEngine) FetchRemoteData(source CollectSource, opts CollectOptions) ([]RemoteCategory, []RemoteVideo, PageResult, error) {
	return e.fetchRemoteData(context.Background(), source, opts)
}

// TestConnection 测试资源站连通性
func (e *ConcurrentEngine) TestConnection(apiURL string) (string, int, error) {
	client := resty.New().SetTimeout(15 * time.Second)

	// 先尝试 JSON
	resp, err := client.R().
		SetQueryParams(map[string]string{"ac": "detail", "pg": "1"}).
		SetHeader("User-Agent", "MacCMS/10.0").
		Get(apiURL)
	if err != nil {
		return "", 0, err
	}

	body := resp.String()

	// 尝试 JSON
	var jsonResp APIResponse
	if err := json.Unmarshal([]byte(body), &jsonResp); err == nil && len(jsonResp.List) > 0 {
		return "json", jsonResp.Total, nil
	}

	// 尝试 XML
	type XMLRSS struct {
		List struct {
			Video []struct{} `xml:"video"`
		} `xml:"list"`
		RecordCount int `xml:"recordcount"`
	}
	var rss XMLRSS
	if err := xml.Unmarshal([]byte(body), &rss); err == nil {
		return "xml", rss.RecordCount, nil
	}

	return "", 0, fmt.Errorf("无法解析响应，请检查API地址")
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int:
			return val
		case int64:
			return int(val)
		}
	}
	return 0
}

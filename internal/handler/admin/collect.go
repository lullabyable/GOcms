package admin

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/response"
	"gocms/internal/service/collect"
	"gorm.io/gorm"
)

// CollectHandler 后台采集管理
type CollectHandler struct {
	db     *gorm.DB
	engine *collect.ConcurrentEngine
}

func NewCollectHandler(db *gorm.DB, engine *collect.ConcurrentEngine) *CollectHandler {
	return &CollectHandler{db: db, engine: engine}
}

// ==================== 资源站管理 ====================

// SourceList 资源站列表
func (h *CollectHandler) SourceList(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if limit < 1 {
		limit = 20
	}

	var total int64
	h.db.Model(&collect.CollectSource{}).Count(&total)

	var list []collect.CollectSource
	h.db.Order("collect_id DESC").Offset((page - 1) * limit).Limit(limit).Find(&list)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{"list": list, "total": total, "page": page, "limit": limit},
	})
}

// SourceDetail 资源站详情
func (h *CollectHandler) SourceDetail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var source collect.CollectSource
	if err := h.db.First(&source, id).Error; err != nil {
		return response.Fail(c, "资源站不存在")
	}
	return response.OK(c, source)
}

// SourceSave 保存资源站
func (h *CollectHandler) SourceSave(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.FormValue("collect_id"))
	name := c.FormValue("collect_name")
	apiURL := c.FormValue("collect_url")
	if name == "" || apiURL == "" {
		return response.Fail(c, "名称和API地址不能为空")
	}

	source := collect.CollectSource{
		Name:       name,
		APIURL:     apiURL,
		Type:       formInt(c, "collect_type", 2),
		Mid:        formInt(c, "collect_mid", 1),
		AppID:      c.FormValue("collect_appid"),
		AppKey:     c.FormValue("collect_appkey"),
		Param:      c.FormValue("collect_param"),
		Filter:     formInt(c, "collect_filter", 0),
		FilterFrom: c.FormValue("collect_filter_from"),
		FilterYear: c.FormValue("collect_filter_year"),
		Opt:        formInt(c, "collect_opt", 0),
		SyncPicOpt: formInt(c, "collect_sync_pic_opt", 0),
	}

	if id > 0 {
		source.ID = id
		h.db.Model(&collect.CollectSource{}).Where("collect_id = ?", id).Updates(map[string]interface{}{
			"collect_name":          source.Name,
			"collect_url":           source.APIURL,
			"collect_type":          source.Type,
			"collect_mid":           source.Mid,
			"collect_appid":         source.AppID,
			"collect_appkey":        source.AppKey,
			"collect_param":         source.Param,
			"collect_filter":        source.Filter,
			"collect_filter_from":   source.FilterFrom,
			"collect_filter_year":   source.FilterYear,
			"collect_opt":           source.Opt,
			"collect_sync_pic_opt":  source.SyncPicOpt,
		})
	} else {
		source.CreatedAt = time.Now().Unix()
		h.db.Create(&source)
	}

	return response.OKMsg(c, "保存成功")
}

// SourceDelete 删除资源站
func (h *CollectHandler) SourceDelete(c *fiber.Ctx) error {
	ids := c.FormValue("ids")
	if ids == "" {
		return response.Fail(c, "请选择要删除的资源站")
	}
	idList := parseIDList(ids)
	h.db.Delete(&collect.CollectSource{}, "collect_id IN ?", idList)
	// 同时删除绑定
	h.db.Delete(&collect.CollectBind{}, "collect_flag LIKE ?", "%%") // TODO: 按 sourceID 精确删除
	return response.OKMsg(c, "删除成功")
}

// TestConnection 测试资源站连通性
func (h *CollectHandler) TestConnection(c *fiber.Ctx) error {
	apiURL := c.FormValue("api_url")
	if apiURL == "" {
		return response.Fail(c, "请输入API地址")
	}

	format, total, err := h.engine.TestConnection(apiURL)
	if err != nil {
		return response.Fail(c, "连接失败: "+err.Error())
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"msg":  "连接成功",
		"data": fiber.Map{"format": format, "total": total},
	})
}

// ==================== 采集操作 ====================

// CollectStart 开始采集
func (h *CollectHandler) CollectStart(c *fiber.Ctx) error {
	sourceID, _ := strconv.Atoi(c.FormValue("source_id"))
	ac := c.FormValue("ac", "cjall")    // cjsel/cjday/cjall
	hours, _ := strconv.Atoi(c.FormValue("h"))
	keyword := c.FormValue("wd")

	var source collect.CollectSource
	if err := h.db.First(&source, sourceID).Error; err != nil {
		return response.Fail(c, "资源站不存在")
	}

	opts := collect.CollectOptions{}
	switch ac {
	case "cjday":
		opts.Hours = 24
	case "cjweek":
		opts.Hours = 168
	case "cjall":
		// 不设限制
	}
	if hours > 0 {
		opts.Hours = hours
	}
	if keyword != "" {
		opts.Keyword = keyword
	}

	ctx := context.Background()
	job := h.engine.CollectFromSource(ctx, source, opts)

	return c.JSON(fiber.Map{
		"code": 1,
		"msg":  "采集任务已启动",
		"data": job,
	})
}

// CollectAPI 采集API入口（兼容maccms10的collect/api路由）
// 支持 ac=list (获取远程列表) 和 ac=cj (执行采集)
func (h *CollectHandler) CollectAPI(c *fiber.Ctx) error {
	cjflag := c.Query("cjflag")
	cjurl := c.Query("cjurl")
	ac := c.Query("ac", "list") // list/cj/cjsel/cjday/cjall
	mid := c.Query("mid", "1")
	ctype := c.Query("type", "2") // 1=xml 2=json

	if cjurl == "" {
		return response.Fail(c, "缺少采集地址")
	}

	source := collect.CollectSource{
		APIURL: cjurl,
		Type:   mustInt(ctype),
		Mid:    mustInt(mid),
		Param:  c.Query("param"),
	}

	_ = cjflag // TODO: 用 cjflag 做安全校验

	// 获取本地分类（用于绑定）
	var types []struct {
		TypeID   int    `json:"type_id"`
		TypeName string `json:"type_name"`
		TypeMid  int    `json:"type_mid"`
	}
	h.db.Table("mac_type").Select("type_id, type_name, type_mid").Find(&types)

	if ac == "list" {
		// 获取远程列表
		opts := collect.CollectOptions{
			Page:    queryInt(c, "pg", 1),
			TypeID:  queryInt(c, "t", 0),
			Hours:   queryInt(c, "h", 0),
			Keyword: c.Query("wd"),
			IDs:     c.Query("ids"),
		}

		categories, videos, pageInfo, err := h.engine.FetchRemoteData(source, opts)
		if err != nil {
			return response.Fail(c, err.Error())
		}

		// 构建绑定信息
		bindMap := h.getBindMap(source.ID)
		for i, cat := range categories {
			if localID, ok := bindMap[cat.TypeID]; ok {
				categories[i].IsBind = true
				categories[i].LocalID = localID
				// 查本地分类名
				for _, t := range types {
					if t.TypeID == localID {
						categories[i].LocalName = t.TypeName
						break
					}
				}
			}
		}

		return c.JSON(fiber.Map{
			"code": 1,
			"data": fiber.Map{
				"type":  categories,
				"list":  videos,
				"page":  pageInfo,
				"total": pageInfo.Total,
			},
		})
	}

	// ac=cj/cjsel/cjday/cjall - 执行采集
	opts := collect.CollectOptions{
		Page:   queryInt(c, "pg", 1),
		Hours:  queryInt(c, "h", 0),
		Keyword: c.Query("wd"),
		IDs:    c.Query("ids"),
	}

	ctx := context.Background()
	job := h.engine.CollectFromSource(ctx, source, opts)

	return c.JSON(fiber.Map{
		"code": 1,
		"msg":  "采集已启动",
		"data": job,
	})
}

// ==================== 分类绑定 ====================

// BindList 获取分类绑定列表
func (h *CollectHandler) BindList(c *fiber.Ctx) error {
	sourceID, _ := strconv.Atoi(c.Query("source_id"))
	var binds []collect.CollectBind
	h.db.Where("collect_flag LIKE ?", strconv.Itoa(sourceID)+"_%").Find(&binds)

	// 获取本地分类
	var types []struct {
		TypeID   int    `json:"type_id"`
		TypeName string `json:"type_name"`
	}
	h.db.Table("mac_type").Select("type_id, type_name").Find(&types)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{"binds": binds, "types": types},
	})
}

// BindSave 保存分类绑定
func (h *CollectHandler) BindSave(c *fiber.Ctx) error {
	sourceID := c.FormValue("source_id")
	remoteTypeID, _ := strconv.Atoi(c.FormValue("remote_type_id"))
	remoteName := c.FormValue("remote_name")
	localTypeID, _ := strconv.Atoi(c.FormValue("local_type_id"))

	if sourceID == "" || remoteTypeID == 0 {
		return response.Fail(c, "参数不完整")
	}

	flag := sourceID + "_" + strconv.Itoa(remoteTypeID)

	// 查本地分类名
	var localName string
	h.db.Table("mac_type").Where("type_id = ?", localTypeID).Select("type_name").Scan(&localName)

	bind := collect.CollectBind{
		CollectFlag:  flag,
		RemoteTypeID: remoteTypeID,
		RemoteName:   remoteName,
		LocalTypeID:  localTypeID,
		LocalName:    localName,
	}

	// Upsert
	var existing collect.CollectBind
	if h.db.Where("collect_flag = ?", flag).First(&existing).Error == nil {
		h.db.Model(&existing).Updates(map[string]interface{}{
			"local_type_id": localTypeID,
			"local_name":    localName,
		})
	} else {
		h.db.Create(&bind)
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"msg":  "绑定成功",
		"data": fiber.Map{"id": flag, "st": 1, "local_type_id": localTypeID, "local_name": localName},
	})
}

// ==================== 采集配置 ====================

// ConfigGet 获取采集配置
func (h *CollectHandler) ConfigGet(c *fiber.Ctx) error {
	cfg := h.engine.GetConfig()
	return response.OK(c, cfg)
}

// ConfigSave 保存采集配置
func (h *CollectHandler) ConfigSave(c *fiber.Ctx) error {
	var cfg collect.CollectConfig
	if err := c.BodyParser(&cfg); err != nil {
		return response.Fail(c, "参数错误")
	}
	h.engine.UpdateConfig(cfg)

	// 持久化到数据库
	// TODO: 保存到 mac_collect_config 表

	return response.OKMsg(c, "配置已保存")
}

// ==================== 任务管理 ====================

// JobList 任务列表
func (h *CollectHandler) JobList(c *fiber.Ctx) error {
	jobs := h.engine.ListJobs()
	return response.OK(c, jobs)
}

// JobStatus 任务状态
func (h *CollectHandler) JobStatus(c *fiber.Ctx) error {
	jobID := c.Query("job_id")
	job := h.engine.GetJob(jobID)
	if job == nil {
		return response.Fail(c, "任务不存在")
	}
	return response.OK(c, job)
}

// JobStop 停止任务
func (h *CollectHandler) JobStop(c *fiber.Ctx) error {
	jobID := c.FormValue("job_id")
	h.engine.StopJob(jobID)
	return response.OKMsg(c, "已停止")
}

// ==================== 采集视频列表（已入库） ====================

// VodList 已采集的视频列表
func (h *CollectHandler) VodList(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	keyword := c.Query("wd")

	type vodRow struct {
		VodID       int    `json:"vod_id"`
		VodName     string `json:"vod_name"`
		TypeName    string `json:"type_name"`
		VodPlayFrom string `json:"vod_play_from"`
		VodTime     string `json:"vod_time"`
		VodRemarks  string `json:"vod_remarks"`
		VodPic      string `json:"vod_pic"`
	}

	query := h.db.Table("mac_vod").
		Select("mac_vod.vod_id, mac_vod.vod_name, mac_type.type_name, mac_vod.vod_play_from, mac_vod.vod_time, mac_vod.vod_remarks, mac_vod.vod_pic").
		Joins("LEFT JOIN mac_type ON mac_type.type_id = mac_vod.type_id")

	if keyword != "" {
		query = query.Where("mac_vod.vod_name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var list []vodRow
	query.Order("mac_vod.vod_id DESC").Offset((page - 1) * limit).Limit(limit).Find(&list)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{"list": list, "total": total, "page": page, "limit": limit},
	})
}

// ==================== 辅助函数 ====================

func (h *CollectHandler) getBindMap(sourceID int) map[int]int {
	var binds []collect.CollectBind
	h.db.Where("collect_flag LIKE ?", strconv.Itoa(sourceID)+"_%").Find(&binds)
	m := make(map[int]int)
	for _, b := range binds {
		m[b.RemoteTypeID] = b.LocalTypeID
	}
	return m
}

func formInt(c *fiber.Ctx, key string, def int) int {
	v, _ := strconv.Atoi(c.FormValue(key))
	if v == 0 {
		return def
	}
	return v
}

func queryInt(c *fiber.Ctx, key string, def int) int {
	v, _ := strconv.Atoi(c.Query(key))
	if v == 0 {
		return def
	}
	return v
}

func mustInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

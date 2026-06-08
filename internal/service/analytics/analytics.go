package analytics

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gocms/internal/model"
)

// Service 数据分析服务
type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// RecordVisit 记录访问
func (s *Service) RecordVisit(url, ip, userAgent, referer string) {
	now := time.Now()
	date := now.Format("2006-01-02")

	// 检查是否为唯一访客（当天同一IP首次访问）
	var count int64
	s.db.Model(&model.Visit{}).Where("ip = ? AND date = ?", ip, date).Count(&count)
	isUnique := 0
	if count == 0 {
		isUnique = 1
	}

	visit := model.Visit{
		URL:       url,
		IP:        ip,
		UserAgent: userAgent,
		Referer:   referer,
		VisitTime: now.Unix(),
		Date:      date,
		IsUnique:  isUnique,
	}
	s.db.Create(&visit)
}

// Dashboard 仪表盘数据
type Dashboard struct {
	TodayPV      int64            `json:"today_pv"`
	TodayUV      int64            `json:"today_uv"`
	TodayIP      int64            `json:"today_ip"`
	TotalPV      int64            `json:"total_pv"`
	TotalVod     int64            `json:"total_vod"`
	TotalArt     int64            `json:"total_art"`
	TotalUser    int64            `json:"total_user"`
	WeekTrend    []DayStat        `json:"week_trend"`
	TopVods      []TopItem        `json:"top_vods"`
	TopArts      []TopItem        `json:"top_arts"`
	Sources      []SourceStat     `json:"sources"`
}

// DayStat 每日统计
type DayStat struct {
	Date string `json:"date"`
	PV   int    `json:"pv"`
	UV   int    `json:"uv"`
	IP   int    `json:"ip"`
}

// TopItem 热门内容项
type TopItem struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Hits  int    `json:"hits"`
}

// SourceStat 来源统计
type SourceStat struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}

// GetDashboard 获取仪表盘数据
func (s *Service) GetDashboard() (*Dashboard, error) {
	dash := &Dashboard{}
	today := time.Now().Format("2006-01-02")

	// 今日 PV/UV/IP
	s.db.Model(&model.Visit{}).Where("date = ?", today).Count(&dash.TodayPV)
	s.db.Model(&model.Visit{}).Where("date = ? AND is_unique = 1", today).Count(&dash.TodayUV)
	s.db.Model(&model.Visit{}).Where("date = ?", today).Distinct("ip").Count(&dash.TodayIP)

	// 总计
	s.db.Model(&model.Visit{}).Count(&dash.TotalPV)
	s.db.Model(&model.Vod{}).Count(&dash.TotalVod)
	s.db.Model(&model.Art{}).Count(&dash.TotalArt)
	s.db.Model(&model.User{}).Count(&dash.TotalUser)

	// 最近7天趋势
	dash.WeekTrend = s.getTrend(7)

	// 热门视频 Top10
	s.db.Model(&model.Vod{}).Select("vod_id as id, vod_name as title, vod_hits as hits").
		Order("vod_hits DESC").Limit(10).Scan(&dash.TopVods)

	// 热门文章 Top10
	s.db.Model(&model.Art{}).Select("art_id as id, art_name as title, art_hits as hits").
		Order("art_hits DESC").Limit(10).Scan(&dash.TopArts)

	// 来源分析
	dash.Sources = s.getSourceStats()

	return dash, nil
}

// GetTrend 获取趋势数据
func (s *Service) GetTrend(days int) []DayStat {
	return s.getTrend(days)
}

func (s *Service) getTrend(days int) []DayStat {
	var stats []DayStat
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	s.db.Model(&model.VisitStat{}).
		Where("date >= ?", startDate).
		Order("date ASC").
		Scan(&stats)

	// 补全没有数据的日期
	if len(stats) == 0 {
		for i := days; i >= 0; i-- {
			d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
			stats = append(stats, DayStat{Date: d})
		}
	}

	return stats
}

// GetTopContent 获取热门内容
func (s *Service) GetTopContent(contentType string, limit int) []TopItem {
	var items []TopItem
	switch contentType {
	case "vod":
		s.db.Model(&model.Vod{}).Select("vod_id as id, vod_name as title, vod_hits as hits").
			Order("vod_hits DESC").Limit(limit).Scan(&items)
	case "art":
		s.db.Model(&model.Art{}).Select("art_id as id, art_name as title, art_hits as hits").
			Order("art_hits DESC").Limit(limit).Scan(&items)
	}
	return items
}

// GetSourceStats 来源分析
func (s *Service) getSourceStats() []SourceStat {
	var stats []SourceStat
	today := time.Now().Format("2006-01-02")

	// 直接访问（无 referer）
	var direct int64
	s.db.Model(&model.Visit{}).Where("date = ? AND (referer = '' OR referer IS NULL)", today).Count(&direct)
	if direct > 0 {
		stats = append(stats, SourceStat{Source: "直接访问", Count: int(direct)})
	}

	// 搜索引擎
	searchEngines := []struct {
		name  string
		likes []string
	}{
		{"百度", []string{"%baidu.com%", "%baidu.cn%"}},
		{"谷歌", []string{"%google.%"}},
		{"搜狗", []string{"%sogou.com%"}},
		{"神马", []string{"%sm.cn%"}},
		{"必应", []string{"%bing.com%"}},
	}

	for _, se := range searchEngines {
		var count int64
		query := s.db.Model(&model.Visit{}).Where("date = ?", today)
		for i, like := range se.likes {
			if i == 0 {
				query = query.Where("referer LIKE ?", like)
			} else {
				query = query.Or("date = ? AND referer LIKE ?", today, like)
			}
		}
		query.Count(&count)
		if count > 0 {
			stats = append(stats, SourceStat{Source: se.name, Count: int(count)})
		}
	}

	// 其他外链
	var otherExternal int64
	s.db.Model(&model.Visit{}).
		Where("date = ? AND referer != '' AND referer IS NOT NULL", today).
		Count(&otherExternal)
	for _, s2 := range stats {
		otherExternal -= int64(s2.Count)
	}
	if otherExternal > 0 {
		stats = append(stats, SourceStat{Source: "其他外链", Count: int(otherExternal)})
	}

	return stats
}

// AggregateDaily 汇总每日统计（定时任务调用）
func (s *Service) AggregateDaily() error {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	var pv, uv, ip, newUsers int64
	s.db.Model(&model.Visit{}).Where("date = ?", yesterday).Count(&pv)
	s.db.Model(&model.Visit{}).Where("date = ? AND is_unique = 1", yesterday).Count(&uv)
	s.db.Model(&model.Visit{}).Where("date = ?", yesterday).Distinct("ip").Count(&ip)
	newUsers = uv // 简化处理

	stat := model.VisitStat{
		Date:     yesterday,
		PV:       int(pv),
		UV:       int(uv),
		IP:       int(ip),
		NewUsers: int(newUsers),
	}

	// 更新或插入
	var existing model.VisitStat
	if err := s.db.Where("date = ?", yesterday).First(&existing).Error; err != nil {
		return s.db.Create(&stat).Error
	}
	return s.db.Model(&existing).Updates(map[string]interface{}{
		"pv":        stat.PV,
		"uv":        stat.UV,
		"ip":        stat.IP,
		"new_users": stat.NewUsers,
	}).Error
}

// VisitMiddleware 访问记录中间件工厂
func (s *Service) VisitMiddleware() func(url, ip, userAgent, referer string) {
	return s.RecordVisit
}

// GetVisitList 获取访问记录列表
func (s *Service) GetVisitList(page, pageSize int, date string) ([]model.Visit, int64, error) {
	var visits []model.Visit
	var total int64
	query := s.db.Model(&model.Visit{})
	if date != "" {
		query = query.Where("date = ?", date)
	}
	query.Count(&total)
	err := query.Order("visit_time DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&visits).Error
	return visits, total, err
}

// GetRegionStats 地域分析（基于IP，简化实现）
func (s *Service) GetRegionStats(date string) []SourceStat {
	var stats []SourceStat
	query := s.db.Model(&model.Visit{})
	if date != "" {
		query = query.Where("date = ?", date)
	}

	// 按IP前缀做简单分组（实际项目应接入IP库）
	rows, _ := query.Select("SUBSTRING_INDEX(ip, '.', 2) as region, COUNT(*) as count").
		Group("SUBSTRING_INDEX(ip, '.', 2)").
		Order("count DESC").
		Limit(20).
		Rows()

	defer rows.Close()
	for rows.Next() {
		var stat SourceStat
		rows.Scan(&stat.Source, &stat.Count)
		if stat.Source != "" {
			stat.Source = fmt.Sprintf("%s.x.x", stat.Source)
			stats = append(stats, stat)
		}
	}
	return stats
}

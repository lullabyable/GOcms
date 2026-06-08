package admin

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/service/analytics"
)

// AnalyticsHandler 数据分析处理器
type AnalyticsHandler struct {
	analytics *analytics.Service
}

func NewAnalyticsHandler(svc *analytics.Service) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: svc}
}

// Dashboard 仪表盘数据
func (h *AnalyticsHandler) Dashboard(c *fiber.Ctx) error {
	data, err := h.analytics.GetDashboard()
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "获取数据失败"})
	}
	return c.JSON(fiber.Map{"code": 1, "data": data})
}

// Trend 趋势数据
func (h *AnalyticsHandler) Trend(c *fiber.Ctx) error {
	days, _ := strconv.Atoi(c.Query("days", "7"))
	if days <= 0 || days > 365 {
		days = 7
	}

	data := h.analytics.GetTrend(days)
	return c.JSON(fiber.Map{"code": 1, "data": data})
}

// TopContent 热门内容
func (h *AnalyticsHandler) TopContent(c *fiber.Ctx) error {
	contentType := c.Query("type", "vod")
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	data := h.analytics.GetTopContent(contentType, limit)
	return c.JSON(fiber.Map{"code": 1, "data": data})
}

// Regions 地域分析
func (h *AnalyticsHandler) Regions(c *fiber.Ctx) error {
	date := c.Query("date", "")
	data := h.analytics.GetRegionStats(date)
	return c.JSON(fiber.Map{"code": 1, "data": data})
}

// VisitList 访问记录列表
func (h *AnalyticsHandler) VisitList(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	date := c.Query("date", "")

	visits, total, err := h.analytics.GetVisitList(page, pageSize, date)
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "查询失败"})
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":      visits,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

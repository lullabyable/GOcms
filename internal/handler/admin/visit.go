package admin

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/model"
	"gocms/internal/response"
	"gorm.io/gorm"
)

// VisitHandler 访问日志处理器
type VisitHandler struct {
	db *gorm.DB
}

func NewVisitHandler(db *gorm.DB) *VisitHandler {
	return &VisitHandler{db: db}
}

// List 访问记录列表
func (h *VisitHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	date := c.Query("date", "")
	ip := c.Query("ip", "")

	query := h.db.Model(&model.Visit{})
	if date != "" {
		query = query.Where("date = ?", date)
	}
	if ip != "" {
		query = query.Where("ip = ?", ip)
	}

	var total int64
	query.Count(&total)

	var visits []model.Visit
	query.Order("visit_id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&visits)

	return response.Page(c, visits, total, page, pageSize)
}

// Stats 访问统计
func (h *VisitHandler) Stats(c *fiber.Ctx) error {
	days, _ := strconv.Atoi(c.Query("days", "7"))
	if days <= 0 || days > 90 {
		days = 7
	}

	var stats []model.VisitStat
	h.db.Order("date DESC").Limit(days).Find(&stats)

	// 反转为时间正序
	for i, j := 0, len(stats)-1; i < j; i, j = i+1, j-1 {
		stats[i], stats[j] = stats[j], stats[i]
	}

	// 今日汇总
	var today model.VisitStat
	h.db.Where("date = ?", c.Query("date", "")).First(&today)

	return response.OK(c, fiber.Map{
		"trend": stats,
		"today": today,
	})
}

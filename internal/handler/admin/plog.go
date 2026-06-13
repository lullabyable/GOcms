package admin

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/model"
	"gorm.io/gorm"
)

// PlogHandler 操作日志处理器
type PlogHandler struct {
	db *gorm.DB
}

func NewPlogHandler(db *gorm.DB) *PlogHandler {
	return &PlogHandler{db: db}
}

// List 操作日志列表
func (h *PlogHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	keyword := c.Query("keyword", "")
	adminName := c.Query("admin_name", "")

	query := h.db.Model(&model.Plog{})
	if keyword != "" {
		query = query.Where("plog_content LIKE ?", "%"+keyword+"%")
	}
	if adminName != "" {
		query = query.Where("admin_name LIKE ?", "%"+adminName+"%")
	}

	var total int64
	query.Count(&total)

	var plogs []model.Plog
	query.Order("plog_id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&plogs)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":      plogs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

package frontend

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

type GbookHandler struct{ db *gorm.DB }

func NewGbookHandler(db *gorm.DB) *GbookHandler { return &GbookHandler{db: db} }

// Index 留言列表
func (h *GbookHandler) Index(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Params("page", "1"))
	pageSize := 20

	var total int64
	h.db.Model(&model.Gbook{}).Where("gbook_status = 1").Count(&total)
	var list []model.Gbook
	h.db.Where("gbook_status = 1").Order("gbook_id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	return c.JSON(fiber.Map{"code": 1, "data": fiber.Map{"list": list, "total": total, "page": page}})
}

// Submit 提交留言
func (h *GbookHandler) Submit(c *fiber.Ctx) error {
	content := c.FormValue("content")
	if content == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "留言内容不能为空"})
	}

	gbook := model.Gbook{
		GbookContent: content,
		GbookTime:    time.Now().Unix(),
		GbookStatus:  0, // 待审核
	}
	h.db.Create(&gbook)
	return c.JSON(fiber.Map{"code": 1, "msg": "留言成功，等待审核"})
}

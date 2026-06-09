package frontend

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
	"gocms/internal/template"
)

type GbookHandler struct {
	db       *gorm.DB
	tplEngine *template.Engine
}

func NewGbookHandler(db *gorm.DB, tplEngine *template.Engine) *GbookHandler {
	return &GbookHandler{db: db, tplEngine: tplEngine}
}

// Index 留言列表
func (h *GbookHandler) Index(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Params("page", "1"))
	pageSize := 20

	var total int64
	h.db.Model(&model.Gbook{}).Where("gbook_status = 1").Count(&total)
	var list []model.Gbook
	h.db.Where("gbook_status = 1").Order("gbook_id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return h.tplEngine.FiberRenderer(c, "gbooklist.html", fiber.Map{
		"site_name":   "GOcms",
		"type_info":   model.Type{TypeName: "留言板"},
		"list":        list,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
		"base_url":    "/gbook",
		"is_gbook":    true,
	})
}

// Submit 提交留言（API，保持JSON）
func (h *GbookHandler) Submit(c *fiber.Ctx) error {
	content := c.FormValue("content")
	if content == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "留言内容不能为空"})
	}

	gbook := model.Gbook{
		GbookContent: content,
		GbookTime:    time.Now().Unix(),
		GbookStatus:  0,
	}
	h.db.Create(&gbook)
	return c.JSON(fiber.Map{"code": 1, "msg": "留言成功，等待审核"})
}

package frontend

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
	"gocms/internal/template"
)

type RoleHandler struct {
	db       *gorm.DB
	tplEngine *template.Engine
}

func NewRoleHandler(db *gorm.DB, tplEngine *template.Engine) *RoleHandler {
	return &RoleHandler{db: db, tplEngine: tplEngine}
}

// Index 角色列表
func (h *RoleHandler) Index(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Params("page", "1"))
	pageSize := 24

	var roles []model.Role
	var total int64
	query := h.db.Where("role_status = 1")
	query.Count(&total)
	query.Order("role_id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&roles)

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return h.tplEngine.FiberRenderer(c, "rolelist.html", fiber.Map{
		"site_name":   "GOcms",
		"type_info":   model.Type{TypeName: "角色"},
		"list":        roles,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
		"base_url":    "/role",
		"is_role":     true,
	})
}

// Detail 角色详情
func (h *RoleHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var role model.Role
	if err := h.db.First(&role, id).Error; err != nil {
		return c.Status(404).SendString("角色不存在")
	}

	return h.tplEngine.FiberRenderer(c, "roledetail.html", fiber.Map{
		"site_name": "GOcms",
		"vod":       role,
		"type_name": "角色",
		"is_role":   true,
	})
}

// Show 角色筛选
func (h *RoleHandler) Show(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize := 24

	var roles []model.Role
	var total int64
	query := h.db.Where("role_status = 1")
	query.Count(&total)
	query.Order("role_id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&roles)

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return h.tplEngine.FiberRenderer(c, "rolelist.html", fiber.Map{
		"site_name":   "GOcms",
		"type_info":   model.Type{TypeName: "角色"},
		"list":        roles,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
		"base_url":    "/roleshow",
		"is_role":     true,
	})
}

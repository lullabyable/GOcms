package frontend

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

type RoleHandler struct {
	db *gorm.DB
}

func NewRoleHandler(db *gorm.DB) *RoleHandler {
	return &RoleHandler{db: db}
}

// Index 角色列表页
func (h *RoleHandler) Index(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Params("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize := 24

	var total int64
	h.db.Model(&model.Role{}).Where("role_status = 1").Count(&total)

	var list []model.Role
	h.db.Where("role_status = 1").
		Order("role_id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":      list,
			"page":      page,
			"total":     total,
			"page_size": pageSize,
		},
	})
}

// Detail 角色详情页
func (h *RoleHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	var role model.Role
	if err := h.db.Where("role_id = ?", id).First(&role).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "msg": "角色不存在"})
	}

	// 关联视频
	var vods []model.Vod
	h.db.Where("vod_name LIKE ? AND vod_status = 1", "%"+role.RoleName+"%").
		Order("vod_time DESC").Limit(12).Find(&vods)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"info": role,
			"vods": vods,
		},
	})
}

// Show 角色筛选页
func (h *RoleHandler) Show(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	letter := c.Query("letter")
	order := c.Query("order", "time")
	if page < 1 {
		page = 1
	}
	pageSize := 24

	query := h.db.Model(&model.Role{}).Where("role_status = 1")
	if letter != "" {
		query = query.Where("role_letter = ?", letter)
	}

	var total int64
	query.Count(&total)

	orderClause := "role_id DESC"
	switch order {
	case "hits":
		orderClause = "role_hits DESC"
	case "name":
		orderClause = "role_name ASC"
	}

	var list []model.Role
	query.Order(orderClause).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":      list,
			"page":      page,
			"total":     total,
			"page_size": pageSize,
		},
	})
}

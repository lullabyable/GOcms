package admin

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

// ActorHandler 演员管理
type ActorHandler struct{ db *gorm.DB }

func NewActorHandler(db *gorm.DB) *ActorHandler { return &ActorHandler{db: db} }

func (h *ActorHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	keyword := c.Query("keyword", "")
	sex, _ := strconv.Atoi(c.Query("sex", "-1"))

	query := h.db.Model(&model.Actor{})
	if keyword != "" {
		query = query.Where("actor_name LIKE ?", "%"+keyword+"%")
	}
	if sex >= 0 {
		query = query.Where("actor_sex = ?", sex)
	}

	var total int64
	query.Count(&total)
	var list []model.Actor
	query.Order("actor_id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	return c.JSON(fiber.Map{"code": 1, "data": fiber.Map{"list": list, "total": total, "page": page, "page_size": pageSize}})
}

func (h *ActorHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var a model.Actor
	if err := h.db.First(&a, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 0, "msg": "演员不存在"})
	}
	return c.JSON(fiber.Map{"code": 1, "data": a})
}

func (h *ActorHandler) Save(c *fiber.Ctx) error {
	var a model.Actor
	if err := c.BodyParser(&a); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 0, "msg": "参数错误"})
	}
	if a.ActorID > 0 {
		h.db.Model(&a).Updates(a)
	} else {
		a.ActorTime = time.Now().Format("2006-01-02 15:04:05")
		h.db.Create(&a)
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "保存成功", "data": a})
}

func (h *ActorHandler) Delete(c *fiber.Ctx) error {
	ids := c.Query("ids")
	if ids == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "缺少参数 ids"})
	}
	h.db.Delete(&model.Actor{}, "actor_id IN ?", strings.Split(ids, ","))
	return c.JSON(fiber.Map{"code": 1, "msg": "删除成功"})
}

// RoleHandler 角色管理
type RoleHandler struct{ db *gorm.DB }

func NewRoleHandler(db *gorm.DB) *RoleHandler { return &RoleHandler{db: db} }

func (h *RoleHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	keyword := c.Query("keyword", "")

	query := h.db.Model(&model.Role{})
	if keyword != "" {
		query = query.Where("role_name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)
	var list []model.Role
	query.Order("role_id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	return c.JSON(fiber.Map{"code": 1, "data": fiber.Map{"list": list, "total": total, "page": page, "page_size": pageSize}})
}

func (h *RoleHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var r model.Role
	if err := h.db.First(&r, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 0, "msg": "角色不存在"})
	}
	return c.JSON(fiber.Map{"code": 1, "data": r})
}

func (h *RoleHandler) Save(c *fiber.Ctx) error {
	var r model.Role
	if err := c.BodyParser(&r); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 0, "msg": "参数错误"})
	}
	if r.RoleID > 0 {
		h.db.Model(&r).Updates(r)
	} else {
		h.db.Create(&r)
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "保存成功", "data": r})
}

func (h *RoleHandler) Delete(c *fiber.Ctx) error {
	ids := c.Query("ids")
	if ids == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "缺少参数 ids"})
	}
	h.db.Delete(&model.Role{}, "role_id IN ?", strings.Split(ids, ","))
	return c.JSON(fiber.Map{"code": 1, "msg": "删除成功"})
}

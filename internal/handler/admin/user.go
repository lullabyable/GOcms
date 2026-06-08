package admin

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

type UserHandler struct{ db *gorm.DB }

func NewUserHandler(db *gorm.DB) *UserHandler { return &UserHandler{db: db} }

func (h *UserHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	keyword := c.Query("keyword", "")
	groupID, _ := strconv.Atoi(c.Query("group_id", "0"))

	query := h.db.Model(&model.User{})
	if groupID > 0 {
		query = query.Where("group_id = ?", groupID)
	}
	if keyword != "" {
		query = query.Where("user_name LIKE ? OR user_nick_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)
	var users []model.User
	query.Order("user_id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users)

	return c.JSON(fiber.Map{"code": 1, "data": fiber.Map{"list": users, "total": total, "page": page, "page_size": pageSize}})
}

func (h *UserHandler) Save(c *fiber.Ctx) error {
	var user model.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 0, "msg": "参数错误"})
	}
	if user.UserID > 0 {
		h.db.Model(&user).Omit("user_pwd").Updates(user)
	} else {
		h.db.Create(&user)
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "保存成功"})
}

func (h *UserHandler) Delete(c *fiber.Ctx) error {
	ids := c.Query("ids")
	if ids == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "缺少参数 ids"})
	}
	h.db.Delete(&model.User{}, "user_id IN ?", strings.Split(ids, ","))
	return c.JSON(fiber.Map{"code": 1, "msg": "删除成功"})
}

func (h *UserHandler) ToggleStatus(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var user model.User
	h.db.First(&user, id)
	newStatus := 1
	if user.UserStatus == 1 {
		newStatus = 0
	}
	h.db.Model(&user).Update("user_status", newStatus)
	return c.JSON(fiber.Map{"code": 1, "msg": "操作成功"})
}

// Group 用户组管理
type GroupHandler struct{ db *gorm.DB }

func NewGroupHandler(db *gorm.DB) *GroupHandler { return &GroupHandler{db: db} }

func (h *GroupHandler) List(c *fiber.Ctx) error {
	var groups []model.Group
	h.db.Find(&groups)
	return c.JSON(fiber.Map{"code": 1, "data": groups})
}

func (h *GroupHandler) Save(c *fiber.Ctx) error {
	var group model.Group
	if err := c.BodyParser(&group); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 0, "msg": "参数错误"})
	}
	if group.GroupID > 0 {
		h.db.Model(&group).Updates(group)
	} else {
		h.db.Create(&group)
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "保存成功"})
}

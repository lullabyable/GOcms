package admin

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

// LiveHandler 直播管理处理器
type LiveHandler struct {
	db *gorm.DB
}

func NewLiveHandler(db *gorm.DB) *LiveHandler {
	return &LiveHandler{db: db}
}

// List 直播列表
func (h *LiveHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	keyword := c.Query("keyword", "")

	var lives []model.Live
	var total int64

	query := h.db.Model(&model.Live{})
	if keyword != "" {
		query = query.Where("live_name LIKE ?", "%"+keyword+"%")
	}
	query.Count(&total)
	query.Order("live_sort ASC, live_id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&lives)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":      lives,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// Detail 直播详情
func (h *LiveHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var live model.Live
	if err := h.db.Where("live_id = ?", id).First(&live).Error; err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "直播不存在"})
	}
	return c.JSON(fiber.Map{"code": 1, "data": live})
}

// Save 保存直播
func (h *LiveHandler) Save(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.FormValue("live_id"))

	live := model.Live{
		LiveName:   c.FormValue("live_name"),
		LiveEn:     c.FormValue("live_en"),
		LiveURL:    c.FormValue("live_url"),
		LiveFrom:   c.FormValue("live_from"),
		LivePic:    c.FormValue("live_pic"),
		LiveTime:   c.FormValue("live_time"),
		LiveSort:   0,
		LiveLevel:  0,
		LiveStatus: 1,
	}

	if s, err := strconv.Atoi(c.FormValue("type_id")); err == nil {
		live.TypeID = s
	}
	if s, err := strconv.Atoi(c.FormValue("live_sort")); err == nil {
		live.LiveSort = s
	}
	if s, err := strconv.Atoi(c.FormValue("live_level")); err == nil {
		live.LiveLevel = s
	}
	if s, err := strconv.Atoi(c.FormValue("live_status")); err == nil {
		live.LiveStatus = s
	}

	if id > 0 {
		live.LiveID = id
		if err := h.db.Save(&live).Error; err != nil {
			return c.JSON(fiber.Map{"code": 0, "msg": "更新失败"})
		}
	} else {
		if err := h.db.Create(&live).Error; err != nil {
			return c.JSON(fiber.Map{"code": 0, "msg": "创建失败"})
		}
	}

	return c.JSON(fiber.Map{"code": 1, "msg": "保存成功", "data": live})
}

// Delete 删除直播
func (h *LiveHandler) Delete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.db.Delete(&model.Live{}, id).Error; err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "删除失败"})
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "删除成功"})
}

// ToggleStatus 切换状态
func (h *LiveHandler) ToggleStatus(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	status, _ := strconv.Atoi(c.FormValue("status", "0"))
	if err := h.db.Model(&model.Live{}).Where("live_id = ?", id).Update("live_status", status).Error; err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "操作失败"})
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "操作成功"})
}

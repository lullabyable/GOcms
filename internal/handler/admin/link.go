package admin

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/model"
	"gorm.io/gorm"
)

// LinkHandler 友情链接管理处理器
type LinkHandler struct {
	db *gorm.DB
}

func NewLinkHandler(db *gorm.DB) *LinkHandler {
	return &LinkHandler{db: db}
}

// List 链接列表
func (h *LinkHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	linkType, _ := strconv.Atoi(c.Query("link_type", "-1"))

	query := h.db.Model(&model.Link{})
	if linkType >= 0 {
		query = query.Where("link_type = ?", linkType)
	}

	var total int64
	query.Count(&total)

	var links []model.Link
	query.Order("link_sort ASC, link_id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&links)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":      links,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// Save 创建/更新链接
func (h *LinkHandler) Save(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.FormValue("link_id"))

	link := model.Link{
		LinkName: c.FormValue("link_name"),
		LinkURL:  c.FormValue("link_url"),
		LinkLogo: c.FormValue("link_logo"),
		LinkSort: 0,
	}

	if s, err := strconv.Atoi(c.FormValue("link_type")); err == nil {
		link.LinkType = s
	}
	if s, err := strconv.Atoi(c.FormValue("link_sort")); err == nil {
		link.LinkSort = s
	}

	now := time.Now().Unix()

	if id > 0 {
		link.LinkID = id
		link.LinkTime = now
		if err := h.db.Model(&model.Link{}).Where("link_id = ?", id).Updates(map[string]interface{}{
			"link_name": link.LinkName,
			"link_type": link.LinkType,
			"link_url":  link.LinkURL,
			"link_logo": link.LinkLogo,
			"link_sort": link.LinkSort,
			"link_time": now,
		}).Error; err != nil {
			return c.JSON(fiber.Map{"code": 0, "msg": "更新失败"})
		}
	} else {
		link.LinkAddTime = now
		link.LinkTime = now
		if err := h.db.Create(&link).Error; err != nil {
			return c.JSON(fiber.Map{"code": 0, "msg": "创建失败"})
		}
	}

	return c.JSON(fiber.Map{"code": 1, "msg": "保存成功", "data": link})
}

// Delete 删除链接（支持批量）
func (h *LinkHandler) Delete(c *fiber.Ctx) error {
	ids := c.Query("ids")
	if ids == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "缺少参数 ids"})
	}

	idList := parseIDList(ids)
	if len(idList) == 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "invalid ids"})
	}
	h.db.Delete(&model.Link{}, "link_id IN ?", idList)
	return c.JSON(fiber.Map{"code": 1, "msg": "删除成功"})
}

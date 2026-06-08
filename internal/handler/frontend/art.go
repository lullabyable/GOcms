package frontend

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

type ArtHandler struct{ db *gorm.DB }

func NewArtHandler(db *gorm.DB) *ArtHandler { return &ArtHandler{db: db} }

func (h *ArtHandler) Type(c *fiber.Ctx) error {
	typeID, _ := strconv.Atoi(c.Params("id"))
	page, _ := strconv.Atoi(c.Params("page", "1"))
	pageSize := 20

	var arts []model.Art
	var total int64
	query := h.db.Where("type_id = ? AND art_status = 1", typeID)
	query.Count(&total)
	query.Order("art_time DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&arts)

	var typeInfo model.Type
	h.db.First(&typeInfo, typeID)

	return c.JSON(fiber.Map{"code": 1, "data": fiber.Map{"type_info": typeInfo, "list": arts, "total": total, "page": page}})
}

func (h *ArtHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var art model.Art
	if err := h.db.First(&art, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 0, "msg": "文章不存在"})
	}
	h.db.Model(&art).UpdateColumn("art_hits", gorm.Expr("art_hits + 1"))
	return c.JSON(fiber.Map{"code": 1, "data": art})
}

func (h *ArtHandler) Search(c *fiber.Ctx) error {
	keyword := c.Query("wd", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize := 20

	if keyword == "" {
		return c.JSON(fiber.Map{"code": 1, "data": fiber.Map{"list": nil, "total": 0}})
	}

	var arts []model.Art
	var total int64
	h.db.Where("art_status = 1 AND art_name LIKE ?", "%"+keyword+"%").Count(&total)
	h.db.Where("art_status = 1 AND art_name LIKE ?", "%"+keyword+"%").
		Order("art_hits DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&arts)

	return c.JSON(fiber.Map{"code": 1, "data": fiber.Map{"keyword": keyword, "list": arts, "total": total, "page": page}})
}

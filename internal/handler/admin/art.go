package admin

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

type ArtHandler struct{ db *gorm.DB }

func NewArtHandler(db *gorm.DB) *ArtHandler { return &ArtHandler{db: db} }

func (h *ArtHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	typeID, _ := strconv.Atoi(c.Query("type_id", "0"))
	keyword := c.Query("keyword", "")

	query := h.db.Model(&model.Art{})
	if typeID > 0 {
		query = query.Where("type_id = ?", typeID)
	}
	if keyword != "" {
		query = query.Where("art_name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)
	var arts []model.Art
	query.Order("art_id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&arts)

	return c.JSON(fiber.Map{"code": 1, "data": fiber.Map{"list": arts, "total": total, "page": page, "page_size": pageSize}})
}

func (h *ArtHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var art model.Art
	if err := h.db.First(&art, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 0, "msg": "文章不存在"})
	}
	return c.JSON(fiber.Map{"code": 1, "data": art})
}

func (h *ArtHandler) Save(c *fiber.Ctx) error {
	var art model.Art
	if err := c.BodyParser(&art); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 0, "msg": "参数错误"})
	}
	if art.ID > 0 {
		h.db.Model(&art).Updates(art)
	} else {
		art.ArtTimeAdd = time.Now().Unix()
		art.ArtTime = time.Now().Format("2006-01-02 15:04:05")
		h.db.Create(&art)
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "保存成功", "data": art})
}

func (h *ArtHandler) Delete(c *fiber.Ctx) error {
	ids := c.Query("ids")
	if ids == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "缺少参数 ids"})
	}
	h.db.Delete(&model.Art{}, "art_id IN ?", strings.Split(ids, ","))
	return c.JSON(fiber.Map{"code": 1, "msg": "删除成功"})
}

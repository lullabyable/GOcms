package frontend

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

type MangaHandler struct {
	db *gorm.DB
}

func NewMangaHandler(db *gorm.DB) *MangaHandler {
	return &MangaHandler{db: db}
}

// Type 漫画分类页
func (h *MangaHandler) Type(c *fiber.Ctx) error {
	typeID, _ := strconv.Atoi(c.Params("id"))
	page, _ := strconv.Atoi(c.Params("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize := 20

	var typeInfo model.Type
	if err := h.db.Where("type_id = ?", typeID).First(&typeInfo).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "msg": "分类不存在"})
	}

	var total int64
	h.db.Model(&model.Manga{}).Where("type_id = ? AND manga_status = 1", typeID).Count(&total)

	var list []model.Manga
	h.db.Where("type_id = ? AND manga_status = 1", typeID).
		Order("manga_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"type_info": typeInfo,
			"list":      list,
			"page":      page,
			"total":     total,
			"page_size": pageSize,
		},
	})
}

// Detail 漫画详情页
func (h *MangaHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	var manga model.Manga
	if err := h.db.Where("manga_id = ?", id).First(&manga).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "msg": "漫画不存在"})
	}

	// 增加点击量
	h.db.Model(&manga).UpdateColumn("manga_hits", gorm.Expr("manga_hits + 1"))

	// 相关漫画（同分类）
	var related []model.Manga
	h.db.Where("type_id = ? AND manga_id != ? AND manga_status = 1", manga.TypeID, id).
		Order("manga_hits DESC").Limit(6).Find(&related)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"info":    manga,
			"related": related,
		},
	})
}

// Show 漫画筛选页
func (h *MangaHandler) Show(c *fiber.Ctx) error {
	typeID, _ := strconv.Atoi(c.Params("id"))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	order := c.Query("order", "time")
	area := c.Query("area")
	year := c.Query("year")
	letter := c.Query("letter")
	if page < 1 {
		page = 1
	}
	pageSize := 20

	query := h.db.Model(&model.Manga{}).Where("manga_status = 1")
	if typeID > 0 {
		query = query.Where("type_id = ?", typeID)
	}
	if area != "" {
		query = query.Where("manga_area = ?", area)
	}
	if year != "" {
		query = query.Where("manga_year = ?", year)
	}
	if letter != "" {
		query = query.Where("manga_letter = ?", letter)
	}

	var total int64
	query.Count(&total)

	// 排序
	orderClause := "manga_time DESC"
	switch order {
	case "hits":
		orderClause = "manga_hits DESC"
	case "score":
		orderClause = "manga_score DESC"
	case "time":
		orderClause = "manga_time DESC"
	}

	var list []model.Manga
	query.Order(orderClause).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":      list,
			"page":      page,
			"total":     total,
			"page_size": pageSize,
			"filters": fiber.Map{
				"type_id": typeID,
				"order":   order,
				"area":    area,
				"year":    year,
				"letter":  letter,
			},
		},
	})
}

// Search 漫画搜索
func (h *MangaHandler) Search(c *fiber.Ctx) error {
	wd := c.Query("wd")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize := 20

	if wd == "" {
		return c.JSON(fiber.Map{"code": 1, "data": fiber.Map{"list": nil, "total": 0}})
	}

	query := h.db.Model(&model.Manga{}).Where("manga_status = 1 AND (manga_name LIKE ? OR manga_actor LIKE ?)",
		"%"+wd+"%", "%"+wd+"%")

	var total int64
	query.Count(&total)

	var list []model.Manga
	query.Order("manga_hits DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":      list,
			"page":      page,
			"total":     total,
			"page_size": pageSize,
			"keyword":   wd,
		},
	})
}

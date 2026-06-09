package frontend

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
	"gocms/internal/template"
)

type MangaHandler struct {
	db       *gorm.DB
	tplEngine *template.Engine
}

func NewMangaHandler(db *gorm.DB, tplEngine *template.Engine) *MangaHandler {
	return &MangaHandler{db: db, tplEngine: tplEngine}
}

// Type 漫画分类页
func (h *MangaHandler) Type(c *fiber.Ctx) error {
	typeID, _ := strconv.Atoi(c.Params("id"))
	page, _ := strconv.Atoi(c.Params("page", "1"))
	pageSize := 20

	var typeInfo model.Type
	h.db.Where("type_id = ?", typeID).First(&typeInfo)

	var total int64
	h.db.Model(&model.Manga{}).Where("type_id = ? AND manga_status = 1", typeID).Count(&total)

	var list []model.Manga
	h.db.Where("type_id = ? AND manga_status = 1", typeID).
		Order("manga_time DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return h.tplEngine.FiberRenderer(c, "vodtype.html", fiber.Map{
		"site_name":   "GOcms",
		"type_info":   typeInfo,
		"list":        list,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
		"base_url":    "/mangatype/" + strconv.Itoa(typeID),
		"is_manga":    true,
	})
}

// Detail 漫画详情页
func (h *MangaHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var manga model.Manga
	if err := h.db.Where("manga_id = ?", id).First(&manga).Error; err != nil {
		return c.Status(404).SendString("漫画不存在")
	}

	h.db.Model(&manga).UpdateColumn("manga_hits", gorm.Expr("manga_hits + 1"))

	var related []model.Manga
	h.db.Where("type_id = ? AND manga_id != ? AND manga_status = 1", manga.TypeID, id).
		Order("manga_hits DESC").Limit(6).Find(&related)

	return h.tplEngine.FiberRenderer(c, "voddetail.html", fiber.Map{
		"site_name": "GOcms",
		"vod":       manga,
		"type_name": "漫画",
		"related":   related,
		"is_manga":  true,
	})
}

// Show 漫画筛选页
func (h *MangaHandler) Show(c *fiber.Ctx) error {
	typeID, _ := strconv.Atoi(c.Params("id"))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	area := c.Query("area", "")
	year := c.Query("year", "")
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

	var total int64
	query.Count(&total)
	var list []model.Manga
	query.Order("manga_time DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	var typeInfo model.Type
	if typeID > 0 {
		h.db.First(&typeInfo, typeID)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return h.tplEngine.FiberRenderer(c, "vodtype.html", fiber.Map{
		"site_name":   "GOcms",
		"type_info":   typeInfo,
		"list":        list,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
		"base_url":    "/mangashow/" + strconv.Itoa(typeID),
		"is_manga":    true,
	})
}

// Search 漫画搜索
func (h *MangaHandler) Search(c *fiber.Ctx) error {
	wd := c.Query("wd", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize := 20

	var list []model.Manga
	var total int64
	query := h.db.Model(&model.Manga{}).Where("manga_status = 1 AND (manga_name LIKE ? OR manga_actor LIKE ?)",
		"%"+wd+"%", "%"+wd+"%")
	query.Count(&total)
	query.Order("manga_hits DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return h.tplEngine.FiberRenderer(c, "vodsearch.html", fiber.Map{
		"site_name":   "GOcms",
		"wd":          wd,
		"list":        list,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
		"base_url":    "/mangasearch?wd=" + wd,
		"is_manga":    true,
	})
}

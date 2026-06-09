package frontend

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
	"gocms/internal/template"
)

type ArtHandler struct {
	db       *gorm.DB
	tplEngine *template.Engine
}

func NewArtHandler(db *gorm.DB, tplEngine *template.Engine) *ArtHandler {
	return &ArtHandler{db: db, tplEngine: tplEngine}
}

// Type 文章分类列表
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

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return h.tplEngine.FiberRenderer(c, "arttype.html", fiber.Map{
		"site_name":   "GOcms",
		"type_info":   typeInfo,
		"list":        arts,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
		"base_url":    "/arttype/" + strconv.Itoa(typeID),
		"is_art":      true,
	})
}

// Detail 文章详情
func (h *ArtHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var art model.Art
	if err := h.db.First(&art, id).Error; err != nil {
		return c.Status(404).SendString("文章不存在")
	}

	h.db.Model(&art).UpdateColumn("art_hits", gorm.Expr("art_hits + 1"))

	var typeInfo model.Type
	h.db.First(&typeInfo, art.TypeID)

	var related []model.Art
	h.db.Where("type_id = ? AND art_id != ? AND art_status = 1", art.TypeID, art.ID).
		Order("art_hits DESC").Limit(6).Find(&related)

	return h.tplEngine.FiberRenderer(c, "artdetail.html", fiber.Map{
		"site_name": "GOcms",
		"vod":       art,
		"type_name": typeInfo.TypeName,
		"related":   related,
		"is_art":    true,
	})
}

// Search 文章搜索
func (h *ArtHandler) Search(c *fiber.Ctx) error {
	keyword := c.Query("wd", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize := 20

	var arts []model.Art
	var total int64
	query := h.db.Where("art_status = 1 AND (art_name LIKE ? OR art_author LIKE ?)",
		"%"+keyword+"%", "%"+keyword+"%")
	query.Count(&total)
	query.Order("art_hits DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&arts)

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return h.tplEngine.FiberRenderer(c, "artsearch.html", fiber.Map{
		"site_name":   "GOcms",
		"wd":          keyword,
		"list":        arts,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
		"base_url":    "/artsearch?wd=" + keyword,
		"is_art":      true,
	})
}

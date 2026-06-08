package frontend

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

type VodHandler struct {
	db *gorm.DB
}

func NewVodHandler(db *gorm.DB) *VodHandler {
	return &VodHandler{db: db}
}

// Type 分类列表页
func (h *VodHandler) Type(c *fiber.Ctx) error {
	typeID, _ := strconv.Atoi(c.Params("id"))
	page, _ := strconv.Atoi(c.Params("page", "1"))
	pageSize := 20

	var vods []model.Vod
	var total int64
	query := h.db.Where("type_id = ? AND vod_status = 1", typeID)
	query.Count(&total)
	query.Order("vod_time DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&vods)

	// 分类信息
	var typeInfo model.Type
	h.db.First(&typeInfo, typeID)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"type_info": typeInfo,
			"list":      vods,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// Detail 详情页
func (h *VodHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var vod model.Vod
	if err := h.db.First(&vod, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 0, "msg": "视频不存在"})
	}

	// 增加点击量
	h.db.Model(&vod).UpdateColumn("vod_hits", gorm.Expr("vod_hits + 1"))

	// 相关视频（同分类）
	var related []model.Vod
	h.db.Where("type_id = ? AND vod_id != ? AND vod_status = 1", vod.TypeID, vod.ID).
		Order("vod_hits DESC").Limit(8).Find(&related)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"info":    vod,
			"related": related,
		},
	})
}

// Play 播放页
func (h *VodHandler) Play(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	sid, _ := strconv.Atoi(c.Params("sid", "1"))
	nid, _ := strconv.Atoi(c.Params("nid", "1"))

	var vod model.Vod
	if err := h.db.First(&vod, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 0, "msg": "视频不存在"})
	}

	// 增加点击量
	h.db.Model(&vod).UpdateColumn("vod_hits", gorm.Expr("vod_hits + 1"))

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"info": vod,
			"sid":  sid,
			"nid":  nid,
		},
	})
}

// Search 搜索
func (h *VodHandler) Search(c *fiber.Ctx) error {
	keyword := c.Query("wd", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize := 20

	if keyword == "" {
		return c.JSON(fiber.Map{"code": 1, "data": fiber.Map{"list": nil, "total": 0}})
	}

	var vods []model.Vod
	var total int64
	query := h.db.Where("vod_status = 1 AND (vod_name LIKE ? OR vod_actor LIKE ? OR vod_director LIKE ?)",
		"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	query.Count(&total)
	query.Order("vod_hits DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&vods)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"keyword": keyword,
			"list":    vods,
			"total":   total,
			"page":    page,
		},
	})
}

// Show 筛选列表
func (h *VodHandler) Show(c *fiber.Ctx) error {
	typeID, _ := strconv.Atoi(c.Params("id"))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	area := c.Query("area", "")
	year := c.Query("year", "")
	lang := c.Query("lang", "")
	order := c.Query("order", "vod_time")
	pageSize := 20

	query := h.db.Where("vod_status = 1")
	if typeID > 0 {
		query = query.Where("type_id = ?", typeID)
	}
	if area != "" {
		query = query.Where("vod_area = ?", area)
	}
	if year != "" {
		query = query.Where("vod_year = ?", year)
	}
	if lang != "" {
		query = query.Where("vod_lang = ?", lang)
	}

	var total int64
	query.Count(&total)
	var vods []model.Vod
	query.Order(order + " DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&vods)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":  vods,
			"total": total,
			"page":  page,
			"filter": fiber.Map{"area": area, "year": year, "lang": lang},
		},
	})
}

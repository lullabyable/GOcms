package frontend

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

type IndexHandler struct {
	db *gorm.DB
}

func NewIndexHandler(db *gorm.DB) *IndexHandler {
	return &IndexHandler{db: db}
}

// Index 首页
func (h *IndexHandler) Index(c *fiber.Ctx) error {
	// 最新视频
	var latestVods []model.Vod
	h.db.Where("vod_status = 1").Order("vod_time DESC").Limit(12).Find(&latestVods)

	// 热门视频
	var hotVods []model.Vod
	h.db.Where("vod_status = 1").Order("vod_hits DESC").Limit(12).Find(&hotVods)

	// 推荐视频（高评分）
	var recVods []model.Vod
	h.db.Where("vod_status = 1 AND vod_level > 0").Order("vod_level DESC, vod_score DESC").Limit(12).Find(&recVods)

	// 最新文章
	var latestArts []model.Art
	h.db.Where("art_status = 1").Order("art_time DESC").Limit(8).Find(&latestArts)

	// 分类列表
	var types []model.Type
	h.db.Where("type_status = 1").Order("type_sort ASC").Find(&types)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"latest_vods": latestVods,
			"hot_vods":    hotVods,
			"rec_vods":    recVods,
			"latest_arts": latestArts,
			"types":       types,
		},
	})
}

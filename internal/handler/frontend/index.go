package frontend

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
	"gocms/internal/template"
)

type IndexHandler struct {
	db       *gorm.DB
	tplEngine *template.Engine
}

func NewIndexHandler(db *gorm.DB, tplEngine *template.Engine) *IndexHandler {
	return &IndexHandler{db: db, tplEngine: tplEngine}
}

// Index 首页
func (h *IndexHandler) Index(c *fiber.Ctx) error {
	// 最新视频
	var latestVods []model.Vod
	h.db.Where("vod_status = 1").Order("vod_time DESC").Limit(12).Find(&latestVods)

	// 热门视频
	var hotVods []model.Vod
	h.db.Where("vod_status = 1").Order("vod_hits DESC").Limit(12).Find(&hotVods)

	// 推荐视频（高评分/高权重）
	var recVods []model.Vod
	h.db.Where("vod_status = 1 AND vod_level > 0").Order("vod_level DESC, vod_score DESC").Limit(6).Find(&recVods)

	// 最新文章
	var latestArts []model.Art
	h.db.Where("art_status = 1").Order("art_time DESC").Limit(4).Find(&latestArts)

	// 分类列表（顶级）
	var types []model.Type
	h.db.Where("type_pid = 0 AND type_status = 1").Order("type_sort ASC").Find(&types)

	data := fiber.Map{
		"site_name":    "GOcms",
		"latest_vods":  latestVods,
		"hot_vods":     hotVods,
		"rec_vods":     recVods,
		"latest_arts":  latestArts,
		"types":        types,
	}

	return h.tplEngine.FiberRenderer(c, "index.html", data)
}

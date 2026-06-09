package frontend

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
	"gocms/internal/template"
)

type VodHandler struct {
	db       *gorm.DB
	tplEngine *template.Engine
}

func NewVodHandler(db *gorm.DB, tplEngine *template.Engine) *VodHandler {
	return &VodHandler{db: db, tplEngine: tplEngine}
}

type PlaySource struct {
	Name     string
	Episodes []PlayEpisode
}

type PlayEpisode struct {
	Name string
	URL  string
}

func parsePlaySources(fromStr, urlStr string) []PlaySource {
	if fromStr == "" || urlStr == "" {
		return nil
	}
	froms := strings.Split(fromStr, "$$$")
	urls := strings.Split(urlStr, "$$$")
	count := len(froms)
	if len(urls) < count {
		count = len(urls)
	}
	var result []PlaySource
	for i := 0; i < count; i++ {
		source := PlaySource{Name: strings.TrimSpace(froms[i])}
		episodes := strings.Split(urls[i], "#")
		for _, ep := range episodes {
			ep = strings.TrimSpace(ep)
			if ep == "" {
				continue
			}
			parts := strings.SplitN(ep, "$", 2)
			if len(parts) == 2 {
				source.Episodes = append(source.Episodes, PlayEpisode{Name: parts[0], URL: parts[1]})
			}
		}
		if len(source.Episodes) > 0 {
			result = append(result, source)
		}
	}
	return result
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

	var typeInfo model.Type
	h.db.First(&typeInfo, typeID)

	// 侧栏热门
	var hotVods []model.Vod
	h.db.Where("type_id = ? AND vod_status = 1", typeID).Order("vod_hits DESC").Limit(10).Find(&hotVods)

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return h.tplEngine.FiberRenderer(c, "vodtype.html", fiber.Map{
		"site_name":   "GOcms",
		"type_info":   typeInfo,
		"list":        vods,
		"hot_vods":    hotVods,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
		"base_url":    "/vodtype/" + strconv.Itoa(typeID),
	})
}

// Detail 详情页
func (h *VodHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var vod model.Vod
	if err := h.db.First(&vod, id).Error; err != nil {
		return c.Status(404).SendString("视频不存在")
	}

	h.db.Model(&vod).UpdateColumn("vod_hits", gorm.Expr("vod_hits + 1"))

	var related []model.Vod
	h.db.Where("type_id = ? AND vod_id != ? AND vod_status = 1", vod.TypeID, vod.ID).
		Order("vod_hits DESC").Limit(8).Find(&related)

	var typeInfo model.Type
	h.db.First(&typeInfo, vod.TypeID)

	playSources := parsePlaySources(vod.VodPlayFrom, vod.VodPlayURL)

	return h.tplEngine.FiberRenderer(c, "voddetail.html", fiber.Map{
		"site_name":    "GOcms",
		"vod":          vod,
		"type_name":    typeInfo.TypeName,
		"related":      related,
		"play_sources": playSources,
	})
}

// Play 播放页
func (h *VodHandler) Play(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	sid, _ := strconv.Atoi(c.Params("sid", "1"))
	nid, _ := strconv.Atoi(c.Params("nid", "1"))

	var vod model.Vod
	if err := h.db.First(&vod, id).Error; err != nil {
		return c.Status(404).SendString("视频不存在")
	}

	h.db.Model(&vod).UpdateColumn("vod_hits", gorm.Expr("vod_hits + 1"))

	playSources := parsePlaySources(vod.VodPlayFrom, vod.VodPlayURL)

	// 找到当前播放地址
	playURL := ""
	playName := ""
	if sid > 0 && sid <= len(playSources) {
		source := playSources[sid-1]
		if nid > 0 && nid <= len(source.Episodes) {
			playURL = source.Episodes[nid-1].URL
			playName = source.Episodes[nid-1].Name
		}
	}

	// 随机推荐
	var recommend []model.Vod
	h.db.Where("vod_status = 1").Order("RAND()").Limit(6).Find(&recommend)

	return h.tplEngine.FiberRenderer(c, "vodplay.html", fiber.Map{
		"site_name":    "GOcms",
		"vod":          vod,
		"sid":          sid,
		"nid":          nid,
		"play_url":     playURL,
		"play_name":    playName,
		"play_sources": playSources,
		"recommend":    recommend,
	})
}

// Search 搜索
func (h *VodHandler) Search(c *fiber.Ctx) error {
	keyword := c.Query("wd", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize := 20

	if keyword == "" {
		return h.tplEngine.FiberRenderer(c, "vodsearch.html", fiber.Map{
			"site_name": "GOcms", "wd": "", "list": nil, "total": 0, "page": 1, "total_pages": 0,
		})
	}

	var vods []model.Vod
	var total int64
	query := h.db.Where("vod_status = 1 AND (vod_name LIKE ? OR vod_actor LIKE ? OR vod_director LIKE ?)",
		"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	query.Count(&total)
	query.Order("vod_hits DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&vods)

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	// 热门搜索
	var hotVods []model.Vod
	h.db.Where("vod_status = 1").Order("vod_hits DESC").Limit(10).Find(&hotVods)

	return h.tplEngine.FiberRenderer(c, "vodsearch.html", fiber.Map{
		"site_name":   "GOcms",
		"wd":          keyword,
		"list":        vods,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
		"base_url":    "/vodsearch?wd=" + keyword,
		"hot_vods":    hotVods,
	})
}

// Show 筛选列表（复用 vodtype 模板）
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
		"list":        vods,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
		"base_url":    "/vodshow/" + strconv.Itoa(typeID),
	})
}

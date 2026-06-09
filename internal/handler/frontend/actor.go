package frontend

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
	"gocms/internal/template"
)

type ActorHandler struct {
	db       *gorm.DB
	tplEngine *template.Engine
}

func NewActorHandler(db *gorm.DB, tplEngine *template.Engine) *ActorHandler {
	return &ActorHandler{db: db, tplEngine: tplEngine}
}

// Index 演员列表
func (h *ActorHandler) Index(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Params("page", "1"))
	pageSize := 24

	var actors []model.Actor
	var total int64
	query := h.db.Where("actor_status = 1")
	query.Count(&total)
	query.Order("actor_id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&actors)

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return h.tplEngine.FiberRenderer(c, "actorlist.html", fiber.Map{
		"site_name":   "GOcms",
		"type_info":   model.Type{TypeName: "演员"},
		"list":        actors,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
		"base_url":    "/actor",
		"is_actor":    true,
	})
}

// Detail 演员详情
func (h *ActorHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var actor model.Actor
	if err := h.db.First(&actor, id).Error; err != nil {
		return c.Status(404).SendString("演员不存在")
	}

	// 该演员的视频
	var vods []model.Vod
	h.db.Where("vod_status = 1 AND (vod_actor LIKE ?)", "%"+actor.ActorName+"%").
		Order("vod_hits DESC").Limit(12).Find(&vods)

	return h.tplEngine.FiberRenderer(c, "actordetail.html", fiber.Map{
		"site_name": "GOcms",
		"vod":       actor,
		"type_name": "演员",
		"related":   vods,
		"is_actor":  true,
	})
}

// Show 演员筛选
func (h *ActorHandler) Show(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	area := c.Query("area", "")
	sex := c.Query("sex", "")
	pageSize := 24

	query := h.db.Where("actor_status = 1")
	if area != "" {
		query = query.Where("actor_area = ?", area)
	}
	if sex != "" {
		query = query.Where("actor_sex = ?", sex)
	}

	var total int64
	query.Count(&total)
	var actors []model.Actor
	query.Order("actor_id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&actors)

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return h.tplEngine.FiberRenderer(c, "actorlist.html", fiber.Map{
		"site_name":   "GOcms",
		"type_info":   model.Type{TypeName: "演员"},
		"list":        actors,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
		"base_url":    "/actorshow",
		"is_actor":    true,
	})
}

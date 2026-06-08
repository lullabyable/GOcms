package frontend

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

type ActorHandler struct {
	db *gorm.DB
}

func NewActorHandler(db *gorm.DB) *ActorHandler {
	return &ActorHandler{db: db}
}

// Index 演员列表页
func (h *ActorHandler) Index(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Params("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize := 24

	var total int64
	h.db.Model(&model.Actor{}).Where("actor_status = 1").Count(&total)

	var list []model.Actor
	h.db.Where("actor_status = 1").
		Order("actor_id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":      list,
			"page":      page,
			"total":     total,
			"page_size": pageSize,
		},
	})
}

// Detail 演员详情页
func (h *ActorHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	var actor model.Actor
	if err := h.db.Where("actor_id = ?", id).First(&actor).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "msg": "演员不存在"})
	}

	// 该演员参演的视频
	var vods []model.Vod
	h.db.Where("vod_actor LIKE ? AND vod_status = 1", "%"+actor.ActorName+"%").
		Order("vod_time DESC").Limit(12).Find(&vods)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"info": actor,
			"vods": vods,
		},
	})
}

// Show 演员筛选页
func (h *ActorHandler) Show(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	area := c.Query("area")
	sex := c.Query("sex")
	letter := c.Query("letter")
	order := c.Query("order", "time")
	if page < 1 {
		page = 1
	}
	pageSize := 24

	query := h.db.Model(&model.Actor{}).Where("actor_status = 1")
	if area != "" {
		query = query.Where("actor_area = ?", area)
	}
	if sex != "" {
		query = query.Where("actor_sex = ?", sex)
	}
	if letter != "" {
		query = query.Where("actor_letter = ?", letter)
	}

	var total int64
	query.Count(&total)

	orderClause := "actor_id DESC"
	switch order {
	case "hits":
		orderClause = "actor_hits DESC"
	case "name":
		orderClause = "actor_name ASC"
	}

	var list []model.Actor
	query.Order(orderClause).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":      list,
			"page":      page,
			"total":     total,
			"page_size": pageSize,
			"filters": fiber.Map{
				"area":   area,
				"sex":    sex,
				"letter": letter,
				"order":  order,
			},
		},
	})
}

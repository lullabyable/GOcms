package frontend

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

type TopicHandler struct {
	db *gorm.DB
}

func NewTopicHandler(db *gorm.DB) *TopicHandler {
	return &TopicHandler{db: db}
}

// Index 专题列表页
func (h *TopicHandler) Index(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Params("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize := 12

	var total int64
	h.db.Model(&model.Topic{}).Where("topic_status = 1").Count(&total)

	var list []model.Topic
	h.db.Where("topic_status = 1").
		Order("topic_sort ASC, topic_id DESC").
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

// Detail 专题详情页
func (h *TopicHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	var topic model.Topic
	if err := h.db.Where("topic_id = ?", id).First(&topic).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "msg": "专题不存在"})
	}

	// 专题关联的视频（通过 vod_topic 或直接查询）
	var vods []model.Vod
	if topic.TopicVodID != "" {
		// 解析关联的视频ID列表
		h.db.Where("FIND_IN_SET(vod_id, ?) AND vod_status = 1", topic.TopicVodID).
			Order("vod_time DESC").Find(&vods)
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"info": topic,
			"vods": vods,
		},
	})
}

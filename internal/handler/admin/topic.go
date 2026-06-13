package admin

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/model"
	"gorm.io/gorm"
)

// TopicHandler 专题管理处理器
type TopicHandler struct {
	db *gorm.DB
}

func NewTopicHandler(db *gorm.DB) *TopicHandler {
	return &TopicHandler{db: db}
}

// List 专题列表
func (h *TopicHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	status, _ := strconv.Atoi(c.Query("status", "-1"))
	keyword := c.Query("keyword", "")

	query := h.db.Model(&model.Topic{})
	if status >= 0 {
		query = query.Where("topic_status = ?", status)
	}
	if keyword != "" {
		query = query.Where("topic_name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var topics []model.Topic
	query.Order("topic_sort ASC, topic_id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&topics)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":      topics,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// Detail 专题详情
func (h *TopicHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var topic model.Topic
	if err := h.db.First(&topic, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 0, "msg": "专题不存在"})
	}
	return c.JSON(fiber.Map{"code": 1, "data": topic})
}

// Save 创建/更新专题
func (h *TopicHandler) Save(c *fiber.Ctx) error {
	var topic model.Topic
	if err := c.BodyParser(&topic); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 0, "msg": "参数错误"})
	}

	if topic.TopicID > 0 {
		topic.TopicTimeMake = time.Now()
		if err := h.db.Model(&topic).Updates(topic).Error; err != nil {
			return c.JSON(fiber.Map{"code": 0, "msg": "更新失败"})
		}
	} else {
		topic.TopicTime = time.Now()
		topic.TopicTimeMake = time.Now()
		if err := h.db.Create(&topic).Error; err != nil {
			return c.JSON(fiber.Map{"code": 0, "msg": "创建失败"})
		}
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "保存成功", "data": topic})
}

// Delete 删除专题（支持批量）
func (h *TopicHandler) Delete(c *fiber.Ctx) error {
	ids := c.Query("ids")
	if ids == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "缺少参数 ids"})
	}

	idList := parseIDList(ids)
	if len(idList) == 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "invalid ids"})
	}
	h.db.Delete(&model.Topic{}, "topic_id IN ?", idList)
	return c.JSON(fiber.Map{"code": 1, "msg": "删除成功"})
}

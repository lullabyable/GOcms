package frontend

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
	"gocms/internal/template"
)

type TopicHandler struct {
	db       *gorm.DB
	tplEngine *template.Engine
}

func NewTopicHandler(db *gorm.DB, tplEngine *template.Engine) *TopicHandler {
	return &TopicHandler{db: db, tplEngine: tplEngine}
}

// Index 专题列表
func (h *TopicHandler) Index(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Params("page", "1"))
	pageSize := 12

	var topics []model.Topic
	var total int64
	query := h.db.Where("topic_status = 1")
	query.Count(&total)
	query.Order("topic_sort ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&topics)

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return h.tplEngine.FiberRenderer(c, "topiclist.html", fiber.Map{
		"site_name":   "GOcms",
		"type_info":   model.Type{TypeName: "专题"},
		"list":        topics,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
		"base_url":    "/topic",
		"is_topic":    true,
	})
}

// Detail 专题详情
func (h *TopicHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var topic model.Topic
	if err := h.db.First(&topic, id).Error; err != nil {
		return c.Status(404).SendString("专题不存在")
	}

	// 解析关联视频ID
	var vods []model.Vod
	if topic.TopicVodID != "" {
		ids := strings.Split(topic.TopicVodID, ",")
		var intIDs []int
		for _, idStr := range ids {
			if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
				intIDs = append(intIDs, id)
			}
		}
		if len(intIDs) > 0 {
			h.db.Where("vod_id IN ? AND vod_status = 1", intIDs).Find(&vods)
		}
	}

	return h.tplEngine.FiberRenderer(c, "topicdetail.html", fiber.Map{
		"site_name": "GOcms",
		"vod":       topic,
		"type_name": "专题",
		"related":   vods,
		"is_topic":  true,
	})
}

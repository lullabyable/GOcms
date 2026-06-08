package admin

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

type CommentHandler struct{ db *gorm.DB }

func NewCommentHandler(db *gorm.DB) *CommentHandler { return &CommentHandler{db: db} }

func (h *CommentHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	status, _ := strconv.Atoi(c.Query("status", "-1"))

	query := h.db.Model(&model.Comment{})
	if status >= 0 {
		query = query.Where("comment_status = ?", status)
	}

	var total int64
	query.Count(&total)
	var comments []model.Comment
	query.Order("comment_id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&comments)

	return c.JSON(fiber.Map{"code": 1, "data": fiber.Map{"list": comments, "total": total}})
}

func (h *CommentHandler) Audit(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	status, _ := strconv.Atoi(c.FormValue("status"))
	h.db.Model(&model.Comment{}).Where("comment_id = ?", id).Update("comment_status", status)
	return c.JSON(fiber.Map{"code": 1, "msg": "操作成功"})
}

func (h *CommentHandler) Delete(c *fiber.Ctx) error {
	ids := c.Query("ids")
	h.db.Delete(&model.Comment{}, "comment_id IN ?", strings.Split(ids, ","))
	return c.JSON(fiber.Map{"code": 1, "msg": "删除成功"})
}

// GbookHandler 留言管理
type GbookHandler struct{ db *gorm.DB }

func NewGbookHandler(db *gorm.DB) *GbookHandler { return &GbookHandler{db: db} }

func (h *GbookHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	var total int64
	h.db.Model(&model.Gbook{}).Count(&total)
	var gbooks []model.Gbook
	h.db.Order("gbook_id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&gbooks)

	return c.JSON(fiber.Map{"code": 1, "data": fiber.Map{"list": gbooks, "total": total}})
}

func (h *GbookHandler) Reply(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	reply := c.FormValue("reply")
	h.db.Model(&model.Gbook{}).Where("gbook_id = ?", id).Updates(map[string]interface{}{
		"gbook_reply":      reply,
		"gbook_status":     1,
		"gbook_reply_time": 0, // TODO: time.Now().Unix()
	})
	return c.JSON(fiber.Map{"code": 1, "msg": "回复成功"})
}

func (h *GbookHandler) Delete(c *fiber.Ctx) error {
	ids := c.Query("ids")
	h.db.Delete(&model.Gbook{}, "gbook_id IN ?", strings.Split(ids, ","))
	return c.JSON(fiber.Map{"code": 1, "msg": "删除成功"})
}

package admin

import (
	"html"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/model"
	"gorm.io/gorm"
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
	return c.JSON(fiber.Map{"code": 1, "msg": "operation successful"})
}

func (h *CommentHandler) Delete(c *fiber.Ctx) error {
	idList := parseIDList(c.Query("ids"))
	if len(idList) == 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "invalid ids"})
	}
	h.db.Delete(&model.Comment{}, "comment_id IN ?", idList)
	return c.JSON(fiber.Map{"code": 1, "msg": "delete successful"})
}

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
	reply := html.EscapeString(c.FormValue("reply"))
	h.db.Model(&model.Gbook{}).Where("gbook_id = ?", id).Updates(map[string]interface{}{
		"gbook_reply":      reply,
		"gbook_status":     1,
		"gbook_reply_time": 0, // TODO: time.Now().Unix()
	})
	return c.JSON(fiber.Map{"code": 1, "msg": "reply successful"})
}

func (h *GbookHandler) Delete(c *fiber.Ctx) error {
	idList := parseIDList(c.Query("ids"))
	if len(idList) == 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "invalid ids"})
	}
	h.db.Delete(&model.Gbook{}, "gbook_id IN ?", idList)
	return c.JSON(fiber.Map{"code": 1, "msg": "delete successful"})
}

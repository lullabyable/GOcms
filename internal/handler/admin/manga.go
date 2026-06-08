package admin

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

type MangaHandler struct{ db *gorm.DB }

func NewMangaHandler(db *gorm.DB) *MangaHandler { return &MangaHandler{db: db} }

func (h *MangaHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	typeID, _ := strconv.Atoi(c.Query("type_id", "0"))
	keyword := c.Query("keyword", "")

	query := h.db.Model(&model.Manga{})
	if typeID > 0 {
		query = query.Where("type_id = ?", typeID)
	}
	if keyword != "" {
		query = query.Where("manga_name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)
	var list []model.Manga
	query.Order("manga_id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	return c.JSON(fiber.Map{"code": 1, "data": fiber.Map{"list": list, "total": total, "page": page, "page_size": pageSize}})
}

func (h *MangaHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var m model.Manga
	if err := h.db.First(&m, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 0, "msg": "漫画不存在"})
	}
	return c.JSON(fiber.Map{"code": 1, "data": m})
}

func (h *MangaHandler) Save(c *fiber.Ctx) error {
	var m model.Manga
	if err := c.BodyParser(&m); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 0, "msg": "参数错误"})
	}
	if m.ID > 0 {
		h.db.Model(&m).Updates(m)
	} else {
		m.MangaTimeAdd = time.Now().Unix()
		m.MangaTime = time.Now().Format("2006-01-02 15:04:05")
		h.db.Create(&m)
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "保存成功", "data": m})
}

func (h *MangaHandler) Delete(c *fiber.Ctx) error {
	ids := c.Query("ids")
	if ids == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "缺少参数 ids"})
	}
	h.db.Delete(&model.Manga{}, "manga_id IN ?", strings.Split(ids, ","))
	return c.JSON(fiber.Map{"code": 1, "msg": "删除成功"})
}

func (h *MangaHandler) Audit(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	status, _ := strconv.Atoi(c.FormValue("status"))
	h.db.Model(&model.Manga{}).Where("manga_id = ?", id).Update("manga_status", status)
	return c.JSON(fiber.Map{"code": 1, "msg": "操作成功"})
}

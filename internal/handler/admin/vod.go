package admin

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/model"
	"gorm.io/gorm"
)

type VodHandler struct {
	db *gorm.DB
}

func NewVodHandler(db *gorm.DB) *VodHandler {
	return &VodHandler{db: db}
}

// List 视频列表
func (h *VodHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	typeID, _ := strconv.Atoi(c.Query("type_id", "0"))
	status, _ := strconv.Atoi(c.Query("status", "-1"))
	keyword := c.Query("keyword", "")

	query := h.db.Model(&model.Vod{})
	if typeID > 0 {
		query = query.Where("type_id = ?", typeID)
	}
	if status >= 0 {
		query = query.Where("vod_status = ?", status)
	}
	if keyword != "" {
		query = query.Where("vod_name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var vods []model.Vod
	query.Order("vod_id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&vods)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":      vods,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// Detail 视频详情
func (h *VodHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var vod model.Vod
	if err := h.db.First(&vod, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 0, "msg": "视频不存在"})
	}
	return c.JSON(fiber.Map{"code": 1, "data": vod})
}

// Save 创建/更新视频
func (h *VodHandler) Save(c *fiber.Ctx) error {
	var vod model.Vod
	if err := c.BodyParser(&vod); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 0, "msg": "参数错误"})
	}

	if vod.ID > 0 {
		vod.VodTimeHits = time.Now().Unix()
		if err := h.db.Model(&vod).Updates(vod).Error; err != nil {
			return c.JSON(fiber.Map{"code": 0, "msg": "更新失败"})
		}
	} else {
		vod.VodTimeAdd = time.Now().Unix()
		vod.VodTime = time.Now().Format("2006-01-02 15:04:05")
		if err := h.db.Create(&vod).Error; err != nil {
			return c.JSON(fiber.Map{"code": 0, "msg": "创建失败"})
		}
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "保存成功", "data": vod})
}

// Delete 删除视频（支持批量）
func (h *VodHandler) Delete(c *fiber.Ctx) error {
	ids := c.Query("ids")
	if ids == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "缺少参数 ids"})
	}

	idList := parseIDList(ids)
	if len(idList) == 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "invalid ids"})
	}
	h.db.Delete(&model.Vod{}, "vod_id IN ?", idList)
	return c.JSON(fiber.Map{"code": 1, "msg": "删除成功"})
}

// Audit 审核/状态变更
func (h *VodHandler) Audit(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	status, _ := strconv.Atoi(c.FormValue("status"))

	h.db.Model(&model.Vod{}).Where("vod_id = ?", id).Update("vod_status", status)
	return c.JSON(fiber.Map{"code": 1, "msg": "操作成功"})
}

// Batch 批量操作
func (h *VodHandler) Batch(c *fiber.Ctx) error {
	action := c.FormValue("action")
	ids := c.FormValue("ids")
	if ids == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "缺少参数 ids"})
	}
	idList := parseIDList(ids)
	if len(idList) == 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "invalid ids"})
	}

	switch action {
	case "delete":
		h.db.Delete(&model.Vod{}, "vod_id IN ?", idList)
	case "audit":
		status, _ := strconv.Atoi(c.FormValue("status"))
		h.db.Model(&model.Vod{}).Where("vod_id IN ?", idList).Update("vod_status", status)
	case "type":
		typeID, _ := strconv.Atoi(c.FormValue("type_id"))
		h.db.Model(&model.Vod{}).Where("vod_id IN ?", idList).Update("type_id", typeID)
	default:
		return c.JSON(fiber.Map{"code": 0, "msg": "未知操作"})
	}

	return c.JSON(fiber.Map{"code": 1, "msg": "批量操作成功"})
}

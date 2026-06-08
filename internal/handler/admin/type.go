package admin

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

type TypeHandler struct {
	db *gorm.DB
}

func NewTypeHandler(db *gorm.DB) *TypeHandler {
	return &TypeHandler{db: db}
}

// List 分类列表（支持树形）
func (h *TypeHandler) List(c *fiber.Ctx) error {
	var types []model.Type
	pid, _ := strconv.Atoi(c.Query("pid", "-1"))

	query := h.db.Order("type_sort ASC, type_id ASC")
	if pid >= 0 {
		query = query.Where("type_pid = ?", pid)
	}
	query.Find(&types)

	return c.JSON(fiber.Map{"code": 1, "data": types})
}

// Tree 分类树
func (h *TypeHandler) Tree(c *fiber.Ctx) error {
	var types []model.Type
	h.db.Order("type_sort ASC, type_id ASC").Find(&types)

	tree := buildTree(types, 0)
	return c.JSON(fiber.Map{"code": 1, "data": tree})
}

type TypeTree struct {
	model.Type
	Children []TypeTree `json:"children"`
}

func buildTree(types []model.Type, pid int) []TypeTree {
	var result []TypeTree
	for _, t := range types {
		if t.TypePID == pid {
			node := TypeTree{Type: t}
			node.Children = buildTree(types, t.TypeID)
			result = append(result, node)
		}
	}
	return result
}

// Detail 分类详情
func (h *TypeHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var t model.Type
	if err := h.db.First(&t, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 0, "msg": "分类不存在"})
	}
	return c.JSON(fiber.Map{"code": 1, "data": t})
}

// Save 创建/更新分类
func (h *TypeHandler) Save(c *fiber.Ctx) error {
	var t model.Type
	if err := c.BodyParser(&t); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 0, "msg": "参数错误"})
	}

	if t.TypeID > 0 {
		// 更新
		if err := h.db.Model(&t).Updates(t).Error; err != nil {
			return c.JSON(fiber.Map{"code": 0, "msg": "更新失败"})
		}
	} else {
		// 创建
		if err := h.db.Create(&t).Error; err != nil {
			return c.JSON(fiber.Map{"code": 0, "msg": "创建失败"})
		}
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "保存成功", "data": t})
}

// Delete 删除分类
func (h *TypeHandler) Delete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	// 检查是否有子分类
	var childCount int64
	h.db.Model(&model.Type{}).Where("type_pid = ?", id).Count(&childCount)
	if childCount > 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "存在子分类，无法删除"})
	}

	// 检查是否有关联内容
	var vodCount int64
	h.db.Model(&model.Vod{}).Where("type_id = ?", id).Count(&vodCount)
	if vodCount > 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "分类下存在视频，无法删除"})
	}

	h.db.Delete(&model.Type{}, id)
	return c.JSON(fiber.Map{"code": 1, "msg": "删除成功"})
}

// Sort 排序
func (h *TypeHandler) Sort(c *fiber.Ctx) error {
	var items []struct {
		ID   int `json:"id"`
		Sort int `json:"sort"`
	}
	if err := c.BodyParser(&items); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 0, "msg": "参数错误"})
	}

	for _, item := range items {
		h.db.Model(&model.Type{}).Where("type_id = ?", item.ID).Update("type_sort", item.Sort)
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "排序成功"})
}

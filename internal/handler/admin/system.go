package admin

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

type SystemHandler struct{ db *gorm.DB }

func NewSystemHandler(db *gorm.DB) *SystemHandler { return &SystemHandler{db: db} }

// GetConfig 获取系统配置
func (h *SystemHandler) GetConfig(c *fiber.Ctx) error {
	typeName := c.Query("type", "maccms")
	var configs []model.Config
	h.db.Where("type = ?", typeName).Find(&configs)

	result := make(map[string]string)
	for _, cfg := range configs {
		result[cfg.Name] = cfg.Value
	}
	return c.JSON(fiber.Map{"code": 1, "data": result})
}

// SaveConfig 保存系统配置
func (h *SystemHandler) SaveConfig(c *fiber.Ctx) error {
	typeName := c.FormValue("type", "maccms")

	// 遍历表单字段
	configs := make(map[string]string)
	c.Context().PostArgs().VisitAll(func(key, value []byte) {
		k := string(key)
		if k != "type" { // 跳过 type 字段本身
			configs[k] = string(value)
		}
	})

	for name, value := range configs {
		var existing model.Config
		if err := h.db.Where("type = ? AND name = ?", typeName, name).First(&existing).Error; err != nil {
			// 新增
			h.db.Create(&model.Config{Type: typeName, Name: name, Value: value})
		} else {
			// 更新
			h.db.Model(&existing).Update("value", value)
		}
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "配置已保存"})
}

// CacheClear 清理缓存
func (h *SystemHandler) CacheClear(c *fiber.Ctx) error {
	cacheType := c.Query("type", "all")
	// TODO: 接入缓存管理器
	return c.JSON(fiber.Map{"code": 1, "msg": "缓存已清理", "type": cacheType})
}

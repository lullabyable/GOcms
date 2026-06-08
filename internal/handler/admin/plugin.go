package admin

import (
	"github.com/gofiber/fiber/v2"
	"gocms/internal/service/plugin"
)

// PluginHandler 插件管理处理器
type PluginHandler struct {
	manager *plugin.Manager
}

func NewPluginHandler(mgr *plugin.Manager) *PluginHandler {
	return &PluginHandler{manager: mgr}
}

// List 插件列表
func (h *PluginHandler) List(c *fiber.Ctx) error {
	plugins, err := h.manager.GetList()
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "获取列表失败"})
	}
	return c.JSON(fiber.Map{"code": 1, "data": plugins})
}

// Install 安装插件
func (h *PluginHandler) Install(c *fiber.Ctx) error {
	name := c.FormValue("name")
	if name == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "插件名称不能为空"})
	}
	if err := h.manager.Install(name); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "安装成功"})
}

// Uninstall 卸载插件
func (h *PluginHandler) Uninstall(c *fiber.Ctx) error {
	name := c.Params("name")
	if err := h.manager.Uninstall(name); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "卸载成功"})
}

// Enable 启用插件
func (h *PluginHandler) Enable(c *fiber.Ctx) error {
	name := c.Params("name")
	if err := h.manager.Enable(name); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "已启用"})
}

// Disable 禁用插件
func (h *PluginHandler) Disable(c *fiber.Ctx) error {
	name := c.Params("name")
	if err := h.manager.Disable(name); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "已禁用"})
}

// Config 获取插件配置
func (h *PluginHandler) Config(c *fiber.Ctx) error {
	name := c.Params("name")
	cfg, err := h.manager.GetConfig(name)
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 1, "data": cfg})
}

// SaveConfig 保存插件配置
func (h *PluginHandler) SaveConfig(c *fiber.Ctx) error {
	name := c.Params("name")
	// 从表单获取配置
	cfg := make(map[string]interface{})
	c.Context().PostArgs().VisitAll(func(key, value []byte) {
		cfg[string(key)] = string(value)
	})
	if err := h.manager.SaveConfig(name, cfg); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "配置已保存"})
}

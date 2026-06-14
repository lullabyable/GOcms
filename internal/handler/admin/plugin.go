package admin

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/model"
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

// Uninstall 卸载插件（支持 URL :name 参数或 body ids 参数）
func (h *PluginHandler) Uninstall(c *fiber.Ctx) error {
	// 优先从 URL 参数取 name
	name := c.Params("name")
	if name != "" {
		if err := h.manager.Uninstall(name); err != nil {
			return c.JSON(fiber.Map{"code": 0, "msg": err.Error()})
		}
		return c.JSON(fiber.Map{"code": 1, "msg": "卸载成功"})
	}

	// 从 body 取 ids（前端传 {ids: "1,2"}）
	ids := c.FormValue("ids")
	if ids == "" {
		var body struct {
			IDs string `json:"ids"`
		}
		if err := c.BodyParser(&body); err == nil && body.IDs != "" {
			ids = body.IDs
		}
	}
	if ids == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "缺少参数"})
	}

	idList := parseIDList(ids)
	for _, id := range idList {
		// 通过 ID 查插件名
		var p model.Plugin
		if err := h.manager.DB().Where("plugin_id = ?", id).First(&p).Error; err == nil {
			h.manager.Uninstall(p.Name)
		}
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

// Config 获取插件配置（支持 name 或 ID 查询）
func (h *PluginHandler) Config(c *fiber.Ctx) error {
	name := c.Params("name")

	// 如果 name 是纯数字，当作 ID 处理
	if id, err := strconv.Atoi(name); err == nil {
		var p model.Plugin
		if err := h.manager.DB().Where("plugin_id = ?", id).First(&p).Error; err != nil {
			return c.JSON(fiber.Map{"code": 0, "msg": "插件不存在"})
		}
		name = p.Name
	}

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

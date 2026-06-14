package admin

import (
	"fmt"
	"net/smtp"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/model"
	"gocms/internal/response"
	"gorm.io/gorm"
)

// SystemSettingsHandler 系统设置处理器
type SystemSettingsHandler struct {
	db *gorm.DB
}

func NewSystemSettingsHandler(db *gorm.DB) *SystemSettingsHandler {
	return &SystemSettingsHandler{db: db}
}

// configGroups 配置分组定义
var configGroups = map[string][]string{
	"site": {
		"site_name", "site_url", "site_keywords", "site_description",
		"site_icp", "site_tj", "site_logo", "template_dir", "mob_template_dir",
	},
	"user": {
		"user_status", "user_reg", "user_email_check", "user_yzm",
	},
	"player": {
		"player_parse", "player_list", "player_from",
	},
	"email": {
		"email_host", "email_port", "email_username", "email_password",
		"email_secure", "email_from", "email_from_name",
	},
	"cache": {
		"cache_open", "cache_type", "cache_time", "cache_host", "cache_port", "cache_password",
	},
	"safe": {
		"safe_yzm_type", "safe_yzm_len", "safe_email_bind", "safe_phone_bind", "safe_ip_blacklist",
	},
	"api": {
		"api_open", "api_token", "api_cache_time",
	},
	"seo": {
		"seo_vod_title", "seo_art_title", "seo_vod_description", "seo_art_description",
	},
	"performance": {
		"pagesize", "makesize", "search_vod_rule", "search_art_rule",
	},
}

// GetAllConfig 获取所有配置
func (h *SystemSettingsHandler) GetAllConfig(c *fiber.Ctx) error {
	var configs []model.Config
	h.db.Where("type = ?", "maccms").Find(&configs)

	result := make(map[string]string)
	for _, cfg := range configs {
		result[cfg.Name] = cfg.Value
	}

	// 按分组组织返回
	grouped := make(map[string]map[string]string)
	for group, keys := range configGroups {
		grouped[group] = make(map[string]string)
		for _, key := range keys {
			if val, ok := result[key]; ok {
				grouped[group][key] = val
			}
		}
	}

	return response.OK(c, fiber.Map{
		"all":    result,
		"grouped": grouped,
	})
}

// GetGroupConfig 获取指定分组配置
func (h *SystemSettingsHandler) GetGroupConfig(c *fiber.Ctx) error {
	group := c.Params("group")
	keys, ok := configGroups[group]
	if !ok {
		return response.Fail(c, "未知的配置分组: "+group)
	}

	var configs []model.Config
	h.db.Where("type = ? AND name IN ?", "maccms", keys).Find(&configs)

	result := make(map[string]string)
	for _, cfg := range configs {
		result[cfg.Name] = cfg.Value
	}
	// 确保所有 key 都有返回
	for _, key := range keys {
		if _, exists := result[key]; !exists {
			result[key] = ""
		}
	}

	return response.OK(c, result)
}

// SaveConfig 保存配置
func (h *SystemSettingsHandler) SaveConfig(c *fiber.Ctx) error {
	// 支持 JSON body 和 form 提交
	var body map[string]string
	if err := c.BodyParser(&body); err != nil {
		return response.Fail(c, "参数解析失败")
	}

	group := body["group"]
	delete(body, "group")

	// 如果指定了分组，只保存该分组的配置
	if group != "" {
		keys, ok := configGroups[group]
		if !ok {
			return response.Fail(c, "未知的配置分组")
		}
		allowed := make(map[string]bool)
		for _, k := range keys {
			allowed[k] = true
		}
		filtered := make(map[string]string)
		for k, v := range body {
			if allowed[k] {
				filtered[k] = v
			}
		}
		body = filtered
	}

	for name, value := range body {
		var existing model.Config
		if err := h.db.Where("type = ? AND name = ?", "maccms", name).First(&existing).Error; err != nil {
			h.db.Create(&model.Config{Type: "maccms", Name: name, Value: value})
		} else {
			h.db.Model(&existing).Update("value", value)
		}
	}

	return response.OKMsg(c, "配置已保存")
}

// TestEmail 测试邮箱配置
func (h *SystemSettingsHandler) TestEmail(c *fiber.Ctx) error {
	to := c.FormValue("to")
	if to == "" {
		return response.Fail(c, "请输入测试收件邮箱")
	}

	// 读取邮箱配置
	getConfig := func(name string) string {
		var cfg model.Config
		if err := h.db.Where("type = ? AND name = ?", "maccms", name).First(&cfg).Error; err != nil {
			return ""
		}
		return cfg.Value
	}

	host := getConfig("email_host")
	port := getConfig("email_port")
	username := getConfig("email_username")
	password := getConfig("email_password")
	secure := getConfig("email_secure")
	from := getConfig("email_from")
	fromName := getConfig("email_from_name")

	if host == "" || username == "" || password == "" {
		return response.Fail(c, "邮箱配置不完整")
	}
	if port == "" {
		port = "587"
	}
	if from == "" {
		from = username
	}
	if fromName == "" {
		fromName = "GOcms"
	}

	addr := host + ":" + port
	auth := smtp.PlainAuth("", username, password, host)

	msg := []byte(fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: GOcms 邮箱测试\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n这是一封来自 GOcms 的测试邮件。", fromName, from, to))

	var err error
	if secure == "ssl" || secure == "tls" {
		// SSL/TLS 连接（使用 465 端口）
		err = sendMailTLS(addr, auth, from, []string{to}, msg)
	} else {
		err = smtp.SendMail(addr, auth, from, []string{to}, msg)
	}

	if err != nil {
		return response.Fail(c, "发送失败: "+err.Error())
	}
	return response.OKMsg(c, "测试邮件已发送")
}

// TestCache 测试缓存连接
func (h *SystemSettingsHandler) TestCache(c *fiber.Ctx) error {
	getConfig := func(name string) string {
		var cfg model.Config
		if err := h.db.Where("type = ? AND name = ?", "maccms", name).First(&cfg).Error; err != nil {
			return ""
		}
		return cfg.Value
	}

	cacheType := getConfig("cache_type")
	if cacheType == "" {
		cacheType = "file"
	}

	switch cacheType {
	case "file":
		return response.OKMsg(c, "文件缓存连接正常")
	case "redis":
		host := getConfig("cache_host")
		port := getConfig("cache_port")
		password := getConfig("cache_password")
		if host == "" {
			return response.Fail(c, "Redis 主机未配置")
		}
		if port == "" {
			port = "6379"
		}
		// 尝试 TCP 连接测试
		addr := host + ":" + port
		_ = password // TODO: 实际 Redis 连接测试
		return response.OKMsg(c, fmt.Sprintf("Redis 连接测试: %s (需接入 redis 客户端)", addr))
	default:
		return response.Fail(c, "未知的缓存类型: "+cacheType)
	}
}

// sendMailTLS 通过 TLS 发送邮件（465 端口）
func sendMailTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// 简单实现：先尝试普通 SMTP
	// 生产环境应使用 crypto/tls + net/smtp 的 TLS 版本
	return smtp.SendMail(addr, auth, from, to, msg)
}

// GetConfigGroupList 获取所有配置分组列表
func (h *SystemSettingsHandler) GetConfigGroupList(c *fiber.Ctx) error {
	groups := make([]fiber.Map, 0)
	for name, keys := range configGroups {
		groups = append(groups, fiber.Map{
			"name":  name,
			"keys":  keys,
			"count": len(keys),
		})
	}
	return response.OK(c, groups)
}

// GetConfigByType 兼容旧接口：按 type 获取配置
func (h *SystemSettingsHandler) GetConfigByType(c *fiber.Ctx) error {
	typeName := c.Query("type", "maccms")
	var configs []model.Config
	h.db.Where("type = ?", typeName).Find(&configs)

	result := make(map[string]string)
	for _, cfg := range configs {
		result[cfg.Name] = cfg.Value
	}
	return response.OK(c, result)
}

// SaveConfigByType 兼容旧接口：按 type 保存配置
func (h *SystemSettingsHandler) SaveConfigByType(c *fiber.Ctx) error {
	typeName := c.FormValue("type", "maccms")

	configs := make(map[string]string)
	c.Context().PostArgs().VisitAll(func(key, value []byte) {
		k := string(key)
		if k != "type" {
			configs[k] = string(value)
		}
	})

	for name, value := range configs {
		var existing model.Config
		if err := h.db.Where("type = ? AND name = ?", typeName, name).First(&existing).Error; err != nil {
			h.db.Create(&model.Config{Type: typeName, Name: name, Value: value})
		} else {
			h.db.Model(&existing).Update("value", value)
		}
	}
	return response.OKMsg(c, "配置已保存")
}

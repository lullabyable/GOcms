package admin

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/model"
	"gocms/internal/response"
	"gorm.io/gorm"
)

// SafetyHandler 安全配置处理器
type SafetyHandler struct {
	db *gorm.DB
}

func NewSafetyHandler(db *gorm.DB) *SafetyHandler {
	return &SafetyHandler{db: db}
}

// IPBlacklist IP黑名单模型
type IPBlacklist struct {
	ID        int    `gorm:"primaryKey;column:id" json:"id"`
	IP        string `gorm:"column:ip;size:45;uniqueIndex" json:"ip"`
	Reason    string `gorm:"column:reason;size:500" json:"reason"`
	AdminName string `gorm:"column:admin_name;size:50" json:"admin_name"`
	ExpireAt  int64  `gorm:"column:expire_at" json:"expire_at"`
	CreatedAt int64  `gorm:"column:created_at" json:"created_at"`
}

func (IPBlacklist) TableName() string { return "mac_ip_blacklist" }

// GetConfig 获取安全配置
func (h *SafetyHandler) GetConfig(c *fiber.Ctx) error {
	var configs []model.Config
	keys := []string{"safe_yzm_type", "safe_yzm_len", "safe_email_bind", "safe_phone_bind", "safe_ip_blacklist"}
	h.db.Where("type = ? AND name IN ?", "maccms", keys).Find(&configs)

	result := make(map[string]string)
	for _, cfg := range configs {
		result[cfg.Name] = cfg.Value
	}
	for _, key := range keys {
		if _, exists := result[key]; !exists {
			result[key] = ""
		}
	}
	return response.OK(c, result)
}

// SaveConfig 保存安全配置
func (h *SafetyHandler) SaveConfig(c *fiber.Ctx) error {
	var body map[string]string
	if err := c.BodyParser(&body); err != nil {
		return response.Fail(c, "参数解析失败")
	}

	allowed := map[string]bool{
		"safe_yzm_type": true, "safe_yzm_len": true,
		"safe_email_bind": true, "safe_phone_bind": true, "safe_ip_blacklist": true,
	}

	for name, value := range body {
		if !allowed[name] {
			continue
		}
		var existing model.Config
		if err := h.db.Where("type = ? AND name = ?", "maccms", name).First(&existing).Error; err != nil {
			h.db.Create(&model.Config{Type: "maccms", Name: name, Value: value})
		} else {
			h.db.Model(&existing).Update("value", value)
		}
	}
	return response.OKMsg(c, "安全配置已保存")
}

// IPBlacklistList IP黑名单列表
func (h *SafetyHandler) IPBlacklistList(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	var total int64
	h.db.Model(&IPBlacklist{}).Count(&total)

	var list []IPBlacklist
	h.db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	return response.Page(c, list, total, page, pageSize)
}

// IPBlacklistAdd 添加IP封禁
func (h *SafetyHandler) IPBlacklistAdd(c *fiber.Ctx) error {
	ip := c.FormValue("ip")
	if ip == "" {
		return response.Fail(c, "IP 地址不能为空")
	}

	reason := c.FormValue("reason")
	adminName := c.FormValue("admin_name")

	entry := IPBlacklist{
		IP:        ip,
		Reason:    reason,
		AdminName: adminName,
		CreatedAt: time.Now().Unix(),
	}

	// 检查是否已存在
	var existing IPBlacklist
	if err := h.db.Where("ip = ?", ip).First(&existing).Error; err == nil {
		return response.Fail(c, "该 IP 已在黑名单中")
	}

	if err := h.db.Create(&entry).Error; err != nil {
		return response.Fail(c, "添加失败")
	}
	return response.OKMsg(c, "已添加到黑名单")
}

// IPBlacklistDelete 删除IP封禁
func (h *SafetyHandler) IPBlacklistDelete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.FormValue("id"))
	if id == 0 {
		// 支持按 IP 删除
		ip := c.FormValue("ip")
		if ip == "" {
			return response.Fail(c, "缺少参数")
		}
		h.db.Where("ip = ?", ip).Delete(&IPBlacklist{})
	} else {
		h.db.Delete(&IPBlacklist{}, id)
	}
	return response.OKMsg(c, "已移除")
}

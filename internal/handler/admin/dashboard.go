package admin

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// Index 仪表盘首页 — 浏览器请求返回管理面板 HTML，API 请求返回 JSON
func (h *DashboardHandler) Index(c *fiber.Ctx) error {
	accept := c.Get("Accept")
	if strings.Contains(accept, "text/html") {
		return c.SendFile("./web/static/admin.html")
	}
	return h.API(c)
}

// API 仪表盘纯 JSON 数据接口
func (h *DashboardHandler) API(c *fiber.Ctx) error {
	var stats struct {
		VodCount    int64 `json:"vod_count"`
		ArtCount    int64 `json:"art_count"`
		UserCount   int64 `json:"user_count"`
		AdminCount  int64 `json:"admin_count"`
		TodayVisits int64 `json:"today_visits"`
	}

	h.db.Table("mac_vod").Where("vod_status = 1").Count(&stats.VodCount)
	h.db.Table("mac_art").Where("art_status = 1").Count(&stats.ArtCount)
	h.db.Table("mac_user").Where("user_status = 1").Count(&stats.UserCount)
	h.db.Table("mac_admin").Where("admin_status = 1").Count(&stats.AdminCount)

	return c.JSON(fiber.Map{"code": 1, "data": stats})
}

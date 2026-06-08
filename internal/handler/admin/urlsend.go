package admin

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
	"gocms/internal/service/urlpush"
)

// URLSendHandler URL推送处理器
type URLSendHandler struct {
	db      *gorm.DB
	manager *urlpush.Manager
}

func NewURLSendHandler(db *gorm.DB, manager *urlpush.Manager) *URLSendHandler {
	return &URLSendHandler{db: db, manager: manager}
}

// PushURLs 手动推送URL
func (h *URLSendHandler) PushURLs(c *fiber.Ctx) error {
	platform := c.FormValue("platform", "all")
	urlsStr := c.FormValue("urls")

	if urlsStr == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "URL列表不能为空"})
	}

	// 解析URL列表（支持换行和逗号分隔）
	var urls []string
	for _, line := range strings.Split(urlsStr, "\n") {
		for _, u := range strings.Split(line, ",") {
			u = strings.TrimSpace(u)
			if u != "" {
				urls = append(urls, u)
			}
		}
	}

	if len(urls) == 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "没有有效的URL"})
	}

	var pushErrs []error
	if platform == "all" {
		pushErrs = h.manager.PushAll(urls)
	} else {
		if err := h.manager.PushTo(platform, urls); err != nil {
			pushErrs = append(pushErrs, err)
		}
	}

	if len(pushErrs) > 0 {
		errMsgs := make([]string, len(pushErrs))
		for i, e := range pushErrs {
			errMsgs[i] = e.Error()
		}
		return c.JSON(fiber.Map{
			"code": 0,
			"msg":  "部分推送失败",
			"data": fiber.Map{"errors": errMsgs},
		})
	}

	return c.JSON(fiber.Map{"code": 1, "msg": "推送成功", "data": fiber.Map{"total": len(urls)}})
}

// PushAll 推送所有未推送的URL
func (h *URLSendHandler) PushAll(c *fiber.Ctx) error {
	// 获取最新的视频URL
	var vods []model.Vod
	h.db.Where("vod_status = 1").Order("vod_id DESC").Limit(1000).Find(&vods)

	var urls []string
	siteURL := c.FormValue("site_url", "")
	if siteURL == "" {
		siteURL = "https://example.com"
	}

	for _, vod := range vods {
		urls = append(urls, siteURL+"/voddetail/"+strconv.Itoa(vod.ID))
	}

	if len(urls) == 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "没有可推送的URL"})
	}

	pushErrs := h.manager.PushAll(urls)
	if len(pushErrs) > 0 {
		errMsgs := make([]string, len(pushErrs))
		for i, e := range pushErrs {
			errMsgs[i] = e.Error()
		}
		return c.JSON(fiber.Map{
			"code": 0,
			"msg":  "部分推送失败",
			"data": fiber.Map{"errors": errMsgs},
		})
	}

	return c.JSON(fiber.Map{"code": 1, "msg": "推送成功", "data": fiber.Map{"total": len(urls)}})
}

// Logs 推送日志
func (h *URLSendHandler) Logs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	platform := c.Query("platform", "")

	logs, total, err := h.manager.GetLogs(platform, page, pageSize)
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "查询失败"})
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":      logs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// Config 获取/保存推送配置
func (h *URLSendHandler) Config(c *fiber.Ctx) error {
	if c.Method() == "GET" {
		// 从数据库读取配置
		var configs []model.Config
		h.db.Where("type = ?", "url_push").Find(&configs)

		cfgMap := make(map[string]string)
		for _, cfg := range configs {
			cfgMap[cfg.Name] = cfg.Value
		}

		return c.JSON(fiber.Map{"code": 1, "data": cfgMap})
	}

	// POST 保存配置
	configs := map[string]string{
		"baidu_site":   c.FormValue("baidu_site"),
		"baidu_token":  c.FormValue("baidu_token"),
		"shenma_site":  c.FormValue("shenma_site"),
		"shenma_token": c.FormValue("shenma_token"),
		"sogou_site":   c.FormValue("sogou_site"),
		"sogou_token":  c.FormValue("sogou_token"),
	}

	for name, value := range configs {
		var existing model.Config
		if err := h.db.Where("type = ? AND name = ?", "url_push", name).First(&existing).Error; err != nil {
			h.db.Create(&model.Config{Type: "url_push", Name: name, Value: value})
		} else {
			h.db.Model(&existing).Update("value", value)
		}
	}

	return c.JSON(fiber.Map{"code": 1, "msg": "配置已保存"})
}

// GenerateSitemap 生成站点地图URL列表
func (h *URLSendHandler) GenerateSitemap(c *fiber.Ctx) error {
	siteURL := c.FormValue("site_url", "https://example.com")
	contentType := c.FormValue("type", "vod")
	limit, _ := strconv.Atoi(c.FormValue("limit", "5000"))

	var urls []string

	switch contentType {
	case "vod":
		var vods []model.Vod
		h.db.Where("vod_status = 1").Order("vod_id DESC").Limit(limit).Find(&vods)
		for _, v := range vods {
			urls = append(urls, siteURL+"/voddetail/"+strconv.Itoa(v.ID))
		}
	case "art":
		var arts []model.Art
		h.db.Where("art_status = 1").Order("art_id DESC").Limit(limit).Find(&arts)
		for _, a := range arts {
			urls = append(urls, siteURL+"/artdetail/"+strconv.Itoa(a.ID))
		}
	case "all":
		var vods []model.Vod
		h.db.Where("vod_status = 1").Order("vod_id DESC").Limit(limit).Find(&vods)
		for _, v := range vods {
			urls = append(urls, siteURL+"/voddetail/"+strconv.Itoa(v.ID))
		}
		var arts []model.Art
		h.db.Where("art_status = 1").Order("art_id DESC").Limit(limit).Find(&arts)
		for _, a := range arts {
			urls = append(urls, siteURL+"/artdetail/"+strconv.Itoa(a.ID))
		}
		urls = append(urls, siteURL)
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"urls":  urls,
			"total": len(urls),
		},
	})
}

// TimerPushCheck 定时推送检查（供调度器调用）
func (h *URLSendHandler) TimerPushCheck() error {
	// 获取最近24小时更新的内容
	since := time.Now().Add(-24 * time.Hour).Unix()

	var vods []model.Vod
	h.db.Where("vod_status = 1 AND vod_time_make > ?", since).Find(&vods)

	var urls []string
	for _, v := range vods {
		urls = append(urls, "/voddetail/"+strconv.Itoa(v.ID))
	}

	if len(urls) == 0 {
		return nil
	}

	h.manager.PushAll(urls)
	return nil
}

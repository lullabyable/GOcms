package admin

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
	"gocms/internal/service/collect"
)

// CollectHandler 后台采集管理
type CollectHandler struct {
	db      *gorm.DB
	engine  *collect.Engine
	htmlCol *collect.HTMLCollector
}

// NewCollectHandler 创建采集管理Handler
func NewCollectHandler(db *gorm.DB) *CollectHandler {
	return &CollectHandler{
		db:      db,
		engine:   collect.NewEngine(db),
		htmlCol: collect.NewHTMLCollector(db),
	}
}

// TestConnection 测试资源站连通性
func (h *CollectHandler) TestConnection(c *fiber.Ctx) error {
	apiURL := c.FormValue("api_url")
	if apiURL == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "请输入API地址"})
	}

	source := collect.CollectSource{APIURL: apiURL}
	opts := collect.CollectOptions{Page: 1}
	result, err := h.engine.CollectFromSource(source, opts)
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "连接失败: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"msg":  "连接成功",
		"data": result,
	})
}

// StartCollect 执行采集
func (h *CollectHandler) StartCollect(c *fiber.Ctx) error {
	apiURL := c.FormValue("api_url")
	if apiURL == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "请输入API地址"})
	}

	source := collect.CollectSource{
		APIURL:      apiURL,
		UpRule:      c.FormValue("uprule", "a"),
		PicDownload: 0,
	}
	if v, _ := strconv.Atoi(c.FormValue("pic")); v > 0 {
		source.PicDownload = v
	}

	opts := collect.CollectOptions{
		Page:    1,
		TypeID:  0,
		Hours:   0,
		Keyword: "",
	}
	if v, _ := strconv.Atoi(c.FormValue("page")); v > 0 {
		opts.Page = v
	}
	if v, _ := strconv.Atoi(c.FormValue("type_id")); v > 0 {
		opts.TypeID = v
	}
	if v, _ := strconv.Atoi(c.FormValue("hours")); v > 0 {
		opts.Hours = v
	}
	opts.Keyword = c.FormValue("keyword")
	opts.IDs = c.FormValue("ids")

	result, err := h.engine.CollectFromSource(source, opts)
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "采集失败: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"msg":  "采集完成",
		"data": result,
	})
}

// CollectArt 采集文章（占位）
func (h *CollectHandler) CollectArt(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"code": 0, "msg": "文章采集暂未实现"})
}

// ==================== 后台弹幕管理 ====================

// DanmakuList 弹幕列表
func (h *CollectHandler) DanmakuList(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	vodID, _ := strconv.Atoi(c.Query("vod_id", "0"))

	var danmakus []model.Danmaku
	var total int64

	query := h.db.Model(&model.Danmaku{})
	if vodID > 0 {
		query = query.Where("vod_id = ?", vodID)
	}
	query.Count(&total)
	query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&danmakus)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{"list": danmakus, "total": total, "page": page, "page_size": pageSize},
	})
}

// DanmakuDelete 删除弹幕
func (h *CollectHandler) DanmakuDelete(c *fiber.Ctx) error {
	ids := c.FormValue("ids")
	if ids == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "请选择要删除的弹幕"})
	}
	// 简单的逗号分割解析
	idList := parseIDs(ids)
	h.db.Where("danmaku_id IN ?", idList).Delete(&model.Danmaku{})
	return c.JSON(fiber.Map{"code": 1, "msg": "删除成功"})
}

// parseIDs 解析逗号分隔的ID列表
func parseIDs(s string) []int {
	var ids []int
	for _, part := range splitAndTrim(s, ",") {
		if id, err := strconv.Atoi(part); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for _, part := range splitString(s, sep) {
		trimmed := trimWhitespace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitString(s string, sep string) []string {
	if len(sep) != 1 {
		return []string{s}
	}
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimWhitespace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

// ==================== 资源站管理 ====================

// CollectSourceModel 资源站模型
type CollectSourceModel struct {
	ID        int    `gorm:"primaryKey;column:id" json:"id"`
	Name      string `gorm:"column:name;size:100" json:"name"`
	APIURL    string `gorm:"column:api_url;size:500" json:"api_url"`
	UpRule    string `gorm:"column:up_rule;size:10" json:"up_rule"`
	Status    int    `gorm:"column:status" json:"status"`
	Priority  int    `gorm:"column:priority" json:"priority"`
	LastSync  int64  `gorm:"column:last_sync" json:"last_sync"`
	CreatedAt int64  `gorm:"column:created_at" json:"created_at"`
}

func (CollectSourceModel) TableName() string { return "mac_collect_source" }

// SourceList 资源站列表
func (h *CollectHandler) SourceList(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	var total int64
	h.db.Model(&CollectSourceModel{}).Count(&total)

	var list []CollectSourceModel
	h.db.Order("priority ASC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{"list": list, "total": total, "page": page, "page_size": pageSize},
	})
}

// SourceDetail 资源站详情
func (h *CollectHandler) SourceDetail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var source CollectSourceModel
	if err := h.db.First(&source, id).Error; err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "资源站不存在"})
	}
	return c.JSON(fiber.Map{"code": 1, "data": source})
}

// SourceSave 保存资源站
func (h *CollectHandler) SourceSave(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.FormValue("id"))
	name := c.FormValue("name")
	apiURL := c.FormValue("api_url")
	if name == "" || apiURL == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "名称和API地址不能为空"})
	}

	source := CollectSourceModel{
		Name:   name,
		APIURL: apiURL,
		UpRule: c.FormValue("up_rule"),
	}
	if s, err := strconv.Atoi(c.FormValue("status")); err == nil {
		source.Status = s
	}
	if p, err := strconv.Atoi(c.FormValue("priority")); err == nil {
		source.Priority = p
	}

	if id > 0 {
		source.ID = id
		h.db.Model(&CollectSourceModel{}).Where("id = ?", id).Updates(map[string]interface{}{
			"name":     source.Name,
			"api_url":  source.APIURL,
			"up_rule":  source.UpRule,
			"status":   source.Status,
			"priority": source.Priority,
		})
	} else {
		source.CreatedAt = time.Now().Unix()
		h.db.Create(&source)
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "保存成功"})
}

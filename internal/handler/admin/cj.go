package admin

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/response"
	"gorm.io/gorm"
)

// CJHandler 自定义采集处理器
type CJHandler struct {
	db *gorm.DB
}

func NewCJHandler(db *gorm.DB) *CJHandler {
	return &CJHandler{db: db}
}

// CJRule 采集规则模型
type CJRule struct {
	ID        int    `gorm:"primaryKey;column:id" json:"id"`
	Name      string `gorm:"column:name;size:100" json:"name"`
	SourceURL string `gorm:"column:source_url;size:500" json:"source_url"`
	ListRule  string `gorm:"column:list_rule;type:text" json:"list_rule"`
	DetailRule string `gorm:"column:detail_rule;type:text" json:"detail_rule"`
	TypeID    int    `gorm:"column:type_id" json:"type_id"`
	Status    int    `gorm:"column:status" json:"status"`
	LastRun   int64  `gorm:"column:last_run" json:"last_run"`
	CreatedAt int64  `gorm:"column:created_at" json:"created_at"`
}

func (CJRule) TableName() string { return "mac_cj_rule" }

// List 采集规则列表
func (h *CJHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	var total int64
	h.db.Model(&CJRule{}).Count(&total)

	var list []CJRule
	h.db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	return response.Page(c, list, total, page, pageSize)
}

// Detail 采集规则详情
func (h *CJHandler) Detail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var rule CJRule
	if err := h.db.First(&rule, id).Error; err != nil {
		return response.Fail(c, "规则不存在")
	}
	return response.OK(c, rule)
}

// Save 保存采集规则
func (h *CJHandler) Save(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.FormValue("id"))
	name := c.FormValue("name")
	if name == "" {
		return response.Fail(c, "规则名称不能为空")
	}

	rule := CJRule{
		Name:       name,
		SourceURL:  c.FormValue("source_url"),
		ListRule:   c.FormValue("list_rule"),
		DetailRule: c.FormValue("detail_rule"),
	}
	if t, err := strconv.Atoi(c.FormValue("type_id")); err == nil {
		rule.TypeID = t
	}
	if s, err := strconv.Atoi(c.FormValue("status")); err == nil {
		rule.Status = s
	}

	if id > 0 {
		rule.ID = id
		h.db.Model(&CJRule{}).Where("id = ?", id).Updates(map[string]interface{}{
			"name":        rule.Name,
			"source_url":  rule.SourceURL,
			"list_rule":   rule.ListRule,
			"detail_rule": rule.DetailRule,
			"type_id":     rule.TypeID,
			"status":      rule.Status,
		})
	} else {
		rule.CreatedAt = time.Now().Unix()
		if err := h.db.Create(&rule).Error; err != nil {
			return response.Fail(c, "保存失败")
		}
	}
	return response.OKMsg(c, "保存成功")
}

// Run 执行采集
func (h *CJHandler) Run(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.FormValue("id"))
	if id == 0 {
		return response.Fail(c, "缺少规则 ID")
	}

	var rule CJRule
	if err := h.db.First(&rule, id).Error; err != nil {
		return response.Fail(c, "规则不存在")
	}

	// TODO: 实际执行采集逻辑
	h.db.Model(&CJRule{}).Where("id = ?", id).Update("last_run", time.Now().Unix())

	return response.OKMsg(c, "采集任务已提交")
}

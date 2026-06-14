package admin

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/response"
	"gorm.io/gorm"
)

// DomainHandler 域名管理处理器
type DomainHandler struct {
	db *gorm.DB
}

func NewDomainHandler(db *gorm.DB) *DomainHandler {
	return &DomainHandler{db: db}
}

// Domain 域名模型
type Domain struct {
	ID        int    `gorm:"primaryKey;column:id" json:"id"`
	Domain    string `gorm:"column:domain;size:255;uniqueIndex" json:"domain"`
	Type      int    `gorm:"column:type" json:"type"` // 1=授权域名 2=跳转域名
	Status    int    `gorm:"column:status" json:"status"`
	Remark    string `gorm:"column:remark;size:500" json:"remark"`
	CreatedAt int64  `gorm:"column:created_at" json:"created_at"`
}

func (Domain) TableName() string { return "mac_domain" }

// List 域名列表
func (h *DomainHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	domainType, _ := strconv.Atoi(c.Query("type", "-1"))

	query := h.db.Model(&Domain{})
	if domainType >= 0 {
		query = query.Where("type = ?", domainType)
	}

	var total int64
	query.Count(&total)

	var list []Domain
	query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	return response.Page(c, list, total, page, pageSize)
}

// Save 保存域名
func (h *DomainHandler) Save(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.FormValue("id"))
	domainName := c.FormValue("domain")
	if domainName == "" {
		return response.Fail(c, "域名不能为空")
	}

	d := Domain{
		Domain: domainName,
		Remark: c.FormValue("remark"),
	}
	if t, err := strconv.Atoi(c.FormValue("type")); err == nil {
		d.Type = t
	}
	if s, err := strconv.Atoi(c.FormValue("status")); err == nil {
		d.Status = s
	}

	if id > 0 {
		d.ID = id
		h.db.Model(&Domain{}).Where("id = ?", id).Updates(map[string]interface{}{
			"domain": d.Domain,
			"type":   d.Type,
			"status": d.Status,
			"remark": d.Remark,
		})
	} else {
		d.CreatedAt = time.Now().Unix()
		if err := h.db.Create(&d).Error; err != nil {
			return response.Fail(c, "保存失败: "+err.Error())
		}
	}
	return response.OKMsg(c, "保存成功")
}

// Delete 删除域名
func (h *DomainHandler) Delete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.FormValue("id"))
	if id == 0 {
		return response.Fail(c, "缺少参数")
	}
	h.db.Delete(&Domain{}, id)
	return response.OKMsg(c, "删除成功")
}

package api

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/model"
	"gorm.io/gorm"
)

// ProvideHandler 资源站API（供其他站采集）
type ProvideHandler struct {
	db *gorm.DB
}

// NewProvideHandler 创建 ProvideHandler
func NewProvideHandler(db *gorm.DB) *ProvideHandler {
	return &ProvideHandler{db: db}
}

// ProvideAPI 苹果CMS标准协议API
// GET /api/provide/:ac
// 参数: ac=list|detail, ids, t(分类), pg(页码), h(小时), wd(关键词), p(每页数量)
func (h *ProvideHandler) ProvideAPI(c *fiber.Ctx) error {
	ac := c.Params("ac")
	output := c.Query("output", "json") // json 或 xml

	switch ac {
	case "list":
		return h.handleList(c, output)
	case "detail":
		return h.handleDetail(c, output)
	default:
		return h.outputResult(c, output, fiber.Map{
			"code": 1,
			"msg":  "参数ac错误，可选值: list, detail",
		})
	}
}

// handleList 列表接口
func (h *ProvideHandler) handleList(c *fiber.Ctx, output string) error {
	page, _ := strconv.Atoi(c.Query("pg", "1"))
	pageSize, _ := strconv.Atoi(c.Query("p", "20"))
	typeID, _ := strconv.Atoi(c.Query("t", "0"))
	hours, _ := strconv.Atoi(c.Query("h", "0"))
	keyword := c.Query("wd", "")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := h.db.Model(&model.Vod{}).Where("vod_status = 1")

	if typeID > 0 {
		query = query.Where("type_id = ?", typeID)
	}
	if hours > 0 {
		since := time.Now().Add(-time.Duration(hours) * time.Hour).Unix()
		query = query.Where("vod_time_make >= ?", since)
	}
	if keyword != "" {
		query = query.Where("vod_name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var vods []model.Vod
	query.Order("vod_id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&vods)

	// 转换为CMS格式
	list := make([]cmsVideoOutput, len(vods))
	for i, v := range vods {
		list[i] = h.vodToCMS(v)
	}

	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}

	result := fiber.Map{
		"code":  1,
		"msg":   "数据列表",
		"count": total,
		"pages": pages,
		"page":  page,
		"limit": pageSize,
		"list":  list,
	}

	return h.outputResult(c, output, result)
}

// handleDetail 详情接口
func (h *ProvideHandler) handleDetail(c *fiber.Ctx, output string) error {
	ids := c.Query("ids")
	if ids == "" {
		return h.outputResult(c, output, fiber.Map{"code": 0, "msg": "缺少ids参数"})
	}

	idList := parseIDList(ids)
	if len(idList) == 0 {
		return h.outputResult(c, output, fiber.Map{"code": 0, "msg": "invalid ids"})
	}
	var vods []model.Vod
	h.db.Where("vod_id IN ? AND vod_status = 1", idList).Find(&vods)

	list := make([]cmsVideoOutput, len(vods))
	for i, v := range vods {
		list[i] = h.vodToCMS(v)
	}

	result := fiber.Map{
		"code": 1,
		"msg":  "数据详情",
		"list": list,
	}

	return h.outputResult(c, output, result)
}

func parseIDList(raw string) []int {
	parts := strings.Split(raw, ",")
	ids := make([]int, 0, len(parts))
	for _, part := range parts {
		id, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

// cmsVideoOutput CMS视频输出格式
type cmsVideoOutput struct {
	XMLName     xml.Name `xml:"video" json:"-"`
	VodID       int      `json:"vod_id" xml:"vod_id"`
	TypeID      int      `json:"type_id" xml:"type_id"`
	TypeName    string   `json:"type_name" xml:"type_name"`
	VodName     string   `json:"vod_name" xml:"vod_name"`
	VodSub      string   `json:"vod_sub" xml:"vod_sub"`
	VodEn       string   `json:"vod_en" xml:"vod_en"`
	VodPic      string   `json:"vod_pic" xml:"vod_pic"`
	VodActor    string   `json:"vod_actor" xml:"vod_actor"`
	VodDirector string   `json:"vod_director" xml:"vod_director"`
	VodWriter   string   `json:"vod_writer" xml:"vod_writer"`
	VodBlurb    string   `json:"vod_blurb" xml:"vod_blurb"`
	VodRemarks  string   `json:"vod_remarks" xml:"vod_remarks"`
	VodContent  string   `json:"vod_content" xml:"vod_content"`
	VodPlayFrom string   `json:"vod_play_from" xml:"vod_play_from"`
	VodPlayURL  string   `json:"vod_play_url" xml:"vod_play_url"`
	VodDownFrom string   `json:"vod_down_from" xml:"vod_down_from"`
	VodDownURL  string   `json:"vod_down_url" xml:"vod_down_url"`
	VodArea     string   `json:"vod_area" xml:"vod_area"`
	VodLang     string   `json:"vod_lang" xml:"vod_lang"`
	VodYear     string   `json:"vod_year" xml:"vod_year"`
	VodClass    string   `json:"vod_class" xml:"vod_class"`
	VodTag      string   `json:"vod_tag" xml:"vod_tag"`
	VodScore    string   `json:"vod_score" xml:"vod_score"`
	VodPubdate  string   `json:"vod_pubdate" xml:"vod_pubdate"`
	VodDuration string   `json:"vod_duration" xml:"vod_duration"`
}

// vodToCMS 将 Vod 模型转为 CMS 输出格式
func (h *ProvideHandler) vodToCMS(v model.Vod) cmsVideoOutput {
	// 查询分类名
	typeName := ""
	var t model.Type
	if err := h.db.Where("type_id = ?", v.TypeID).First(&t).Error; err == nil {
		typeName = t.TypeName
	}

	return cmsVideoOutput{
		VodID:       v.ID,
		TypeID:      v.TypeID,
		TypeName:    typeName,
		VodName:     v.VodName,
		VodSub:      v.VodSub,
		VodEn:       v.VodEn,
		VodPic:      v.VodPic,
		VodActor:    v.VodActor,
		VodDirector: v.VodDirector,
		VodWriter:   v.VodWriter,
		VodBlurb:    v.VodBlurb,
		VodRemarks:  v.VodRemarks,
		VodContent:  v.VodContent,
		VodPlayFrom: v.VodPlayFrom,
		VodPlayURL:  v.VodPlayURL,
		VodDownFrom: v.VodDownFrom,
		VodDownURL:  v.VodDownURL,
		VodArea:     v.VodArea,
		VodLang:     v.VodLang,
		VodYear:     v.VodYear,
		VodClass:    v.VodClass,
		VodTag:      v.VodTag,
		VodScore:    v.VodScore,
		VodPubdate:  v.VodPubdate,
		VodDuration: v.VodDuration,
	}
}

// outputResult 根据格式输出结果
func (h *ProvideHandler) outputResult(c *fiber.Ctx, output string, data interface{}) error {
	switch output {
	case "xml":
		c.Set("Content-Type", "application/xml; charset=utf-8")
		xmlData, err := xml.MarshalIndent(data, "", "  ")
		if err != nil {
			return c.Status(500).SendString("XML编码错误")
		}
		return c.SendString(xml.Header + string(xmlData))
	default:
		return c.JSON(data)
	}
}

// ProvideSearch 搜索接口（供其他站调用）
func (h *ProvideHandler) ProvideSearch(c *fiber.Ctx) error {
	keyword := c.Query("wd", "")
	page, _ := strconv.Atoi(c.Query("pg", "1"))
	pageSize, _ := strconv.Atoi(c.Query("p", "20"))

	if keyword == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "缺少搜索关键词"})
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	query := h.db.Model(&model.Vod{}).Where("vod_status = 1").
		Where("vod_name LIKE ? OR vod_sub LIKE ? OR vod_actor LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	query.Count(&total)

	var vods []model.Vod
	query.Order("vod_id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&vods)

	list := make([]cmsVideoOutput, len(vods))
	for i, v := range vods {
		list[i] = h.vodToCMS(v)
	}

	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}

	return c.JSON(fiber.Map{
		"code":  1,
		"msg":   fmt.Sprintf("搜索结果: %s", keyword),
		"count": total,
		"pages": pages,
		"page":  page,
		"limit": pageSize,
		"list":  list,
	})
}

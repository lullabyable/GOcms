package collect

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"gorm.io/gorm"
	"gocms/internal/model"
)

// HTMLCollector 自定义HTML采集器
type HTMLCollector struct {
	db     *gorm.DB
	client *resty.Client
}

// NewHTMLCollector 创建HTML采集器
func NewHTMLCollector(db *gorm.DB) *HTMLCollector {
	client := resty.New()
	client.SetTimeout(30 * time.Second)
	client.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	client.SetRetryCount(2)

	return &HTMLCollector{
		db:     db,
		client: client,
	}
}

// CollectResult HTML采集结果
type HTMLCollectResult struct {
	Imported int    `json:"imported"`
	Errors   int    `json:"errors"`
	Message  string `json:"message"`
}

// CollectFromHTML 从HTML页面采集
func (h *HTMLCollector) CollectFromHTML(rule CollectRule) (*HTMLCollectResult, error) {
	result := &HTMLCollectResult{}

	for page := 1; page <= rule.MaxPage; page++ {
		listURL := strings.Replace(rule.URLPattern, "{page}", fmt.Sprintf("%d", page), 1)

		// 请求列表页
		body, err := h.fetchPage(listURL, rule.Charset)
		if err != nil {
			result.Errors++
			continue
		}

		// 解析列表页
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
		if err != nil {
			result.Errors++
			continue
		}

		// 提取详情页链接
		var detailURLs []string
		doc.Find(rule.ListSelector).Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists {
				detailURLs = append(detailURLs, h.resolveURL(listURL, href))
			}
		})

		if len(detailURLs) == 0 {
			break
		}

		// 逐条采集详情页
		for _, detailURL := range detailURLs {
			video, err := h.collectDetail(detailURL, rule)
			if err != nil {
				result.Errors++
				continue
			}

			// 入库
			engine := NewEngine(h.db)
			if err := engine.processVideo(video, CollectSource{}, CollectOptions{}); err != nil {
				result.Errors++
				continue
			}
			result.Imported++
		}
	}

	result.Message = fmt.Sprintf("采集完成: 导入 %d, 错误 %d", result.Imported, result.Errors)
	return result, nil
}

// collectDetail 采集详情页
func (h *HTMLCollector) collectDetail(url string, rule CollectRule) (SourceVideo, error) {
	body, err := h.fetchPage(url, rule.Charset)
	if err != nil {
		return SourceVideo{}, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return SourceVideo{}, err
	}

	video := SourceVideo{
		CustomFields: make(map[string]string),
	}

	// 根据规则提取字段
	for modelField, selector := range rule.ProgramConfig.Map {
		value := h.extractField(doc, selector, rule.ProgramConfig.Funcs[modelField])

		switch modelField {
		case "vod_name":
			video.Name = value
		case "vod_pic":
			video.Pic = value
		case "vod_content":
			video.Des = value
		case "vod_actor":
			video.Actor = value
		case "vod_director":
			video.Director = value
		case "vod_year":
			video.Year = value
		case "vod_area":
			video.Area = value
		case "vod_lang":
			video.Lang = value
		case "type_name":
			video.TypeName = value
		case "vod_remarks":
			video.Remarks = value
		}
	}

	// 自定义字段
	for _, custom := range rule.CustomizeConfig {
		value := h.extractField(doc, custom.Rule, "")
		video.CustomFields[custom.EnName] = value
	}

	return video, nil
}

// extractField 根据选择器提取字段值
func (h *HTMLCollector) extractField(doc *goquery.Document, selector string, funcs string) string {
	parts := strings.SplitN(selector, "@", 2)
	sel := parts[0]
	attr := ""
	if len(parts) > 1 {
		attr = parts[1]
	}

	var value string
	if attr == "" || attr == "text" {
		value = strings.TrimSpace(doc.Find(sel).Text())
	} else if attr == "html" {
		value, _ = doc.Find(sel).Html()
	} else {
		value, _ = doc.Find(sel).Attr(attr)
	}

	if funcs != "" {
		value = h.applyFuncs(value, funcs)
	}
	return value
}

// applyFuncs 应用处理函数链
func (h *HTMLCollector) applyFuncs(value string, funcs string) string {
	for _, fn := range strings.Split(funcs, "|") {
		fn = strings.TrimSpace(fn)
		switch fn {
		case "trim":
			value = strings.TrimSpace(value)
		case "strip_tags":
			value = stripHTMLTags(value)
		case "html2text":
			value = stripHTMLTags(value)
		}
	}
	return value
}

// fetchPage 请求页面（支持字符集转换）
func (h *HTMLCollector) fetchPage(url string, charset string) (string, error) {
	resp, err := h.client.R().Get(url)
	if err != nil {
		return "", err
	}

	body := resp.String()

	// 字符集转换
	if charset == "gbk" || charset == "gb2312" {
		reader := transform.NewReader(strings.NewReader(body), simplifiedchinese.GBK.NewDecoder())
		data, err := io.ReadAll(reader)
		if err == nil {
			body = string(data)
		}
	}

	return body, nil
}

// resolveURL 补全相对路径URL
func (h *HTMLCollector) resolveURL(baseURL, href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	idx := strings.Index(baseURL, "://")
	if idx < 0 {
		return href
	}
	scheme := baseURL[:idx+3]
	rest := baseURL[idx+3:]
	hostEnd := strings.Index(rest, "/")
	if hostEnd < 0 {
		hostEnd = len(rest)
	}
	host := rest[:hostEnd]

	if strings.HasPrefix(href, "/") {
		return scheme + host + href
	}
	basePath := rest[hostEnd:]
	lastSlash := strings.LastIndex(basePath, "/")
	if lastSlash >= 0 {
		basePath = basePath[:lastSlash+1]
	}
	return scheme + host + basePath + href
}

// stripHTMLTags 去除HTML标签
func stripHTMLTags(s string) string {
	result := make([]rune, 0, len(s))
	inTag := false
	for _, c := range s {
		if c == '<' {
			inTag = true
			continue
		}
		if c == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result = append(result, c)
		}
	}
	return strings.TrimSpace(string(result))
}

// _ 保证 model 包被引用
var _ = model.Vod{}

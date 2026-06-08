package template

import (
	"fmt"
	"html/template"
	"strings"
)

// Paginator 分页器
type Paginator struct {
	CurrentPage int
	TotalPages  int
	TotalItems  int
	PageSize    int
	BaseURL     string // URL模板，含 {page} 占位符
	PageParam   string // 分页参数占位符，默认 "{page}"
	ShowFirst   bool
	ShowLast    bool
	ShowPrev    bool
	ShowNext    bool
	ShowPages   int // 显示的页码数量
}

func NewPaginator(currentPage, totalItems, pageSize int, baseURL string) *Paginator {
	totalPages := totalItems / pageSize
	if totalItems%pageSize > 0 {
		totalPages++
	}
	if totalPages < 1 {
		totalPages = 1
	}
	if currentPage < 1 {
		currentPage = 1
	}
	if currentPage > totalPages {
		currentPage = totalPages
	}

	return &Paginator{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
		PageSize:    pageSize,
		BaseURL:     baseURL,
		PageParam:   "{page}",
		ShowFirst:   true,
		ShowLast:    true,
		ShowPrev:    true,
		ShowNext:    true,
		ShowPages:   5,
	}
}

// Render 渲染分页 HTML（兼容原 macCMS 分页样式）
func (p *Paginator) Render() template.HTML {
	if p.TotalPages <= 1 {
		return ""
	}

	var buf strings.Builder

	// 首页
	if p.ShowFirst && p.CurrentPage > 1 {
		buf.WriteString(fmt.Sprintf(`<a href="%s">首页</a>`, p.buildURL(1)))
	}

	// 上一页
	if p.ShowPrev && p.CurrentPage > 1 {
		buf.WriteString(fmt.Sprintf(`<a href="%s">上一页</a>`, p.buildURL(p.CurrentPage-1)))
	}

	// 页码
	start, end := p.calcPageRange()
	for i := start; i <= end; i++ {
		if i == p.CurrentPage {
			buf.WriteString(fmt.Sprintf(`<span class="current">%d</span>`, i))
		} else {
			buf.WriteString(fmt.Sprintf(`<a href="%s">%d</a>`, p.buildURL(i), i))
		}
	}

	// 下一页
	if p.ShowNext && p.CurrentPage < p.TotalPages {
		buf.WriteString(fmt.Sprintf(`<a href="%s">下一页</a>`, p.buildURL(p.CurrentPage+1)))
	}

	// 末页
	if p.ShowLast && p.CurrentPage < p.TotalPages {
		buf.WriteString(fmt.Sprintf(`<a href="%s">末页</a>`, p.buildURL(p.TotalPages)))
	}

	// 共 X 页 / X 条
	buf.WriteString(fmt.Sprintf(`<span class="total">共 %d 页 / %d 条</span>`,
		p.TotalPages, p.TotalItems))

	return template.HTML(buf.String())
}

func (p *Paginator) buildURL(page int) string {
	url := p.BaseURL
	url = strings.Replace(url, p.PageParam, fmt.Sprintf("%d", page), 1)
	return url
}

func (p *Paginator) calcPageRange() (int, int) {
	half := p.ShowPages / 2
	start := p.CurrentPage - half
	end := p.CurrentPage + half

	if start < 1 {
		start = 1
		end = p.ShowPages
		if end > p.TotalPages {
			end = p.TotalPages
		}
	}
	if end > p.TotalPages {
		end = p.TotalPages
		start = end - p.ShowPages + 1
		if start < 1 {
			start = 1
		}
	}
	return start, end
}

// HasPrev 是否有上一页
func (p *Paginator) HasPrev() bool { return p.CurrentPage > 1 }

// HasNext 是否有下一页
func (p *Paginator) HasNext() bool { return p.CurrentPage < p.TotalPages }

// PrevPage 上一页页码
func (p *Paginator) PrevPage() int {
	if p.CurrentPage > 1 {
		return p.CurrentPage - 1
	}
	return 1
}

// NextPage 下一页页码
func (p *Paginator) NextPage() int {
	if p.CurrentPage < p.TotalPages {
		return p.CurrentPage + 1
	}
	return p.TotalPages
}

// GetPaginatorTemplateFuncs 返回分页相关的模板函数
func GetPaginatorTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"maccms_page": func(p *Paginator) template.HTML {
			if p == nil {
				return ""
			}
			return p.Render()
		},
	}
}

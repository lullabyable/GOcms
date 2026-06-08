package template

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Engine 模板引擎
type Engine struct {
	dir      string
	funcMap   template.FuncMap
	templates map[string]*template.Template
}

func NewEngine(dir string) *Engine {
	return &Engine{
		dir:      dir,
		funcMap:   defaultFuncMap(),
		templates: make(map[string]*template.Template),
	}
}

// Load 加载所有模板
func (e *Engine) Load() error {
	// 加载前台模板
	pattern := filepath.Join(e.dir, "*.html")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	for _, file := range files {
		name := filepath.Base(file)
		tpl := template.New(name).Funcs(e.funcMap)
		tpl, err := tpl.ParseFiles(file)
		if err != nil {
			return err
		}
		// 加载 layout
		layout := filepath.Join(e.dir, "layout", "*.html")
		if layouts, _ := filepath.Glob(layout); len(layouts) > 0 {
			tpl, err = tpl.ParseGlob(layout)
			if err != nil {
				return err
			}
		}
		e.templates[name] = tpl
	}
	return nil
}

// Render 渲染模板
func (e *Engine) Render(w io.Writer, name string, data interface{}) error {
	tpl, ok := e.templates[name]
	if !ok {
		// 动态加载
		file := filepath.Join(e.dir, name)
		tpl = template.New(name).Funcs(e.funcMap)
		var err error
		tpl, err = tpl.ParseFiles(file)
		if err != nil {
			return err
		}
		e.templates[name] = tpl
	}
	return tpl.Execute(w, data)
}

// FiberRenderer 适配 Fiber 的渲染接口
func (e *Engine) FiberRenderer(c *fiber.Ctx, name string, data interface{}) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return e.Render(c.Context().Response.BodyWriter(), name, data)
}

// defaultFuncMap 默认模板函数
func defaultFuncMap() template.FuncMap {
	return template.FuncMap{
		// 字符串处理
		"substr": func(s string, start, length int) string {
			runes := []rune(s)
			if start >= len(runes) {
				return ""
			}
			end := start + length
			if end > len(runes) {
				end = len(runes)
			}
			return string(runes[start:end])
		},
		"default": func(val, defaultVal interface{}) interface{} {
			if val == nil || val == "" {
				return defaultVal
			}
			return val
		},
		"str_replace": strings.ReplaceAll,
		"nl2br": func(s string) template.HTML {
			return template.HTML(strings.ReplaceAll(s, "\n", "<br>"))
		},
		"date": func(ts int64) string {
			if ts == 0 {
				return ""
			}
			return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
		},
		// URL 生成（占位，后续接入 URL 规则引擎）
		"mac_url_vod_detail": func(info map[string]interface{}) string {
			if id, ok := info["vod_id"]; ok {
				return "/voddetail/" + fmt.Sprintf("%v", id) + ".html"
			}
			return "#"
		},
		"mac_url_vod_play": func(info map[string]interface{}, sid, nid int) string {
			if id, ok := info["vod_id"]; ok {
				return fmt.Sprintf("/vodplay/%v-%d-%d.html", id, sid, nid)
			}
			return "#"
		},
		"mac_url_art_detail": func(info map[string]interface{}) string {
			if id, ok := info["art_id"]; ok {
				return "/artdetail/" + fmt.Sprintf("%v", id) + ".html"
			}
			return "#"
		},
		"mac_url_type": func(info map[string]interface{}, page int) string {
			if id, ok := info["type_id"]; ok {
				return fmt.Sprintf("/vodtype/%v-%d.html", id, page)
			}
			return "#"
		},
		"mac_url_search": func(params map[string]interface{}) string {
			return "/vodsearch.html"
		},
	}
}

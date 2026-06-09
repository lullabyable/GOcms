package template

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Engine 模板引擎（支持多主题）
type Engine struct {
	baseDir   string             // web/templates
	theme     string             // 当前主题名，如 "default"
	funcMap   template.FuncMap
	templates map[string]*template.Template
	mu        sync.RWMutex
}

// NewEngine 创建模板引擎
// baseDir: 模板根目录 (如 "./web/templates")
// theme:   主题名 (如 "default")
func NewEngine(baseDir, theme string) *Engine {
	if theme == "" {
		theme = "default"
	}
	return &Engine{
		baseDir:   baseDir,
		theme:     theme,
		funcMap:   defaultFuncMap(),
		templates: make(map[string]*template.Template),
	}
}

// ThemeDir 当前主题目录
func (e *Engine) ThemeDir() string {
	return filepath.Join(e.baseDir, e.theme)
}

// Load 加载当前主题的所有模板
func (e *Engine) Load() error {
	themeDir := e.ThemeDir()

	// 检查主题目录是否存在
	if _, err := os.Stat(themeDir); os.IsNotExist(err) {
		return fmt.Errorf("主题目录不存在: %s", themeDir)
	}

	// 加载主题根目录下的所有 .html 文件
	pattern := filepath.Join(themeDir, "*.html")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("扫描模板文件失败: %w", err)
	}

	for _, file := range files {
		name := filepath.Base(file)
		if err := e.loadTemplate(name, file); err != nil {
			return fmt.Errorf("加载模板 %s 失败: %w", name, err)
		}
	}

	// 加载子目录中的模板 (如 partials/, layout/)
	subDirs, _ := filepath.Glob(filepath.Join(themeDir, "*"))
	for _, sub := range subDirs {
		info, err := os.Stat(sub)
		if err != nil || !info.IsDir() {
			continue
		}
		subName := filepath.Base(sub)
		subFiles, _ := filepath.Glob(filepath.Join(sub, "*.html"))
		for _, file := range subFiles {
			tplName := subName + "/" + filepath.Base(file)
			if err := e.loadTemplate(tplName, file); err != nil {
				return fmt.Errorf("加载子模板 %s 失败: %w", tplName, err)
			}
		}
	}

	return nil
}

// loadTemplate 加载单个模板（自动解析 include 引用）
func (e *Engine) loadTemplate(name, filePath string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	themeDir := e.ThemeDir()

	// 读取模板内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	tpl := template.New(name).Funcs(e.funcMap)

	// 解析主模板
	tpl, err = tpl.Parse(string(content))
	if err != nil {
		return fmt.Errorf("解析模板 %s: %w", name, err)
	}

	// 解析同目录下的其他模板作为 include 来源
	allFiles, _ := filepath.Glob(filepath.Join(themeDir, "*.html"))
	for _, f := range allFiles {
		if f == filePath {
			continue
		}
		includeContent, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		includeName := filepath.Base(f)
		tpl.New(includeName).Parse(string(includeContent))
	}

	// 解析 partials 子目录
	partialsDir := filepath.Join(themeDir, "partials")
	if partials, _ := filepath.Glob(filepath.Join(partialsDir, "*.html")); len(partials) > 0 {
		for _, f := range partials {
			includeContent, err := os.ReadFile(f)
			if err != nil {
				continue
			}
			includeName := "partials/" + filepath.Base(f)
			tpl.New(includeName).Parse(string(includeContent))
		}
	}

	e.templates[name] = tpl
	return nil
}

// Render 渲染模板到 writer
func (e *Engine) Render(w io.Writer, name string, data interface{}) error {
	e.mu.RLock()
	tpl, ok := e.templates[name]
	e.mu.RUnlock()

	if !ok {
		// 尝试动态加载
		file := filepath.Join(e.ThemeDir(), name)
		if err := e.loadTemplate(name, file); err != nil {
			return fmt.Errorf("模板 %s 未找到", name)
		}
		e.mu.RLock()
		tpl = e.templates[name]
		e.mu.RUnlock()
	}

	return tpl.Execute(w, data)
}

// RenderString 渲染模板并返回字符串
func (e *Engine) RenderString(name string, data interface{}) (string, error) {
	var buf strings.Builder
	if err := e.Render(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// FiberRenderer 适配 Fiber 渲染接口
func (e *Engine) FiberRenderer(c *fiber.Ctx, name string, data interface{}) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return e.Render(c.Context().Response.BodyWriter(), name, data)
}

// SetTheme 切换主题（热切换）
func (e *Engine) SetTheme(theme string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	newDir := filepath.Join(e.baseDir, theme)
	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		return fmt.Errorf("主题 %s 不存在", theme)
	}

	e.theme = theme
	e.templates = make(map[string]*template.Template)

	// 重新加载
	return e.Load()
}

// Theme 当前主题名
func (e *Engine) Theme() string {
	return e.theme
}

// ListThemes 列出所有可用主题
func (e *Engine) ListThemes() ([]string, error) {
	entries, err := os.ReadDir(e.baseDir)
	if err != nil {
		return nil, err
	}

	var themes []string
	for _, entry := range entries {
		if entry.IsDir() {
			// 检查目录下是否有 .html 文件
			htmls, _ := filepath.Glob(filepath.Join(e.baseDir, entry.Name(), "*.html"))
			if len(htmls) > 0 {
				themes = append(themes, entry.Name())
			}
		}
	}
	return themes, nil
}

// defaultFuncMap 默认模板函数
func defaultFuncMap() template.FuncMap {
	return template.FuncMap{
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
		"plus": func(a, b int) int { return a + b },
		"minus": func(a, b int) int { return a - b },
		"page_range": func(page, totalPages int) []int {
			start := page - 2
			if start < 1 {
				start = 1
			}
			end := start + 4
			if end > totalPages {
				end = totalPages
			}
			start = end - 4
			if start < 1 {
				start = 1
			}
			var nums []int
			for i := start; i <= end; i++ {
				nums = append(nums, i)
			}
			return nums
		},
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

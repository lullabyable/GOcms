package admin

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// PageHandler 后台页面渲染处理器（兼容原版 maccms10 模板）
type PageHandler struct {
	db          *gorm.DB
	templateDir string
	funcMap     template.FuncMap
	cache       map[string]*template.Template
	mu          sync.RWMutex
}

// NewPageHandler 创建后台页面处理器
func NewPageHandler(db *gorm.DB, templateDir string) *PageHandler {
	if templateDir == "" {
		templateDir = "./web/templates/admin"
	}
	h := &PageHandler{
		db:          db,
		templateDir: templateDir,
		cache:       make(map[string]*template.Template),
	}
	h.funcMap = h.buildFuncMap()
	return h
}

// Index 后台首页（主框架）
func (h *PageHandler) Index(c *fiber.Ctx) error {
	menus := h.getMenus()
	data := fiber.Map{
		"menus":   menus,
		"title":   "GOcms 后台管理",
		"config":  fiber.Map{"app": fiber.Map{"lang": "zh-cn"}},
		"version": "v10",
		"langs":   []string{"zh-cn"},
	}
	return h.renderPage(c, "index/index.html", data)
}

// Welcome 后台欢迎页
func (h *PageHandler) Welcome(c *fiber.Ctx) error {
	// 统计数据
	var vodCount, artCount, userCount, adminCount int64
	h.db.Table("mac_vod").Count(&vodCount)
	h.db.Table("mac_art").Count(&artCount)
	h.db.Table("mac_user").Count(&userCount)
	h.db.Table("mac_admin").Count(&adminCount)

	data := fiber.Map{
		"title":       "欢迎使用",
		"version":     "v10",
		"update_sql":  false,
		"mac_lang":    "zh-cn",
		"admin":       fiber.Map{"admin_name": "admin"},
		"vod_count":   vodCount,
		"art_count":   artCount,
		"user_count":  userCount,
		"admin_count": adminCount,
		"dashboard_data": fiber.Map{
			"vod_total":   vodCount,
			"art_total":   artCount,
			"user_total":  userCount,
			"admin_total": adminCount,
		},
		"os_data": fiber.Map{
			"cpu_usage": 0,
			"mem_total": 0,
			"mem_used":  0,
		},
		"show_os_guide": true,
	}
	return h.renderPage(c, "index/welcome.html", data)
}

// Login 登录页面
func (h *PageHandler) Login(c *fiber.Ctx) error {
	return h.renderPage(c, "index/login.html", fiber.Map{
		"title":      "登录",
		"background": "",
	})
}

// VodData 视频列表页
func (h *PageHandler) VodData(c *fiber.Ctx) error {
	return h.renderDataPage(c, "vod")
}

// VodInfo 视频编辑页
func (h *PageHandler) VodInfo(c *fiber.Ctx) error {
	return h.renderInfoPage(c, "vod")
}

// ArtData 文章列表页
func (h *PageHandler) ArtData(c *fiber.Ctx) error {
	return h.renderDataPage(c, "art")
}

// ArtInfo 文章编辑页
func (h *PageHandler) ArtInfo(c *fiber.Ctx) error {
	return h.renderInfoPage(c, "art")
}

// TopicData 专题列表页
func (h *PageHandler) TopicData(c *fiber.Ctx) error {
	return h.renderDataPage(c, "topic")
}

// TopicInfo 专题编辑页
func (h *PageHandler) TopicInfo(c *fiber.Ctx) error {
	return h.renderInfoPage(c, "topic")
}

// LinkIndex 链接列表页
func (h *PageHandler) LinkIndex(c *fiber.Ctx) error {
	return h.renderDataPage(c, "link")
}

// LinkInfo 链接编辑页
func (h *PageHandler) LinkInfo(c *fiber.Ctx) error {
	return h.renderInfoPage(c, "link")
}

// TypeIndex 分类列表页
func (h *PageHandler) TypeIndex(c *fiber.Ctx) error {
	return h.renderDataPage(c, "type")
}

// TypeInfo 分类编辑页
func (h *PageHandler) TypeInfo(c *fiber.Ctx) error {
	return h.renderInfoPage(c, "type")
}

// ActorData 演员列表页
func (h *PageHandler) ActorData(c *fiber.Ctx) error {
	return h.renderDataPage(c, "actor")
}

// ActorInfo 演员编辑页
func (h *PageHandler) ActorInfo(c *fiber.Ctx) error {
	return h.renderInfoPage(c, "actor")
}

// RoleData 角色列表页
func (h *PageHandler) RoleData(c *fiber.Ctx) error {
	return h.renderDataPage(c, "role")
}

// RoleInfo 角色编辑页
func (h *PageHandler) RoleInfo(c *fiber.Ctx) error {
	return h.renderInfoPage(c, "role")
}

// UserIndex 用户列表页
func (h *PageHandler) UserIndex(c *fiber.Ctx) error {
	return h.renderDataPage(c, "user")
}

// UserInfo 用户编辑页
func (h *PageHandler) UserInfo(c *fiber.Ctx) error {
	return h.renderInfoPage(c, "user")
}

// AdminIndex 管理员列表页
func (h *PageHandler) AdminIndex(c *fiber.Ctx) error {
	return h.renderDataPage(c, "admin")
}

// AdminInfo 管理员编辑页
func (h *PageHandler) AdminInfo(c *fiber.Ctx) error {
	return h.renderInfoPage(c, "admin")
}

// CommentIndex 评论列表页
func (h *PageHandler) CommentIndex(c *fiber.Ctx) error {
	return h.renderDataPage(c, "comment")
}

// GbookIndex 留言列表页
func (h *PageHandler) GbookIndex(c *fiber.Ctx) error {
	return h.renderDataPage(c, "gbook")
}

// DatabaseExport 数据库备份页
func (h *PageHandler) DatabaseExport(c *fiber.Ctx) error {
	return h.renderDataPage(c, "database")
}

// DatabaseSQL 数据库SQL页
func (h *PageHandler) DatabaseSQL(c *fiber.Ctx) error {
	return h.renderPage(c, "database/sql.html", fiber.Map{
		"title": "执行SQL",
		"param": fiber.Map{},
	})
}

// DatabaseRep 数据库替换页
func (h *PageHandler) DatabaseRep(c *fiber.Ctx) error {
	return h.renderPage(c, "database/rep.html", fiber.Map{
		"title": "数据替换",
		"param": fiber.Map{},
		"tables": []fiber.Map{
			{"name": "mac_vod", "fields": "vod_name,vod_sub,vod_actor,vod_director,vod_class,vod_area,vod_lang,vod_year,vod_content"},
			{"name": "mac_art", "fields": "art_name,art_sub,art_author,art_class,art_content"},
			{"name": "mac_topic", "fields": "topic_name,topic_content,topic_desc"},
		},
	})
}

// TemplateIndex 模板管理页
func (h *PageHandler) TemplateIndex(c *fiber.Ctx) error {
	return h.renderDataPage(c, "template")
}

// PlogIndex 操作日志页
func (h *PageHandler) PlogIndex(c *fiber.Ctx) error {
	return h.renderDataPage(c, "plog")
}

// SystemConfig 系统配置页
func (h *PageHandler) SystemConfig(c *fiber.Ctx) error {
	return h.renderPage(c, "system/config.html", fiber.Map{
		"title": "系统配置",
		"param": fiber.Map{},
	})
}

// CollectIndex 采集页
func (h *PageHandler) CollectIndex(c *fiber.Ctx) error {
	return h.renderDataPage(c, "collect")
}

// OrderIndex 订单列表页
func (h *PageHandler) OrderIndex(c *fiber.Ctx) error {
	return h.renderDataPage(c, "order")
}

// MangaData 漫画列表页
func (h *PageHandler) MangaData(c *fiber.Ctx) error {
	return h.renderDataPage(c, "manga")
}

// MangaInfo 漫画编辑页
func (h *PageHandler) MangaInfo(c *fiber.Ctx) error {
	return h.renderInfoPage(c, "manga")
}

// LiveIndex 直播列表页
func (h *PageHandler) LiveIndex(c *fiber.Ctx) error {
	return h.renderDataPage(c, "live")
}

// renderDataPage 渲染列表页（通用逻辑）
func (h *PageHandler) renderDataPage(c *fiber.Ctx, module string) error {
	page, _ := c.ParamsInt("page", 1)
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	limit := 20
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	param := fiber.Map{
		"page":   page,
		"limit":  limit,
		"status": c.Query("status", ""),
		"wd":     c.Query("wd", ""),
		"type":   c.Query("type", ""),
		"level":  c.Query("level", ""),
		"lock":   c.Query("lock", ""),
		"order":  c.Query("order", ""),
	}

	// 根据模块查询数据
	list, total := h.queryModuleData(module, page, limit, param)

	// 获取分类树（视频/文章需要）
	typeTree := h.getTypeTree(module)

	data := fiber.Map{
		"title":     h.getModuleTitle(module),
		"param":     param,
		"list":      list,
		"total":     total,
		"page":      page,
		"limit":     limit,
		"type_tree": typeTree,
	}

	tplName := module + "/index.html"
	// 有些模块的列表页叫 data.html
	if _, err := os.Stat(filepath.Join(h.templateDir, module, "index.html")); os.IsNotExist(err) {
		tplName = module + "/data.html"
		if _, err := os.Stat(filepath.Join(h.templateDir, tplName)); os.IsNotExist(err) {
			// 尝试其他可能的名称
			tplName = module + "/index.html"
		}
	}

	return h.renderPage(c, tplName, data)
}

// renderInfoPage 渲染编辑页（通用逻辑）
func (h *PageHandler) renderInfoPage(c *fiber.Ctx, module string) error {
	id, _ := c.ParamsInt("id", 0)
	if qid := c.Query("id"); qid != "" {
		fmt.Sscanf(qid, "%d", &id)
	}

	info := h.queryModuleInfo(module, id)
	typeTree := h.getTypeTree(module)

	data := fiber.Map{
		"title":     h.getModuleTitle(module) + " - 编辑",
		"info":      info,
		"type_tree": typeTree,
		"param":     fiber.Map{},
	}

	return h.renderPage(c, module+"/info.html", data)
}

// renderPage 渲染页面（预处理 ThinkPHP 语法后执行）
func (h *PageHandler) renderPage(c *fiber.Ctx, name string, data interface{}) error {
	tpl, err := h.getTemplate(name)
	if err != nil {
		return c.Status(500).SendString(fmt.Sprintf("模板加载失败: %s - %v", name, err))
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return tpl.Execute(c.Context().Response.BodyWriter(), data)
}

// getTemplate 获取（带缓存的）模板
func (h *PageHandler) getTemplate(name string) (*template.Template, error) {
	h.mu.RLock()
	tpl, ok := h.cache[name]
	h.mu.RUnlock()
	if ok {
		return tpl, nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// double check
	if tpl, ok = h.cache[name]; ok {
		return tpl, nil
	}

	tpl, err := h.loadTemplate(name)
	if err != nil {
		return nil, err
	}
	h.cache[name] = tpl
	return tpl, nil
}

// loadTemplate 加载并预处理单个模板
func (h *PageHandler) loadTemplate(name string) (*template.Template, error) {
	filePath := filepath.Join(h.templateDir, name)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取模板 %s 失败: %w", filePath, err)
	}

	processed := h.preprocess(string(content))

	tpl := template.New(name).Funcs(h.funcMap)

	// 加载 public/ 目录下的公共模板
	publicDir := filepath.Join(h.templateDir, "public")
	publicFiles, _ := filepath.Glob(filepath.Join(publicDir, "*.html"))
	for _, pf := range publicFiles {
		pContent, err := os.ReadFile(pf)
		if err != nil {
			continue
		}
		pName := "public/" + filepath.Base(pf)
		pProcessed := h.preprocess(string(pContent))
		tpl.New(pName).Parse(pProcessed)
	}

	tpl, err = tpl.Parse(processed)
	if err != nil {
		return nil, fmt.Errorf("解析模板 %s: %w", name, err)
	}

	return tpl, nil
}

// preprocess ThinkPHP → Go 模板语法转换
func (h *PageHandler) preprocess(content string) string {
	// 1. 路径常量
	content = strings.ReplaceAll(content, "__STATIC__", "/static")
	content = strings.ReplaceAll(content, "__ROOT__", "")

	// 2. include
	re := regexp.MustCompile(`\{include\s+file="[^"]*?(public/[^"]+)"\s*/?\s*\}`)
	content = re.ReplaceAllString(content, `{{template "$1" .}}`)

	// 也处理不带 public/ 前缀的 include
	re2 := regexp.MustCompile(`\{include\s+file="[^"]*?/((?:public/)?[^"/]+\.html)"\s*/?\s*\}`)
	content = re2.ReplaceAllStringFunc(content, func(m string) string {
		sub := re2.FindStringSubmatch(m)
		if len(sub) > 1 {
			name := sub[1]
			if !strings.HasPrefix(name, "public/") {
				name = "public/" + name
			}
			return fmt.Sprintf(`{{template "%s" .}}`, name)
		}
		return m
	})

	// 3. {:lang('key')} → {{lang "key"}}
	re = regexp.MustCompile(`\{:lang\(['"]([^'"]+)['"]\)\}`)
	content = re.ReplaceAllString(content, `{{lang "$1"}}`)

	// 4. {:url('ctrl/act')} → {{url "ctrl/act"}}
	re = regexp.MustCompile(`\{:url\(['"]([^'"]+)['"]\)\}`)
	content = re.ReplaceAllString(content, `{{url "$1"}}`)

	// 5. {:url('ctrl/act', $param)} → {{url "ctrl/act"}}
	re = regexp.MustCompile(`\{:url\(['"]([^'"]+)['"],\s*\$param\)\}`)
	content = re.ReplaceAllString(content, `{{url "$1"}}`)

	// 6. {:mac_url_*($vo)} → {{mac_url_* .}}
	re = regexp.MustCompile(`\{:(mac_url_\w+)\(\$vo\)\}`)
	content = re.ReplaceAllString(content, `{{$1 .}}`)

	// 7. 函数调用 {:func()} → {{func}}
	re = regexp.MustCompile(`\{:(\w+)\(\)\}`)
	content = re.ReplaceAllString(content, `{{$1}}`)

	// 8. {$var|filter1|filter2} → {{.Var}}
	re = regexp.MustCompile(`\{\$([\w.]+)\|[\w|]+\}`)
	content = re.ReplaceAllStringFunc(content, func(m string) string {
		sub := regexp.MustCompile(`\{\$([\w.]+)\|`).FindStringSubmatch(m)
		if len(sub) > 1 {
			return h.varToGo(sub[1])
		}
		return m
	})

	// 8b. {$var['key']} → {{index $var "key"}} (ThinkPHP 数组访问语法)
	re = regexp.MustCompile(`\{\$([\w]+)\['([\w]+)'\]\}`)
	content = re.ReplaceAllString(content, `{{index $$$1 "$2"}}`)

	// 9. {$var} → {{.Var}} 或 {{$var}} (在 range 内)
	re = regexp.MustCompile(`\{\$([\w.]+)\}`)
	content = re.ReplaceAllStringFunc(content, func(m string) string {
		sub := re.FindStringSubmatch(m)
		if len(sub) > 1 {
			return h.varToGo(sub[1])
		}
		return m
	})

	// 10. volist → range
	re = regexp.MustCompile(`\{volist\s+name="(\w+)"\s+id="(\w+)"(?:\s+key="(\w+)")?\s*\}`)
	content = re.ReplaceAllStringFunc(content, func(m string) string {
		sub := re.FindStringSubmatch(m)
		listName := capitalizeFirst(sub[1])
		if len(sub) > 3 && sub[3] != "" {
			return fmt.Sprintf(`{{range $%s, $%s := .%s}}`, sub[3], sub[2], listName)
		}
		return fmt.Sprintf(`{{range $%s := .%s}}`, sub[2], listName)
	})
	content = strings.ReplaceAll(content, "{/volist}", "{{end}}")

	// 11. if conditions
	content = h.convertIfBlocks(content)
	content = strings.ReplaceAll(content, "{/if}", "{{end}}")
	content = strings.ReplaceAll(content, "{else /}", "{{else}}")
	content = strings.ReplaceAll(content, "{else}", "{{else}}")

	// 12. PHP 嵌入代码块 → 删除
	re = regexp.MustCompile(`(?s)<\?php.*?\?>`)
	content = re.ReplaceAllString(content, "")

	// 13. 删除引用 $GLOBALS 的 if 块（PHP 特有，Go 不需要）
	re = regexp.MustCompile(`(?s)\{if\s+condition="[^"]*\$GLOBALS[^"]*"\s*\}.*?\{/if\}`)
	content = re.ReplaceAllString(content, "")

	return content
}

// varToGo 变量路径转换
func (h *PageHandler) varToGo(path string) string {
	parts := strings.Split(path, ".")
	if len(parts) == 1 {
		return "." + capitalizeFirst(parts[0])
	}
	// 在 range 中的 $vo.field → $vo.field (保持原样，Go 模板可以通过 index 访问 map)
	return "$" + path
}

// convertIfBlocks 转换 if 条件块
func (h *PageHandler) convertIfBlocks(content string) string {
	// {if condition="$var eq 'value'"} → {{if eq .Var "value"}}
	re := regexp.MustCompile(`\{if\s+condition="([^"]+)"\s*\}`)
	content = re.ReplaceAllStringFunc(content, func(m string) string {
		sub := re.FindStringSubmatch(m)
		if len(sub) > 1 {
			return h.convertCond(sub[1], false)
		}
		return m
	})

	// {elseif condition="..."} → {{else if ...}}
	re = regexp.MustCompile(`\{elseif\s+condition="([^"]+)"\s*/?\s*\}`)
	content = re.ReplaceAllStringFunc(content, func(m string) string {
		sub := re.FindStringSubmatch(m)
		if len(sub) > 1 {
			return h.convertCond(sub[1], true)
		}
		return m
	})

	return content
}

// convertCond 转换单个条件
func (h *PageHandler) convertCond(cond string, isElse bool) string {
	prefix := "{{if"
	if isElse {
		prefix = "{{else if"
	}

	cond = strings.TrimSpace(cond)

	// $var eq 'value'
	re := regexp.MustCompile(`^\$([\w.]+)\s+eq\s+'([^']*)'$`)
	if m := re.FindStringSubmatch(cond); len(m) > 0 {
		return fmt.Sprintf(`%s eq %s "%s"}}`, prefix, h.varToGo(m[1]), m[2])
	}

	// $var eq "value"
	re = regexp.MustCompile(`^\$([\w.]+)\s+eq\s+"([^"]*)"$`)
	if m := re.FindStringSubmatch(cond); len(m) > 0 {
		return fmt.Sprintf(`%s eq %s "%s"}}`, prefix, h.varToGo(m[1]), m[2])
	}

	// $var neq 'value'
	re = regexp.MustCompile(`^\$([\w.]+)\s+neq\s+'([^']*)'$`)
	if m := re.FindStringSubmatch(cond); len(m) > 0 {
		return fmt.Sprintf(`%s ne %s "%s"}}`, prefix, h.varToGo(m[1]), m[2])
	}

	// $var eq numeric
	re = regexp.MustCompile(`^\$([\w.]+)\s+eq\s+(\d+)$`)
	if m := re.FindStringSubmatch(cond); len(m) > 0 {
		return fmt.Sprintf(`%s eq (int %s) %s}}`, prefix, h.varToGo(m[1]), m[2])
	}

	// $var neq ''
	re = regexp.MustCompile(`^\$([\w.]+)\s+neq\s+''$`)
	if m := re.FindStringSubmatch(cond); len(m) > 0 {
		return fmt.Sprintf(`%s ne %s ""}}`, prefix, h.varToGo(m[1]))
	}

	// $var eq ''
	re = regexp.MustCompile(`^\$([\w.]+)\s+eq\s+''$`)
	if m := re.FindStringSubmatch(cond); len(m) > 0 {
		return fmt.Sprintf(`%s eq %s ""}}`, prefix, h.varToGo(m[1]))
	}

	// $var gt numeric
	re = regexp.MustCompile(`^\$([\w.]+)\s+gt\s+(\d+)$`)
	if m := re.FindStringSubmatch(cond); len(m) > 0 {
		return fmt.Sprintf(`%s gt (int %s) %s}}`, prefix, h.varToGo(m[1]), m[2])
	}

	// $var lt numeric
	re = regexp.MustCompile(`^\$([\w.]+)\s+lt\s+(\d+)$`)
	if m := re.FindStringSubmatch(cond); len(m) > 0 {
		return fmt.Sprintf(`%s lt (int %s) %s}}`, prefix, h.varToGo(m[1]), m[2])
	}

	// fallback
	return fmt.Sprintf(`%s .%s}}`, prefix, capitalizeFirst(strings.TrimPrefix(cond, "$")))
}

// buildFuncMap 构建模板函数映射
func (h *PageHandler) buildFuncMap() template.FuncMap {
	return template.FuncMap{
		"lang": func(key string) string {
			if v, ok := langMap[key]; ok {
				return v
			}
			return key
		},
		"url": func(path string) string {
			return "/admin/" + path
		},
		"mac_filter_xss":         func(s string) string { return s },
		"mac_restore_htmlfilter": func(s string) string { return s },
		"mac_day": func(s string, args ...string) string { return s },
		"htmlspecialchars": func(s string) string { return template.HTMLEscapeString(s) },
		"str_replace": func(args ...interface{}) string {
			if len(args) >= 3 {
				search, _ := args[0].(string)
				replace, _ := args[1].(string)
				subject, _ := args[2].(string)
				return strings.ReplaceAll(subject, search, replace)
			}
			return ""
		},
		"mac_url_vod_detail": func(vo map[string]interface{}) string {
			if id, ok := vo["vod_id"]; ok {
				return fmt.Sprintf("/voddetail/%v.html", id)
			}
			return "#"
		},
		"mac_url_topic_detail": func(vo map[string]interface{}) string {
			if id, ok := vo["topic_id"]; ok {
				return fmt.Sprintf("/topicdetail/%v.html", id)
			}
			return "#"
		},
		"mac_url_art_detail": func(vo map[string]interface{}) string {
			if id, ok := vo["art_id"]; ok {
				return fmt.Sprintf("/artdetail/%v.html", id)
			}
			return "#"
		},
		"int": func(v interface{}) int {
			switch val := v.(type) {
			case int:
				return val
			case int64:
				return int(val)
			case float64:
				return int(val)
			default:
				return 0
			}
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
	}
}

// queryModuleData 查询模块列表数据
func (h *PageHandler) queryModuleData(module string, page, limit int, param fiber.Map) ([]map[string]interface{}, int64) {
	var total int64
	var list []map[string]interface{}

	offset := (page - 1) * limit

	switch module {
	case "vod":
		query := h.db.Table("mac_vod")
		if v, ok := param["status"].(string); ok && v != "" {
			query = query.Where("vod_status = ?", v)
		}
		if v, ok := param["type"].(string); ok && v != "" {
			query = query.Where("type_id = ?", v)
		}
		if v, ok := param["wd"].(string); ok && v != "" {
			query = query.Where("vod_name LIKE ?", "%"+v+"%")
		}
		if v, ok := param["level"].(string); ok && v != "" {
			query = query.Where("vod_level = ?", v)
		}
		if v, ok := param["lock"].(string); ok && v != "" {
			query = query.Where("vod_lock = ?", v)
		}
		query.Count(&total)
		query.Order("vod_id DESC").Offset(offset).Limit(limit).Find(&list)
		// 添加 type 信息
		h.enrichWithType(list, "type_id")

	case "art":
		query := h.db.Table("mac_art")
		if v, ok := param["status"].(string); ok && v != "" {
			query = query.Where("art_status = ?", v)
		}
		if v, ok := param["wd"].(string); ok && v != "" {
			query = query.Where("art_name LIKE ?", "%"+v+"%")
		}
		query.Count(&total)
		query.Order("art_id DESC").Offset(offset).Limit(limit).Find(&list)

	case "topic":
		query := h.db.Table("mac_topic")
		if v, ok := param["status"].(string); ok && v != "" {
			query = query.Where("topic_status = ?", v)
		}
		if v, ok := param["wd"].(string); ok && v != "" {
			query = query.Where("topic_name LIKE ?", "%"+v+"%")
		}
		query.Count(&total)
		query.Order("topic_id DESC").Offset(offset).Limit(limit).Find(&list)

	case "link":
		h.db.Table("mac_link").Count(&total)
		h.db.Table("mac_link").Order("link_sort ASC, link_id DESC").Offset(offset).Limit(limit).Find(&list)

	case "type":
		h.db.Table("mac_type").Count(&total)
		h.db.Table("mac_type").Order("type_id ASC").Offset(offset).Limit(limit).Find(&list)

	case "actor":
		query := h.db.Table("mac_actor")
		if v, ok := param["wd"].(string); ok && v != "" {
			query = query.Where("actor_name LIKE ?", "%"+v+"%")
		}
		query.Count(&total)
		query.Order("actor_id DESC").Offset(offset).Limit(limit).Find(&list)

	case "role":
		query := h.db.Table("mac_role")
		if v, ok := param["wd"].(string); ok && v != "" {
			query = query.Where("role_name LIKE ?", "%"+v+"%")
		}
		query.Count(&total)
		query.Order("role_id DESC").Offset(offset).Limit(limit).Find(&list)

	case "user":
		query := h.db.Table("mac_user")
		if v, ok := param["wd"].(string); ok && v != "" {
			query = query.Where("user_name LIKE ?", "%"+v+"%")
		}
		query.Count(&total)
		query.Order("user_id DESC").Offset(offset).Limit(limit).Find(&list)

	case "admin":
		h.db.Table("mac_admin").Count(&total)
		h.db.Table("mac_admin").Order("admin_id ASC").Offset(offset).Limit(limit).Find(&list)

	case "comment":
		h.db.Table("mac_comment").Count(&total)
		h.db.Table("mac_comment").Order("comment_id DESC").Offset(offset).Limit(limit).Find(&list)

	case "gbook":
		h.db.Table("mac_gbook").Count(&total)
		h.db.Table("mac_gbook").Order("gbook_id DESC").Offset(offset).Limit(limit).Find(&list)

	case "order":
		h.db.Table("mac_order").Count(&total)
		h.db.Table("mac_order").Order("order_id DESC").Offset(offset).Limit(limit).Find(&list)

	case "manga":
		query := h.db.Table("mac_manga")
		if v, ok := param["status"].(string); ok && v != "" {
			query = query.Where("manga_status = ?", v)
		}
		if v, ok := param["wd"].(string); ok && v != "" {
			query = query.Where("manga_name LIKE ?", "%"+v+"%")
		}
		query.Count(&total)
		query.Order("manga_id DESC").Offset(offset).Limit(limit).Find(&list)

	case "plog":
		query := h.db.Table("mac_plog")
		if v, ok := param["wd"].(string); ok && v != "" {
			query = query.Where("plog_content LIKE ?", "%"+v+"%")
		}
		query.Count(&total)
		query.Order("plog_id DESC").Offset(offset).Limit(limit).Find(&list)

	case "collect":
		// 采集资源站列表
		h.db.Table("mac_collect").Count(&total)
		h.db.Table("mac_collect").Order("collect_id DESC").Offset(offset).Limit(limit).Find(&list)

	case "live":
		h.db.Table("mac_live").Count(&total)
		h.db.Table("mac_live").Order("live_sort ASC, live_id DESC").Offset(offset).Limit(limit).Find(&list)

	default:
		return nil, 0
	}

	return list, total
}

// queryModuleInfo 查询单条记录
func (h *PageHandler) queryModuleInfo(module string, id int) map[string]interface{} {
	var info map[string]interface{}

	switch module {
	case "vod":
		h.db.Table("mac_vod").Where("vod_id = ?", id).First(&info)
	case "art":
		h.db.Table("mac_art").Where("art_id = ?", id).First(&info)
	case "topic":
		h.db.Table("mac_topic").Where("topic_id = ?", id).First(&info)
	case "link":
		h.db.Table("mac_link").Where("link_id = ?", id).First(&info)
	case "type":
		h.db.Table("mac_type").Where("type_id = ?", id).First(&info)
	case "actor":
		h.db.Table("mac_actor").Where("actor_id = ?", id).First(&info)
	case "role":
		h.db.Table("mac_role").Where("role_id = ?", id).First(&info)
	case "user":
		h.db.Table("mac_user").Where("user_id = ?", id).First(&info)
	case "admin":
		h.db.Table("mac_admin").Where("admin_id = ?", id).First(&info)
	case "manga":
		h.db.Table("mac_manga").Where("manga_id = ?", id).First(&info)
	default:
		info = make(map[string]interface{})
	}

	return info
}

// getTypeTree 获取分类树
func (h *PageHandler) getTypeTree(module string) []map[string]interface{} {
	mid := 0
	switch module {
	case "vod":
		mid = 1
	case "art":
		mid = 2
	case "manga":
		mid = 3
	default:
		return nil
	}

	var types []map[string]interface{}
	h.db.Table("mac_type").Where("type_mid = ?", mid).Order("type_sort ASC, type_id ASC").Find(&types)

	// 构建树结构
	var tree []map[string]interface{}
	for _, t := range types {
		pid := 0
		if v, ok := t["type_pid"]; ok {
			switch vv := v.(type) {
			case int:
				pid = vv
			case int64:
				pid = int(vv)
			}
		}
		if pid == 0 {
			t["child"] = []map[string]interface{}{}
			tree = append(tree, t)
		}
	}
	// 添加子分类
	for _, parent := range tree {
		parentID := 0
		if v, ok := parent["type_id"]; ok {
			switch vv := v.(type) {
			case int:
				parentID = vv
			case int64:
				parentID = int(vv)
			}
		}
		for _, t := range types {
			pid := 0
			if v, ok := t["type_pid"]; ok {
				switch vv := v.(type) {
				case int:
					pid = vv
				case int64:
					pid = int(vv)
				}
			}
			if pid == parentID {
				if children, ok := parent["child"].([]map[string]interface{}); ok {
					parent["child"] = append(children, t)
				}
			}
		}
	}

	return tree
}

// enrichWithType 为列表添加分类名称
func (h *PageHandler) enrichWithType(list []map[string]interface{}, typeKey string) {
	if len(list) == 0 {
		return
	}
	// 收集所有 type_id
	typeIDs := make(map[int]bool)
	for _, item := range list {
		if v, ok := item[typeKey]; ok {
			switch vv := v.(type) {
			case int:
				typeIDs[vv] = true
			case int64:
				typeIDs[int(vv)] = true
			}
		}
	}
	if len(typeIDs) == 0 {
		return
	}

	ids := make([]int, 0, len(typeIDs))
	for id := range typeIDs {
		ids = append(ids, id)
	}

	var types []map[string]interface{}
	h.db.Table("mac_type").Where("type_id IN ?", ids).Find(&types)

	typeMap := make(map[int]string)
	for _, t := range types {
		id := 0
		name := ""
		if v, ok := t["type_id"]; ok {
			switch vv := v.(type) {
			case int:
				id = vv
			case int64:
				id = int(vv)
			}
		}
		if v, ok := t["type_name"]; ok {
			name, _ = v.(string)
		}
		typeMap[id] = name
	}

	for _, item := range list {
		tid := 0
		if v, ok := item[typeKey]; ok {
			switch vv := v.(type) {
			case int:
				tid = vv
			case int64:
				tid = int(vv)
			}
		}
		if name, ok := typeMap[tid]; ok {
			item["type"] = map[string]interface{}{"type_name": name, "type_id": tid}
		}
	}
}

// getModuleTitle 获取模块标题
func (h *PageHandler) getModuleTitle(module string) string {
	titles := map[string]string{
		"vod":     "视频管理",
		"art":     "文章管理",
		"topic":   "专题管理",
		"link":    "友情链接",
		"type":    "分类管理",
		"actor":   "演员管理",
		"role":    "角色管理",
		"user":    "用户管理",
		"admin":   "管理员",
		"comment": "评论管理",
		"gbook":   "留言管理",
		"order":   "订单管理",
		"manga":   "漫画管理",
		"plog":    "操作日志",
		"collect": "采集管理",
		"live":    "直播管理",
		"database": "数据库",
		"template": "模板管理",
	}
	if v, ok := titles[module]; ok {
		return v
	}
	return module
}

// getMenus 获取后台菜单结构
func (h *PageHandler) getMenus() []fiber.Map {
	return []fiber.Map{
		{
			"name": "首页",
			"icon": "xe625",
			"sub": []fiber.Map{
				{"name": "后台首页", "url": "/admin/page/welcome", "icon": ""},
			},
		},
		{
			"name": "系统",
			"icon": "xe62e",
			"sub": []fiber.Map{
				{"name": "系统配置", "url": "/admin/system/config", "icon": ""},
				{"name": "定时任务", "url": "/admin/timming/list", "icon": ""},
			},
		},
		{
			"name": "基础",
			"icon": "xe64b",
			"sub": []fiber.Map{
				{"name": "分类管理", "url": "/admin/page/type/index", "icon": ""},
				{"name": "专题管理", "url": "/admin/page/topic/data", "icon": ""},
				{"name": "友情链接", "url": "/admin/page/link/index", "icon": ""},
				{"name": "留言管理", "url": "/admin/page/gbook/data", "icon": ""},
				{"name": "评论管理", "url": "/admin/page/comment/data", "icon": ""},
			},
		},
		{
			"name": "视频",
			"icon": "xe639",
			"sub": []fiber.Map{
				{"name": "视频数据", "url": "/admin/page/vod/data", "icon": ""},
				{"name": "添加视频", "url": "/admin/page/vod/info", "icon": ""},
				{"name": "演员管理", "url": "/admin/page/actor/data", "icon": ""},
				{"name": "角色管理", "url": "/admin/page/role/data", "icon": ""},
			},
		},
		{
			"name": "文章",
			"icon": "xe616",
			"sub": []fiber.Map{
				{"name": "文章数据", "url": "/admin/page/art/data", "icon": ""},
				{"name": "添加文章", "url": "/admin/page/art/info", "icon": ""},
			},
		},
		{
			"name": "漫画",
			"icon": "xe616",
			"sub": []fiber.Map{
				{"name": "漫画数据", "url": "/admin/page/manga/data", "icon": ""},
				{"name": "添加漫画", "url": "/admin/page/manga/info", "icon": ""},
			},
		},
		{
			"name": "直播",
			"icon": "xe62b",
			"sub": []fiber.Map{
				{"name": "直播频道", "url": "/admin/page/live/index", "icon": ""},
			},
		},
		{
			"name": "用户",
			"icon": "xe62c",
			"sub": []fiber.Map{
				{"name": "管理员", "url": "/admin/page/admin/index", "icon": ""},
				{"name": "用户管理", "url": "/admin/page/user/data", "icon": ""},
				{"name": "订单管理", "url": "/admin/page/order/index", "icon": ""},
				{"name": "操作日志", "url": "/admin/page/plog/index", "icon": ""},
			},
		},
		{
			"name": "模板",
			"icon": "xe72d",
			"sub": []fiber.Map{
				{"name": "模板管理", "url": "/admin/page/template/index", "icon": ""},
			},
		},
		{
			"name": "采集",
			"icon": "xe727",
			"sub": []fiber.Map{
				{"name": "资源站", "url": "/admin/page/collect/index", "icon": ""},
			},
		},
		{
			"name": "数据库",
			"icon": "xe621",
			"sub": []fiber.Map{
				{"name": "数据库备份", "url": "/admin/page/database/export", "icon": ""},
				{"name": "执行SQL", "url": "/admin/page/database/sql", "icon": ""},
				{"name": "数据替换", "url": "/admin/page/database/rep", "icon": ""},
			},
		},
		{
			"name": "应用",
			"icon": "xe621",
			"sub": []fiber.Map{
				{"name": "插件管理", "url": "/admin/plugin/list", "icon": ""},
				{"name": "URL推送", "url": "/admin/urlsend/config", "icon": ""},
			},
		},
	}
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// langMap 中文语言包
var langMap = map[string]string{
	"admin/public/head/title":         "苹果CMS后台",
	"admin/index/index/name":          "GOcms 后台管理",
	"admin/index/index/menu_switch":   "菜单切换",
	"admin/index/index/menu_welcome":  "首页",
	"admin/index/index/menu_opt":      "操作",
	"admin/index/index/menu_close_all":   "关闭全部",
	"admin/index/index/menu_close_other": "关闭其它",
	"admin/index/index/menu_max":      "最多打开10个选项卡",
	"admin/index/index/menu_close_empty": "没有其他选项卡了",
	"admin/index/index/menu_cache_clear": "清除缓存",
	"admin/index/index/menu_lock":     "锁屏",
	"admin/index/index/menu_logout":   "退出登录",
	"admin/index/index/menu_index":    "前台首页",
	"admin/index/index/new_version":   "新版",
	"admin/index/cache_data":          "缓存数据",
	"maccms_copyright":                "© GOcms",
	"select_type":                     "选择分类",
	"select_status":                   "选择状态",
	"select_level":                    "选择推荐",
	"select_lock":                     "选择锁定",
	"select_sort":                     "排序方式",
	"select_pic":                      "图片",
	"wd":                              "搜索",
	"btn_search":                      "搜索",
	"id":                              "ID",
	"name":                            "名称",
	"hits":                            "点击",
	"hits_week":                       "周点击",
	"hits_month":                      "月点击",
	"hits_day":                        "日点击",
	"score":                           "评分",
	"level":                           "推荐",
	"browse":                          "浏览",
	"player":                          "播放器",
	"update_time":                     "更新时间",
	"opt":                             "操作",
	"add":                             "添加",
	"edit":                            "编辑",
	"del":                             "删除",
	"reviewed":                        "已审核",
	"reviewed_not":                    "未审核",
	"lock":                            "锁定",
	"unlock":                          "解锁",
	"close":                           "关闭",
	"open":                            "开启",
	"base_info":                       "基本信息",
	"other_info":                      "其他信息",
	"param":                           "参数",
	"status":                          "状态",
	"type":                            "分类",
	"pic_sync":                        "图片同步",
	"pic_empty":                       "无图",
	"pic_remote":                      "远程图片",
	"pic_sync_err":                    "同步失败",
	"make_page":                       "生成页面",
	"plot":                            "详情",
	"sort":                            "排序",
	"save":                            "保存",
	"cancel":                          "取消",
	"confirm":                         "确认",
	"tip":                             "提示",
	"success":                         "成功",
	"fail":                            "失败",
	"admin/vod/select_weekday":        "选择星期",
	"admin/vod/select_area":           "选择地区",
	"admin/vod/select_lang":           "选择语言",
	"admin/vod/select_server":         "选择服务器",
	"admin/vod/select_player":         "选择播放器",
	"admin/vod/select_downer":         "选择下载组",
	"admin/vod/select_isend":          "选择完结",
	"admin/vod/select_copyright":      "版权",
	"admin/vod/select_plot":           "剧情",
	"admin/vod/select_role":           "角色",
	"admin/vod/no_end":                "未完结",
	"admin/vod/is_end":                "已完结",
	"admin/vod/serialize":             "连载:",
	"admin/vod/player_empty":          "无播放器",
	"admin/vod/downer_empty":          "无下载组",
	"admin/vod/copyright":             "版权",
	"admin/vod/copyright_close":       "关闭版权",
	"admin/vod/copyright_open":        "开启版权",
	"admin/vod/role":                  "角色",
	"admin/vod/no":                    "无",
	"admin/vod/have":                  "有",
	"admin/recycle/back_list":         "返回列表",
	"admin/recycle/bin":               "回收站",
	"recycle_restore":                 "恢复",
	"recycle_purge":                   "彻底删除",
	"select_return":                   "选择返回",
	"update_repeat_cache":             "更新重复缓存",
	"del_auto_keep_min":               "删除重复(保留最小ID)",
	"del_auto_keep_max":               "删除重复(保留最大ID)",
	"type_name":                       "分类名称",
	"admin/ai_seo/status_col":         "SEO状态",
	"admin/ai_seo/status_optimized":   "已优化",
	"admin/ai_seo/status_fallback":    "降级",
	"admin/ai_seo/status_none":        "未处理",
	"admin/ai_seo/list_skip_label":    "跳过已优化",
	"admin/ai_seo/list_batch_btn":     "批量SEO",
	"admin/ai_seo/list_skip_hint":     "勾选后批量SEO将跳过已优化的内容",
	"admin/ai_seo/list_no_selection":  "请先选择内容",
	"admin/ai_seo/list_progress_title": "SEO优化进度",
	"admin/ai_seo/list_progress_line": "正在处理第 {0} 条，共 {1} 条",
	"admin/ai_seo/list_progress_sub":  "当前: {0}",
	"admin/ai_seo/list_summary":       "完成！成功 {0} 条，失败 {1} 条",
	"admin/ai_seo/list_summary_cancelled": "已取消",
	"admin/ai_seo/list_cancel":        "取消",
	"admin/ai_seo/list_row_btn":       "AI SEO",
	"admin/ai_seo/msg_generate_fail":  "生成失败",
	"admin/ai_seo/msg_request_fail":   "请求失败",
	"admin/assistant/title":           "AI 助手",
	"admin/assistant/open":            "打开助手",
	"admin/assistant/placeholder":     "输入问题...",
	"admin/assistant/send":            "发送",
	"admin/assistant/thinking":        "思考中...",
	"admin/assistant/close":           "关闭",
	"admin/assistant/role_user":       "你",
	"admin/assistant/role_assistant":  "助手",
	"admin/assistant/disabled_hint":   "AI 助手未启用",
	"menu/index":                      "首页",
	"menu/system":                     "系统",
	"menu/config":                     "系统配置",
	"menu/configseo":                  "SEO配置",
	"menu/configuser":                 "用户配置",
	"menu/configcomment":              "评论配置",
	"menu/configupload":               "上传配置",
	"menu/configcollect":              "采集配置",
	"menu/configinterface":            "接口配置",
	"menu/configapi":                  "API配置",
	"menu/configconnect":              "第三方登录",
	"menu/configpay":                  "支付配置",
	"menu/configweixin":               "微信配置",
	"menu/configemail":                "邮件配置",
	"menu/configsms":                  "短信配置",
	"menu/configplay":                 "播放器配置",
	"menu/configurl":                  "URL配置",
	"menu/configaiseo":                "AI SEO",
	"menu/configaisearch":             "AI搜索",
	"menu/configassistant":            "AI助手",
	"menu/timming":                    "定时任务",
	"menu/domain":                     "站群",
	"menu/base":                       "基础",
	"menu/type":                       "分类管理",
	"menu/topic":                      "专题管理",
	"menu/link":                       "友情链接",
	"menu/gbook":                      "留言管理",
	"menu/comment":                    "评论管理",
	"menu/chatroom":                   "聊天室",
	"menu/danmaku":                    "弹幕",
	"menu/task":                       "任务",
	"menu/task_log":                   "任务日志",
	"menu/images":                     "附件管理",
	"menu/vod":                        "视频",
	"menu/server":                     "服务器组",
	"menu/player":                     "播放器",
	"menu/downer":                     "下载器",
	"menu/vod_data":                   "视频数据",
	"menu/vod_add":                    "添加视频",
	"menu/vod_data_url_empty":         "无播放地址",
	"menu/vod_data_lock":              "已锁定",
	"menu/vod_data_audit":             "待审核",
	"menu/vod_data_points":            "收费视频",
	"menu/vod_data_plot":              "剧情",
	"menu/vod_batch":                  "批量操作",
	"menu/vod_repeat":                 "重复数据",
	"menu/actor":                      "演员管理",
	"menu/role":                       "角色管理",
	"menu/art":                        "文章",
	"menu/art_data":                   "文章数据",
	"menu/art_add":                    "添加文章",
	"menu/art_data_lock":              "已锁定",
	"menu/art_data_audit":             "待审核",
	"menu/art_batch":                  "批量操作",
	"menu/art_repeat":                 "重复数据",
	"manga":                           "漫画",
	"admin/manga/title":               "漫画数据",
	"menu/users":                      "用户",
	"menu/admin":                      "管理员",
	"menu/group":                      "会员组",
	"menu/user":                       "用户管理",
	"menu/card":                       "充值卡",
	"menu/order":                      "订单管理",
	"menu/ulog":                       "访问日志",
	"menu/plog":                       "积分日志",
	"menu/analytics":                  "统计分析",
	"menu/adminaudit":                 "操作审计",
	"menu/cash":                       "提现管理",
	"menu/templates":                  "模板",
	"menu/template":                   "模板管理",
	"menu/ads":                        "广告管理",
	"menu/wizard":                     "向导",
	"menu/make":                       "生成",
	"menu/make_opt":                   "生成选项",
	"menu/make_index":                 "生成首页",
	"menu/make_index_wap":             "生成WAP首页",
	"menu/make_map":                   "生成地图",
	"menu/cjs":                        "采集",
	"menu/union":                      "联盟采集",
	"menu/collect_timming":            "定时采集",
	"menu/collect":                    "自定义资源",
	"menu/cj":                         "自定义采集",
	"menu/db":                         "数据库",
	"menu/database":                   "数据库备份",
	"menu/database_sql":               "执行SQL",
	"menu/database_rep":               "数据替换",
	"menu/apps":                       "应用",
	"menu/addon":                      "插件管理",
	"menu/urlsend":                    "URL推送",
	"menu/safety_file":                "文件安全",
	"menu/safety_data":                "数据安全",
	"menu/live":                       "直播",
	"menu/live_list":                  "直播频道",
	"menu/live_category":              "直播分类",
	"menu/website":                    "网址",
	"menu/website_data":               "网址数据",
	"menu/website_add":                "添加网址",
	"menu/website_data_lock":          "已锁定",
	"menu/website_data_audit":         "待审核",
	"menu/website_batch":              "批量操作",
	"menu/website_repeat":             "重复数据",
	"menu/sign_milestone":             "签到里程碑",
	"menu/theme/config":               "主题配置",
	"multi_set":                       "批量设置",
	"duplicate_data":                  "重复数据",
	"admin/datareplace/title":         "数据替换",
	"admin/resourcehub/title":         "资源站",
	"admin/index/login/title":         "GOcms 后台登录",
	"admin/index/login/tip_welcome":   "欢迎使用GOcms后台",
	"admin/index/login/tip_sys":       "后台登录",
	"admin/index/login/btn_submit":    "立即登录",
	"admin/index/login/tip_declare":   "声明",
	"admin/index/login/tip_declare_txt": "GOcms 后台管理系统",
	"admin/index/login/verify_no":     "请输入用户名",
	"admin/index/login/verify_pass":   "请输入密码",
	"admin/index/login/verify_verify": "请输入验证码",
	"account":                         "用户名",
	"pass":                            "密码",
	"verify":                          "验证码",
	"wait_submit":                     "正在提交...",
	"admin/index/welcome/title":       "后台首页",
	"admin/index/title":               "后台管理",
}

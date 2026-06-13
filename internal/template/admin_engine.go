package template

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// AdminEngine 后台模板引擎（兼容 ThinkPHP 模板语法）
type AdminEngine struct {
	baseDir   string
	funcMap   template.FuncMap
	templates map[string]*template.Template
	mu        sync.RWMutex
}

// NewAdminEngine 创建后台模板引擎
func NewAdminEngine(baseDir string) *AdminEngine {
	return &AdminEngine{
		baseDir:   baseDir,
		funcMap:   adminFuncMap(),
		templates: make(map[string]*template.Template),
	}
}

// Load 加载所有后台模板
func (e *AdminEngine) Load() error {
	// 加载 public 目录下的基础模板
	publicDir := filepath.Join(e.baseDir, "public")
	publicFiles, _ := filepath.Glob(filepath.Join(publicDir, "*.html"))

	// 加载所有子目录
	subDirs, _ := os.ReadDir(e.baseDir)
	for _, sub := range subDirs {
		if !sub.IsDir() || sub.Name() == "public" {
			continue
		}
		subPath := filepath.Join(e.baseDir, sub.Name())
		htmlFiles, _ := filepath.Glob(filepath.Join(subPath, "*.html"))
		for _, file := range htmlFiles {
			name := sub.Name() + "/" + filepath.Base(file)
			if err := e.loadTemplate(name, file, publicFiles); err != nil {
				// 记录但不中断
				fmt.Printf("Warning: 模板 %s 加载失败: %v\n", name, err)
			}
		}
	}

	// 加载根目录的模板（如 index.html, login.html）
	rootFiles, _ := filepath.Glob(filepath.Join(e.baseDir, "*.html"))
	for _, file := range rootFiles {
		name := filepath.Base(file)
		if err := e.loadTemplate(name, file, publicFiles); err != nil {
			fmt.Printf("Warning: 模板 %s 加载失败: %v\n", name, err)
		}
	}

	return nil
}

// loadTemplate 加载单个模板并预处理 ThinkPHP 语法
func (e *AdminEngine) loadTemplate(name, filePath string, publicFiles []string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// 预处理：将 ThinkPHP 语法转换为 Go 模板语法
	processed := e.preprocess(string(content))

	tpl := template.New(name).Funcs(e.funcMap)

	// 注册 public 模板
	for _, pf := range publicFiles {
		pContent, err := os.ReadFile(pf)
		if err != nil {
			continue
		}
		pName := "public/" + filepath.Base(pf)
		pProcessed := e.preprocess(string(pContent))
		tpl.New(pName).Parse(pProcessed)
	}

	// 解析主模板
	tpl, err = tpl.Parse(processed)
	if err != nil {
		return fmt.Errorf("解析模板 %s: %w\nContent snippet: %s", name, err, processed[:min(200, len(processed))])
	}

	e.mu.Lock()
	e.templates[name] = tpl
	e.mu.Unlock()

	return nil
}

// preprocess 将 ThinkPHP 模板语法转换为 Go html/template 语法
func (e *AdminEngine) preprocess(content string) string {
	// 1. 替换路径常量
	content = strings.ReplaceAll(content, "__STATIC__", "/static")
	content = strings.ReplaceAll(content, "__ROOT__", "")

	// 2. 替换 {include file="path" /} → {{template "path" .}}
	re := regexp.MustCompile(`\{include\s+file="[^"]*?/?(public/[^"]+?)"\s*/?\s*\}`)
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		sub := re.FindStringSubmatch(match)
		if len(sub) > 1 {
			tplName := sub[1]
			// 只保留 public/xxx 部分
			parts := strings.Split(tplName, "/")
			for i, p := range parts {
				if p == "public" {
					tplName = strings.Join(parts[i:], "/")
					break
				}
			}
			return fmt.Sprintf(`{{template "%s" .}}`, tplName)
		}
		return match
	})

	// 3. 替换 {:lang('key')} 和 {:lang("key")} → {{lang "key"}}
	re = regexp.MustCompile(`\{:lang\(['"]([^'"]+)['"]\)\}`)
	content = re.ReplaceAllString(content, `{{lang "$1"}}`)

	// 4. 替换 {:url('path')} → {{url "path"}}
	re = regexp.MustCompile(`\{:url\(['"]([^'"]+)['"]\)\}`)
	content = re.ReplaceAllString(content, `{{url "$1"}}`)

	// 5. 替换 {:url('path', $param)} → {{url_param "path" .Param}}
	re = regexp.MustCompile(`\{:url\(['"]([^'"]+)['"],\s*\$param\)\}`)
	content = re.ReplaceAllString(content, `{{url_param "$1" .Param}}`)

	// 6. 替换 {:mac_url_vod_detail($vo)} → {{mac_url_vod_detail .}}
	re = regexp.MustCompile(`\{:mac_url_vod_detail\(\$vo\)\}`)
	content = re.ReplaceAllString(content, `{{mac_url_vod_detail .}}`)

	// 7. 替换 {:mac_url_topic_detail($vo)} → {{mac_url_topic_detail .}}
	re = regexp.MustCompile(`\{:mac_url_topic_detail\(\$vo\)\}`)
	content = re.ReplaceAllString(content, `{{mac_url_topic_detail .}}`)

	// 8. 替换 {:mac_url_art_detail($vo)} → {{mac_url_art_detail .}}
	re = regexp.MustCompile(`\{:mac_url_art_detail\(\$vo\)\}`)
	content = re.ReplaceAllString(content, `{{mac_url_art_detail .}}`)

	// 9. 替换 htmlspecialchars($var) → {{html_escape .Var}}
	// {:htmlspecialchars($vo.field)} → {{html_escape .Vo.Field}}

	// 10. 替换函数调用 {:func()} → {{func}}
	re = regexp.MustCompile(`\{:(\w+)\(\)\}`)
	content = re.ReplaceAllString(content, `{{$1}}`)

	// 11. 替换 {$var|mac_filter_xss|mac_restore_htmlfilter} → {{.Var}}
	// 简化处理：去掉 filter 链，直接输出
	re = regexp.MustCompile(`\{\$([\w.]+)\|[\w|]+\}`)
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		sub := regexp.MustCompile(`\{\$([\w.]+)\|`).FindStringSubmatch(match)
		if len(sub) > 1 {
			return goVarRef(sub[1])
		}
		return match
	})

	// 12. 替换 {$var.field} 和 {$var}
	re = regexp.MustCompile(`\{\$([\w.]+)\}`)
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		sub := re.FindStringSubmatch(match)
		if len(sub) > 1 {
			return goVarRef(sub[1])
		}
		return match
	})

	// 13. 替换 {volist name="list" id="vo"} → {{range .List}}
	// 和 {volist name="list" id="vo" key="key"}
	re = regexp.MustCompile(`\{volist\s+name="(\w+)"\s+id="(\w+)"(?:\s+key="(\w+)")?\s*\}`)
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		sub := re.FindStringSubmatch(match)
		name := sub[1]
		id := sub[2]
		goName := capitalize(name)
		if len(sub) > 3 && sub[3] != "" {
			key := sub[3]
			return fmt.Sprintf(`{{range $%s, $%s := .%s}}`, key, id, goName)
		}
		return fmt.Sprintf(`{{range $%s := .%s}}`, id, goName)
	})
	content = strings.ReplaceAll(content, "{/volist}", "{{end}}")

	// 14. 替换 {if condition="..."} → Go if
	// 处理各种条件格式
	content = convertIfConditions(content)
	content = strings.ReplaceAll(content, "{/if}", "{{end}}")
	content = strings.ReplaceAll(content, "{else /}", "{{else}}")
	content = strings.ReplaceAll(content, "{else}", "{{else}}")

	// 15. 替换 {$i} (循环索引) → 用 $key 或计数
	// 这在 volist 中自动处理

	// 16. 处理 ThinkPHP 的 eq/neq/gt/lt 比较
	// 在 convertIfConditions 中处理

	return content
}

// convertIfConditions 转换 ThinkPHP if 条件到 Go 语法
func convertIfConditions(content string) string {
	// 匹配 {if condition="$var eq 'value'"}
	re := regexp.MustCompile(`\{if\s+condition="([^"]+)"\s*\}`)
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		sub := re.FindStringSubmatch(match)
		if len(sub) > 1 {
			return convertCondition(sub[1])
		}
		return match
	})

	// 匹配 {elseif condition="$var eq 'value'"}
	re = regexp.MustCompile(`\{elseif\s+condition="([^"]+)"\s*/?\s*\}`)
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		sub := re.FindStringSubmatch(match)
		if len(sub) > 1 {
			return convertCondition("else " + sub[1])
		}
		return match
	})

	return content
}

// convertCondition 转换单个条件表达式
func convertCondition(cond string) string {
	cond = strings.TrimSpace(cond)

	// else 分支
	if strings.HasPrefix(cond, "else ") {
		inner := strings.TrimPrefix(cond, "else ")
		return convertSingleCondition(inner, true)
	}

	return convertSingleCondition(cond, false)
}

func convertSingleCondition(cond string, isElse bool) string {
	prefix := "{{if"
	if isElse {
		prefix = "{{else if"
	}

	// $var eq 'value'
	re := regexp.MustCompile(`^\$([\w.]+)\s+eq\s+'([^']*)'$`)
	if m := re.FindStringSubmatch(cond); len(m) > 0 {
		return fmt.Sprintf(`%s eq %s "%s"}}`, prefix, goVarRef(m[1]), m[2])
	}

	// $var eq "value"
	re = regexp.MustCompile(`^\$([\w.]+)\s+eq\s+"([^"]*)"$`)
	if m := re.FindStringSubmatch(cond); len(m) > 0 {
		return fmt.Sprintf(`%s eq %s "%s"}}`, prefix, goVarRef(m[1]), m[2])
	}

	// $var neq 'value'
	re = regexp.MustCompile(`^\$([\w.]+)\s+neq\s+'([^']*)'$`)
	if m := re.FindStringSubmatch(cond); len(m) > 0 {
		return fmt.Sprintf(`%s ne %s "%s"}}`, prefix, goVarRef(m[1]), m[2])
	}

	// $var eq value (without quotes, numeric)
	re = regexp.MustCompile(`^\$([\w.]+)\s+eq\s+(\d+)$`)
	if m := re.FindStringSubmatch(cond); len(m) > 0 {
		return fmt.Sprintf(`%s eq (int %s) %s}}`, prefix, goVarRef(m[1]), m[2])
	}

	// $var neq '' (empty check)
	re = regexp.MustCompile(`^\$([\w.]+)\s+neq\s+''$`)
	if m := re.FindStringSubmatch(cond); len(m) > 0 {
		return fmt.Sprintf(`%s ne %s ""}}`, prefix, goVarRef(m[1]))
	}

	// $var neq "" 
	re = regexp.MustCompile(`^\$([\w.]+)\s+neq\s+""$`)
	if m := re.FindStringSubmatch(cond); len(m) > 0 {
		return fmt.Sprintf(`%s ne %s ""}}`, prefix, goVarRef(m[1]))
	}

	// $var eq '' (is empty)
	re = regexp.MustCompile(`^\$([\w.]+)\s+eq\s+''$`)
	if m := re.FindStringSubmatch(cond); len(m) > 0 {
		return fmt.Sprintf(`%s eq %s ""}}`, prefix, goVarRef(m[1]))
	}

	// $var gt 0
	re = regexp.MustCompile(`^\$([\w.]+)\s+gt\s+(\d+)$`)
	if m := re.FindStringSubmatch(cond); len(m) > 0 {
		return fmt.Sprintf(`%s gt (int %s) %s}}`, prefix, goVarRef(m[1]), m[2])
	}

	// fallback: just wrap it
	return fmt.Sprintf(`%s .%s}}`, prefix, strings.TrimPrefix(cond, "$"))
}

// goVarRef 将 ThinkPHP 变量引用转换为 Go 模板变量引用
// "vo.vod_name" → "$vo.VodName" 或直接 ".Vo.VodName"
// 但 Go 模板中用 map 访问更方便：.vod_name
func goVarRef(varPath string) string {
	parts := strings.Split(varPath, ".")
	if len(parts) == 1 {
		// 单个变量如 "total" → ".Total"
		return "." + capitalize(parts[0])
	}
	// 多级如 "vo.vod_name" → "$vo.vod_name" (在 range 中)
	// 实际使用 map 访问
	return "$" + varPath
}

// capitalize 首字母大写
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Render 渲染模板
func (e *AdminEngine) Render(name string, data interface{}) (string, error) {
	e.mu.RLock()
	tpl, ok := e.templates[name]
	e.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("模板 %s 未找到", name)
	}

	var buf strings.Builder
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("渲染模板 %s: %w", name, err)
	}
	return buf.String(), nil
}

// adminFuncMap 后台模板函数
func adminFuncMap() template.FuncMap {
	return template.FuncMap{
		"lang": func(key string) string {
			// 简单的中文语言包
			return langMap[key]
		},
		"url": func(path string) string {
			return "/admin/" + path
		},
		"url_param": func(path string, param map[string]interface{}) string {
			base := "/admin/" + path
			// 构建查询参数
			var parts []string
			for k, v := range param {
				if k == "page" || k == "limit" || k == "wd" || k == "type" || k == "status" ||
					k == "level" || k == "lock" || k == "order" || k == "copyright" || k == "isend" {
					parts = append(parts, fmt.Sprintf("%s=%v", k, v))
				}
			}
			if len(parts) > 0 {
				return base + "?" + strings.Join(parts, "&")
			}
			return base
		},
		"mac_filter_xss": func(s string) string { return s },
		"mac_restore_htmlfilter": func(s string) string { return s },
		"mac_day": func(s string, style string) string { return s },
		"htmlspecialchars": func(s string) string { return template.HTMLEscapeString(s) },
		"html_escape": func(s string) string { return template.HTMLEscapeString(s) },
		"str_replace": func(search, replace, subject string) string {
			return strings.ReplaceAll(subject, search, replace)
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
		"seq": func(n int) []int {
			result := make([]int, n)
			for i := range result {
				result[i] = i + 1
			}
			return result
		},
	}
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
}

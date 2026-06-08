package template

import (
	"fmt"
	"regexp"
	"strings"
)

// TagParser 模板标签解析器
// 将原 macCMS 模板标签语法转换为 Go html/template 语法
type TagParser struct {
	funcs map[string]string
}

func NewTagParser() *TagParser {
	return &TagParser{
		funcs: make(map[string]string),
	}
}

// Parse 将 macCMS 模板标签转换为 Go template 语法
func (p *TagParser) Parse(content string) string {
	content = p.parseComments(content)
	content = p.removePHPTags(content)
	content = p.parseIncludes(content)
	content = p.parseFunctionCalls(content)
	content = p.parseVariables(content)
	content = p.parseIfTags(content)
	content = p.parseLoopTags(content)
	content = p.parseVolistTags(content)
	content = p.parsePageTags(content)
	content = p.parseConstants(content)
	return content
}

// 1. 注释标签 <!--*/ ... /*--> → 移除
func (p *TagParser) parseComments(content string) string {
	re := regexp.MustCompile(`<!--\*/[\s\S]*?/\*-->`)
	return re.ReplaceAllString(content, "")
}

// 2. PHP代码标签 {php}...{/php} → 移除
func (p *TagParser) removePHPTags(content string) string {
	re := regexp.MustCompile(`\{php\}[\s\S]*?\{/php\}`)
	return re.ReplaceAllString(content, "")
}

// 3. include 标签 {include file="header" /} → {{template "header.html" .}}
func (p *TagParser) parseIncludes(content string) string {
	re := regexp.MustCompile(`\{include\s+file="([^"]+)"\s*/?\}`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		sub := re.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		file := sub[1]
		if !strings.HasSuffix(file, ".html") {
			file = file + ".html"
		}
		return fmt.Sprintf(`{{template "%s" .}}`, file)
	})
}

// 4. 函数调用标签 {:func($var)} → {{func .Var}}
func (p *TagParser) parseFunctionCalls(content string) string {
	re := regexp.MustCompile(`\{:(\w+)\(([^)]*)\)\}`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		sub := re.FindStringSubmatch(match)
		if len(sub) < 3 {
			return match
		}
		funcName := sub[1]
		args := p.convertArgs(sub[2])
		return fmt.Sprintf("{{%s %s}}", funcName, args)
	})
}

// convertArgs 将 PHP 风格参数转换为 Go template 风格
func (p *TagParser) convertArgs(args string) string {
	args = strings.TrimSpace(args)
	if args == "" {
		return ""
	}
	var result []string
	parts := strings.Split(args, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "$") {
			varName := strings.TrimPrefix(part, "$")
			varName = strings.ReplaceAll(varName, ".", ".")
			result = append(result, "."+varName)
		} else {
			result = append(result, part)
		}
	}
	return strings.Join(result, " ")
}

// 5. 变量标签 {$vo.vod_name} → {{.vo.vod_name}}
func (p *TagParser) parseVariables(content string) string {
	// 移除 Think.* 变量
	re1 := regexp.MustCompile(`\{\$Think\.[^}]+\}`)
	content = re1.ReplaceAllString(content, "")

	// 普通变量
	re2 := regexp.MustCompile(`\{\$([a-zA-Z_][a-zA-Z0-9_.]*)\}`)
	return re2.ReplaceAllStringFunc(content, func(match string) string {
		sub := re2.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		return "{{." + sub[1] + "}}"
	})
}

// 6. 条件标签 {maccms:if condition="..."} → {{if ...}}
func (p *TagParser) parseIfTags(content string) string {
	re := regexp.MustCompile(`\{maccms:if\s+condition="([^"]+)"\s*\}`)
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		sub := re.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		condition := p.convertCondition(sub[1])
		return "{{if " + condition + "}}"
	})

	content = strings.ReplaceAll(content, "{maccms:else}", "{{else}}")
	content = strings.ReplaceAll(content, "{/maccms:if}", "{{end}}")

	re2 := regexp.MustCompile(`\{maccms:elseif\s+condition="([^"]+)"\s*\}`)
	content = re2.ReplaceAllStringFunc(content, func(match string) string {
		sub := re2.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		condition := p.convertCondition(sub[1])
		return "{{else if " + condition + "}}"
	})

	return content
}

// convertCondition 将 PHP 条件转换为 Go template 条件
func (p *TagParser) convertCondition(cond string) string {
	cond = strings.TrimSpace(cond)

	// $var → .var
	re := regexp.MustCompile(`\$([a-zA-Z_][a-zA-Z0-9_.]*)`)
	cond = re.ReplaceAllString(cond, ".$1")

	// 运算符映射
	cond = strings.ReplaceAll(cond, " egt ", " ge ")
	cond = strings.ReplaceAll(cond, " elt ", " le ")
	cond = strings.ReplaceAll(cond, " neq ", " ne ")
	cond = strings.ReplaceAll(cond, " lgt ", " lt ")
	cond = regexp.MustCompile(`\s*==\s*`).ReplaceAllString(cond, " eq ")
	cond = regexp.MustCompile(`\s*!=\s*`).ReplaceAllString(cond, " ne ")
	cond = regexp.MustCompile(`\s*>=\s*`).ReplaceAllString(cond, " ge ")
	cond = regexp.MustCompile(`\s*<=\s*`).ReplaceAllString(cond, " le ")
	cond = regexp.MustCompile(`\s*>\s*`).ReplaceAllString(cond, " gt ")
	cond = regexp.MustCompile(`\s*<\s*`).ReplaceAllString(cond, " lt ")

	return cond
}

// 7. 循环标签 {maccms:vod ...}...{/maccms:vod}
func (p *TagParser) parseLoopTags(content string) string {
	re := regexp.MustCompile(`\{maccms:(vod|art|manga|actor|role|type|topic|link|label|gbook|comment|website|user)\s+([^}]*)\}([\s\S]*?)\{/maccms:\1\}`)

	for re.MatchString(content) {
		content = re.ReplaceAllStringFunc(content, func(match string) string {
			sub := re.FindStringSubmatch(match)
			if len(sub) < 4 {
				return match
			}
			tagType := sub[1]
			attrs := sub[2]
			inner := sub[3]
			return p.convertLoopTag(tagType, attrs, inner)
		})
	}
	return content
}

func (p *TagParser) convertLoopTag(tagType, attrs, inner string) string {
	attrMap := p.parseAttrs(attrs)

	varName := "list"
	if v, ok := attrMap["name"]; ok {
		varName = v
	}

	dataSource := fmt.Sprintf(".%s_%s_data", tagType, varName)
	inner = p.Parse(inner) // 递归解析嵌套标签

	return fmt.Sprintf("{{range $vo := %s}}%s{{end}}", dataSource, inner)
}

func (p *TagParser) parseAttrs(attrs string) map[string]string {
	result := make(map[string]string)
	re := regexp.MustCompile(`(\w+)="([^"]*)"`)
	matches := re.FindAllStringSubmatch(attrs, -1)
	for _, m := range matches {
		if len(m) >= 3 {
			result[m[1]] = m[2]
		}
	}
	return result
}

// 8. volist 标签
func (p *TagParser) parseVolistTags(content string) string {
	re := regexp.MustCompile(`\{maccms:volist\s+([^}]*)\}([\s\S]*?)\{/maccms:volist\}`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		sub := re.FindStringSubmatch(match)
		if len(sub) < 3 {
			return match
		}
		attrs := p.parseAttrs(sub[1])
		inner := sub[2]

		name := attrs["name"]
		id := attrs["id"]
		if id == "" {
			id = "vo"
		}
		offset := attrs["offset"]

		inner = p.Parse(inner)

		var result strings.Builder
		if offset != "" {
			result.WriteString(fmt.Sprintf("{{range $i, $%s := .%s}}", id, name))
			result.WriteString(fmt.Sprintf("{{if ge $i %s}}", offset))
			result.WriteString(inner)
			result.WriteString("{{end}}{{end}}")
		} else {
			result.WriteString(fmt.Sprintf("{{range $%s := .%s}}", id, name))
			result.WriteString(inner)
			result.WriteString("{{end}}")
		}

		return result.String()
	})
}

// 9. 分页标签 {maccms:page} → {{maccms_page}}
func (p *TagParser) parsePageTags(content string) string {
	content = strings.ReplaceAll(content, "{maccms:page}", "{{maccms_page}}")
	content = strings.ReplaceAll(content, "{maccms:pagelist}", "{{maccms_pagelist}}")
	return content
}

// 10. 常量标签 {maccms:path} → {{.maccms.path}}
func (p *TagParser) parseConstants(content string) string {
	re := regexp.MustCompile(`\{maccms:(\w+)\}`)
	return re.ReplaceAllString(content, "{{.maccms.$1}}")
}

// RegisterFunc 注册自定义函数标签
func (p *TagParser) RegisterFunc(name, goFunc string) {
	p.funcs[name] = goFunc
}

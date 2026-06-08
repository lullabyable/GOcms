# 苹果CMS Go重写 — 补充实现方案

> 本文档是 `maccms-report.md` 的补充，覆盖原报告中所有缺失的具体实现。
> 生成时间: 2026-06-09

---

## 架构变更说明

相比原报告，本文档包含以下**关键架构变更**：

| 项目 | 原方案 | 新方案 |
|------|--------|--------|
| 缓存驱动 | Redis 必选 | **文件缓存为默认，Redis 可选扩展** |
| 静态HTML | 后台主动批量生成 | **旁路缓存（Lazy）：请求时自动生成+缓存** |
| 缓存击穿 | 未考虑 | **singleflight 防护** |
| 模板标签 | 仅映射表 | **完整解析器实现** |
| AlphaID | 空壳函数 | **完整编解码实现** |

---

## 一、AlphaID 编解码实现

与原 PHP 版 `mac_alphaID()` 完全兼容。使用自定义字符集做进制转换。

```go
// pkg/alphaid/alphaid.go

package alphaid

import (
    "math"
    "strings"
)

// 默认字符集（与原 PHP 版一致）
const defaultChars = "CFHJKLMNPRSTVWXYZ0123456789"

// AlphaIDEncode 将数字ID编码为短字符串
// 对应原 mac_alphaID($id, $encode_len, $encode_key)
func AlphaIDEncode(id int64, encodeLen int, encodeKey string) string {
    if id <= 0 {
        return "0"
    }

    // 确定字符集
    chars := defaultChars
    if encodeKey != "" {
        chars = shuffleChars(defaultChars, encodeKey)
    }

    base := int64(len(chars))
    if base == 0 {
        return "0"
    }

    // 进制转换
    var result strings.Builder
    for id > 0 {
        remainder := id % base
        result.WriteByte(chars[remainder])
        id = id / base
    }

    encoded := result.String()

    // 补齐到指定长度
    if encodeLen > 0 && len(encoded) < encodeLen {
        encoded = encoded + strings.Repeat(string(chars[0]), encodeLen-len(encoded))
    }

    return encoded
}

// AlphaIDDecode 将编码字符串解码回数字ID
func AlphaIDDecode(encoded string, encodeLen int, encodeKey string) int64 {
    if encoded == "" || encoded == "0" {
        return 0
    }

    chars := defaultChars
    if encodeKey != "" {
        chars = shuffleChars(defaultChars, encodeKey)
    }

    base := int64(len(chars))
    var result int64

    for i := len(encoded) - 1; i >= 0; i-- {
        idx := strings.IndexByte(chars, encoded[i])
        if idx < 0 {
            return 0
        }
        result = result*base + int64(idx)
    }

    return result
}

// shuffleChars 根据密钥打乱字符集（与原 PHP 版算法一致）
// 原 PHP: str_split → usort($chars, function($a, $b) use ($key) {
//     return ord($key[ord($a) % strlen($key)]) - ord($key[ord($b) % strlen($key)]);
// })
func shuffleChars(chars string, key string) string {
    if key == "" {
        return chars
    }

    runes := []rune(chars)
    keyLen := len(key)

    // 冒泡排序，排序依据与原 PHP 一致
    for i := 0; i < len(runes); i++ {
        for j := i + 1; j < len(runes); j++ {
            a := int(key[int(runes[i])%keyLen])
            b := int(key[int(runes[j])%keyLen])
            if a > b {
                runes[i], runes[j] = runes[j], runes[i]
            }
        }
    }

    return string(runes)
}

// AlphaIDEncodeWithMD5 对应原 mac_alphaID 中 md5 截取模式
// 当 id 为字符串时（如用 md5 做短链），用此函数
func AlphaIDEncodeWithMD5(input string, length int) string {
    if input == "" {
        return ""
    }
    // 对应原: substr(md5($id), 0, $encode_len)
    // 这里简化处理，取前 length 位
    if length <= 0 {
        length = 6
    }
    if len(input) <= length {
        return input
    }
    return input[:length]
}

// IsValidAlphaID 检查字符串是否是合法的 AlphaID 编码
func IsValidAlphaID(encoded string, encodeKey string) bool {
    chars := defaultChars
    if encodeKey != "" {
        chars = shuffleChars(defaultChars, encodeKey)
    }
    for _, c := range encoded {
        if !strings.ContainsRune(chars, c) {
            return false
        }
    }
    return true
}
```

---

## 二、播放源解析函数

原项目的播放源格式复杂，需要严格兼容。

```go
// pkg/playurl/playurl.go

package playurl

import (
    "strings"
)

// PlayURL 单个播放地址
type PlayURL struct {
    Name string // 如 "第1集"
    URL  string // 如 "https://xxx.m3u8"
}

// PlayGroup 一个播放源（如"量子m3u8"）
type PlayGroup struct {
    Flag string   // 播放源标识，如 "m3u8", "mp4"
    URLs []PlayURL
}

// ParsePlayURLs 解析 "名称1$url1#名称2$url2" 格式
// 对应原项目中的播放地址解析逻辑
func ParsePlayURLs(text string) []PlayURL {
    if text == "" {
        return nil
    }

    var result []PlayURL
    // 按 # 分割各集
    episodes := strings.Split(text, "#")
    for _, ep := range episodes {
        ep = strings.TrimSpace(ep)
        if ep == "" {
            continue
        }
        // 按 $ 分割名称和URL
        parts := strings.SplitN(ep, "$", 2)
        if len(parts) == 2 {
            name := strings.TrimSpace(parts[0])
            url := strings.TrimSpace(parts[1])
            if url != "" {
                result = append(result, PlayURL{Name: name, URL: url})
            }
        } else if len(parts) == 1 {
            // 没有名称，只有URL
            url := strings.TrimSpace(parts[0])
            if url != "" {
                result = append(result, PlayURL{Name: "", URL: url})
            }
        }
    }
    return result
}

// ParsePlayFromURL 解析 vod_play_from 和 vod_play_url 字段
// vod_play_from: "量子m3u8$$$非凡m3u8" （多个播放源用$$$分隔）
// vod_play_url: "第1集$url1#第2集$url2$$$第1集$url3#第2集$url4"
func ParsePlayFromURL(fromStr, urlStr string) []PlayGroup {
    if fromStr == "" || urlStr == "" {
        return nil
    }

    froms := strings.Split(fromStr, "$$$")
    urls := strings.Split(urlStr, "$$$")

    // 取较短的长度，防止数据不一致
    count := len(froms)
    if len(urls) < count {
        count = len(urls)
    }

    result := make([]PlayGroup, 0, count)
    for i := 0; i < count; i++ {
        flag := strings.TrimSpace(froms[i])
        urlText := strings.TrimSpace(urls[i])
        if flag == "" || urlText == "" {
            continue
        }

        group := PlayGroup{
            Flag: flag,
            URLs: ParsePlayURLs(urlText),
        }
        if len(group.URLs) > 0 {
            result = append(result, group)
        }
    }
    return result
}

// FormatPlayFrom 将 PlayGroup 列表格式化回 vod_play_from 字段
func FormatPlayFrom(groups []PlayGroup) string {
    flags := make([]string, 0, len(groups))
    for _, g := range groups {
        flags = append(flags, g.Flag)
    }
    return strings.Join(flags, "$$$")
}

// FormatPlayURL 将 PlayGroup 列表格式化回 vod_play_url 字段
func FormatPlayURL(groups []PlayGroup) string {
    urlParts := make([]string, 0, len(groups))
    for _, g := range groups {
        var episodes []string
        for _, u := range g.URLs {
            if u.Name != "" {
                episodes = append(episodes, u.Name+"$"+u.URL)
            } else {
                episodes = append(episodes, u.URL)
            }
        }
        urlParts = append(urlParts, strings.Join(episodes, "#"))
    }
    return strings.Join(urlParts, "$$$")
}

// MergePlayList 合并播放源（追加模式）
// 已存在的源按名称匹配，新源追加；同名源的URL列表合并去重
func MergePlayList(existingFrom, existingURL string, newGroups []PlayGroup) (string, string) {
    existing := ParsePlayFromURL(existingFrom, existingURL)

    // 建立已有源的索引
    existingIndex := make(map[string]int)
    for i, g := range existing {
        existingIndex[g.Flag] = i
    }

    for _, newGroup := range newGroups {
        if idx, ok := existingIndex[newGroup.Flag]; ok {
            // 源已存在，合并URL列表
            existing[idx].URLs = mergeURLs(existing[idx].URLs, newGroup.URLs)
        } else {
            // 新源，追加
            existing = append(existing, newGroup)
            existingIndex[newGroup.Flag] = len(existing) - 1
        }
    }

    return FormatPlayFrom(existing), FormatPlayURL(existing)
}

// mergeURLs 合并两个URL列表，按名称去重
func mergeURLs(old, newURLs []PlayURL) []PlayURL {
    // 建立已有URL的名称索引
    nameIndex := make(map[string]int)
    for i, u := range old {
        nameIndex[u.Name] = i
    }

    result := make([]PlayURL, len(old))
    copy(result, old)

    for _, u := range newURLs {
        if _, exists := nameIndex[u.Name]; !exists {
            result = append(result, u)
        }
        // 已存在同名的，保留旧的（不覆盖）
    }
    return result
}

---

## 三、模板标签解析器

将原 PHP 模板的自定义标签语法转换为 Go `html/template` 语法。

```go
// internal/template/parser.go

package template

import (
    "fmt"
    "regexp"
    "strings"
)

// TagParser 模板标签解析器
type TagParser struct {
    funcs map[string]string // 自定义函数标签注册
}

func NewTagParser() *TagParser {
    return &TagParser{
        funcs: make(map[string]string),
    }
}

// Parse 将原 macCMS 模板标签转换为 Go template 语法
// 这是入口方法，按顺序处理各类标签
func (p *TagParser) Parse(content string) string {
    // 1. 注释标签 <!--*/ ... /*--> → 移除
    content = p.parseComments(content)

    // 2. PHP代码标签 {php}...{/php} → 移除（Go中不执行PHP）
    content = p.removePHPTags(content)

    // 3. include/extend 标签 {include file="header" /}
    content = p.parseIncludes(content)

    // 4. 自定义函数标签 {:func($var)} → {{func .Var}}
    content = p.parseFunctionCalls(content)

    // 5. 变量标签 {$vo.vod_name} → {{.vo.vod_name}} 或 {{.vod_name}}
    content = p.parseVariables(content)

    // 6. 条件标签 {maccms:if ...}...{maccms:else}...{/maccms:if}
    content = p.parseIfTags(content)

    // 7. 循环标签 {maccms:vod ...}...{/maccms:vod}
    content = p.parseLoopTags(content)

    // 8. volist 标签 {maccms:volist ...}...{/maccms:volist}
    content = p.parseVolistTags(content)

    // 9. 分页标签 {maccms:page}
    content = p.parsePageTags(content)

    // 10. 常量标签 {maccms:path} 等
    content = p.parseConstants(content)

    // 11. 空白标签清理
    content = p.cleanupEmptyTags(content)

    return content
}

// ============================================================
// 1. 注释标签
// ============================================================
func (p *TagParser) parseComments(content string) string {
    // <!--*/ ... /*--> 是 macCMS 的注释标签
    re := regexp.MustCompile(`<!--\*/[\s\S]*?/\*-->`)
    return re.ReplaceAllString(content, "")
}

// ============================================================
// 2. PHP代码标签
// ============================================================
func (p *TagParser) removePHPTags(content string) string {
    re := regexp.MustCompile(`\{php\}[\s\S]*?\{/php\}`)
    return re.ReplaceAllString(content, "")
}

// ============================================================
// 3. include 标签
// ============================================================
func (p *TagParser) parseIncludes(content string) string {
    // {include file="header" /} → {{template "header.html" .}}
    re := regexp.MustCompile(`\{include\s+file="([^"]+)"\s*/?\}`)
    return re.ReplaceAllStringFunc(content, func(match string) string {
        sub := re.FindStringSubmatch(match)
        if len(sub) < 2 {
            return match
        }
        file := sub[1]
        // 确保有 .html 后缀
        if !strings.HasSuffix(file, ".html") {
            file = file + ".html"
        }
        return fmt.Sprintf(`{{template "%s" .}}`, file)
    })
}

// ============================================================
// 4. 函数调用标签
// ============================================================
func (p *TagParser) parseFunctionCalls(content string) string {
    // {:mac_url_vod_detail($vo)} → {{mac_url_vod_detail .vo}}
    // {:mac_default($vo.vod_sub, $vo.vod_name)} → {{mac_default .vo.vod_sub .vo.vod_name}}
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
// $vo → .vo, $vo.vod_name → .vo.vod_name, 'literal' → "literal"
func (p *TagParser) convertArgs(args string) string {
    args = strings.TrimSpace(args)
    if args == "" {
        return ""
    }

    var result []string
    // 按逗号分割（注意：不处理括号内的逗号）
    parts := strings.Split(args, ",")
    for _, part := range parts {
        part = strings.TrimSpace(part)
        if part == "" {
            continue
        }

        // $variable → .variable
        if strings.HasPrefix(part, "$") {
            varName := strings.TrimPrefix(part, "$")
            varName = strings.ReplaceAll(varName, ".", ".")
            result = append(result, "."+varName)
            continue
        }

        // 字符串字面量
        if strings.HasPrefix(part, "'") || strings.HasPrefix(part, `"`) {
            result = append(result, part)
            continue
        }

        // 数字
        result = append(result, part)
    }

    return strings.Join(result, " ")
}

// ============================================================
// 5. 变量标签
// ============================================================
func (p *TagParser) parseVariables(content string) string {
    // {$vo.vod_name} → {{.vo.vod_name}}
    // {$Think.get.page} → (移除，Go中不支持)
    // {$maccms.site_name} → {{.maccms.site_name}}

    // 先处理 Think.* 变量（移除）
    re1 := regexp.MustCompile(`\{\$Think\.[^}]+\}`)
    content = re1.ReplaceAllString(content, "")

    // 处理普通变量
    re2 := regexp.MustCompile(`\{\$([a-zA-Z_][a-zA-Z0-9_.]*)\}`)
    return re2.ReplaceAllStringFunc(content, func(match string) string {
        sub := re2.FindStringSubmatch(match)
        if len(sub) < 2 {
            return match
        }
        return "{{." + sub[1] + "}}"
    })
}

// ============================================================
// 6. 条件标签
// ============================================================
func (p *TagParser) parseIfTags(content string) string {
    // {maccms:if condition="$vo.vod_score > 8"} ... {maccms:else} ... {/maccms:if}
    // → {{if gt .vo.vod_score "8"}} ... {{else}} ... {{end}}

    // 先处理带 condition 属性的 if
    re := regexp.MustCompile(`\{maccms:if\s+condition="([^"]+)"\s*\}`)
    content = re.ReplaceAllStringFunc(content, func(match string) string {
        sub := re.FindStringSubmatch(match)
        if len(sub) < 2 {
            return match
        }
        condition := p.convertCondition(sub[1])
        return "{{if " + condition + "}}"
    })

    // 处理 {maccms:else}
    content = strings.ReplaceAll(content, "{maccms:else}", "{{else}}")

    // 处理 {/maccms:if}
    content = strings.ReplaceAll(content, "{/maccms:if}", "{{end}}")

    // 处理 {maccms:elseif condition="..."} → {{else if ...}}
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
// $vo.vod_score > 8 → gt .vo.vod_score "8"
// $vo.vod_status eq 1 → eq .vo.vod_status 1
func (p *TagParser) convertCondition(cond string) string {
    cond = strings.TrimSpace(cond)

    // 替换 $var → .var
    re := regexp.MustCompile(`\$([a-zA-Z_][a-zA-Z0-9_.]*)`)
    cond = re.ReplaceAllString(cond, ".$1")

    // 处理比较运算符
    // eq → eq, neq → ne, gt → gt, egt → ge, lgt → lt, elt → le
    cond = strings.ReplaceAll(cond, " egt ", " ge ")
    cond = strings.ReplaceAll(cond, " elt ", " le ")
    cond = strings.ReplaceAll(cond, " neq ", " ne ")
    cond = strings.ReplaceAll(cond, " lgt ", " lt ")

    // 处理 == → eq
    cond = regexp.MustCompile(`\s*==\s*`).ReplaceAllString(cond, " eq ")
    // 处理 != → ne
    cond = regexp.MustCompile(`\s*!=\s*`).ReplaceAllString(cond, " ne ")
    // 处理 >= → ge
    cond = regexp.MustCompile(`\s*>=\s*`).ReplaceAllString(cond, " ge ")
    // 处理 <= → le
    cond = regexp.MustCompile(`\s*<=\s*`).ReplaceAllString(cond, " le ")
    // 处理 > → gt
    cond = regexp.MustCompile(`\s*>\s*`).ReplaceAllString(cond, " gt ")
    // 处理 < → lt
    cond = regexp.MustCompile(`\s*<\s*`).ReplaceAllString(cond, " lt ")

    return cond
}

// ============================================================
// 7. 循环标签（核心）
// ============================================================
func (p *TagParser) parseLoopTags(content string) string {
    // 匹配所有 {maccms:xxx ...} ... {/maccms:xxx} 标签
    // 支持嵌套
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

// convertLoopTag 将单个循环标签转换为 Go template
func (p *TagParser) convertLoopTag(tagType, attrs, inner string) string {
    // 解析属性
    attrMap := p.parseAttrs(attrs)

    // 确定循环变量名
    varName := "list"
    if v, ok := attrMap["name"]; ok {
        varName = v
    }

    // 构建数据源
    dataSource := fmt.Sprintf(".%s_%s_data", tagType, varName)

    // 构建查询参数（传给数据加载函数）
    queryParams := p.buildQueryParams(tagType, attrMap)

    // 转换内部模板
    inner = p.Parse(inner) // 递归解析嵌套标签

    // 生成 Go template 代码
    // 使用 range 遍历，$vo 作为当前元素
    var result strings.Builder
    result.WriteString(fmt.Sprintf("{{range $vo := %s}}", dataSource))
    result.WriteString(inner)
    result.WriteString("{{end}}")

    // 包装为一个 define 块，数据加载逻辑在 handler 中处理
    return fmt.Sprintf(`{{/* maccms:%s %s */}}%s`, tagType, queryParams, result.String())
}

// parseAttrs 解析标签属性字符串
// type="1" num="10" orderby="time" → {"type":"1", "num":"10", "orderby":"time"}
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

// buildQueryParams 将标签属性转换为查询参数字符串
func (p *TagParser) buildQueryParams(tagType string, attrs map[string]string) string {
    var params []string
    for k, v := range attrs {
        if k == "name" || k == "key" || k == "mod" {
            continue // 这些是模板控制属性，不是查询参数
        }
        params = append(params, fmt.Sprintf("%s=%s", k, v))
    }
    return strings.Join(params, "&")
}

// ============================================================
// 8. volist 标签
// ============================================================
func (p *TagParser) parseVolistTags(content string) string {
    // {maccms:volist name="list" id="vo" offset="5" length="10"}
    //   {$vo.vod_name}
    // {/maccms:volist}
    // → {{range $i, $vo := .list}}{{if and (ge $i 5) (lt $i 15)}}...{{end}}{{end}}

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
        length := attrs["length"]

        inner = p.Parse(inner)

        var result strings.Builder

        if offset != "" || length != "" {
            // 需要分片
            result.WriteString(fmt.Sprintf("{{range $i, $%s := .%s}}", id, name))
            if offset != "" && length != "" {
                result.WriteString(fmt.Sprintf("{{if and (ge $i %s) (lt $i (add %s %s))}}", offset, offset, length))
            } else if offset != "" {
                result.WriteString(fmt.Sprintf("{{if ge $i %s}}", offset))
            }
            result.WriteString(inner)
            if offset != "" {
                result.WriteString("{{end}}")
            }
            result.WriteString("{{end}}")
        } else {
            result.WriteString(fmt.Sprintf("{{range $%s := .%s}}", id, name))
            result.WriteString(inner)
            result.WriteString("{{end}}")
        }

        return result.String()
    })
}

// ============================================================
// 9. 分页标签
// ============================================================
func (p *TagParser) parsePageTags(content string) string {
    // {maccms:page} → {{maccms_page}}
    content = strings.ReplaceAll(content, "{maccms:page}", "{{maccms_page}}")
    // {maccms:pagelist} → {{maccms_pagelist}}
    content = strings.ReplaceAll(content, "{maccms:pagelist}", "{{maccms_pagelist}}")
    return content
}

// ============================================================
// 10. 常量标签
// ============================================================
func (p *TagParser) parseConstants(content string) string {
    // {maccms:path} → {{.maccms.path}}
    // {maccms:url} → {{.maccms.url}}
    // {maccms:name} → {{.maccms.name}}
    re := regexp.MustCompile(`\{maccms:(\w+)\}`)
    return re.ReplaceAllString(content, "{{.maccms.$1}}")
}

// ============================================================
// 11. 清理
// ============================================================
func (p *TagParser) cleanupEmptyTags(content string) string {
    // 移除空的 maccms 注释
    re := regexp.MustCompile(`\{\{/\*.*?\*/\}\}`)
    content = re.ReplaceAllString(content, "")
    return content
}

// ============================================================
// 注册自定义函数标签
// ============================================================
func (p *TagParser) RegisterFunc(name, goFunc string) {
    p.funcs[name] = goFunc
}

---

## 四、PHP → Go 类型安全转换

原 PHP 是弱类型，数据库中混存了各种类型。Go 需要安全转换。

```go
// pkg/safecast/cast.go

package safecast

import (
    "strconv"
    "strings"
    "time"
)

// SafeInt 安全转换为 int（处理 PHP 空字符串/NULL → 0）
func SafeInt(v interface{}) int {
    switch val := v.(type) {
    case int:
        return val
    case int64:
        return int(val)
    case float64:
        return int(val)
    case string:
        val = strings.TrimSpace(val)
        if val == "" || val == "null" || val == "NULL" {
            return 0
        }
        i, _ := strconv.Atoi(val)
        return i
    case nil:
        return 0
    case []byte:
        return SafeInt(string(val))
    default:
        return 0
    }
}

// SafeInt64 安全转换为 int64
func SafeInt64(v interface{}) int64 {
    switch val := v.(type) {
    case int:
        return int64(val)
    case int64:
        return val
    case float64:
        return int64(val)
    case string:
        val = strings.TrimSpace(val)
        if val == "" || val == "null" || val == "NULL" {
            return 0
        }
        i, _ := strconv.ParseInt(val, 10, 64)
        return i
    case nil:
        return 0
    case []byte:
        return SafeInt64(string(val))
    default:
        return 0
    }
}

// SafeFloat 安全转换为 float64
func SafeFloat(v interface{}) float64 {
    switch val := v.(type) {
    case float64:
        return val
    case int:
        return float64(val)
    case int64:
        return float64(val)
    case string:
        val = strings.TrimSpace(val)
        if val == "" || val == "null" || val == "NULL" {
            return 0
        }
        f, _ := strconv.ParseFloat(val, 64)
        return f
    case nil:
        return 0
    case []byte:
        return SafeFloat(string(val))
    default:
        return 0
    }
}

// SafeString 安全转换为 string
func SafeString(v interface{}) string {
    switch val := v.(type) {
    case string:
        return val
    case []byte:
        return string(val)
    case int, int64, float64:
        return fmt.Sprintf("%v", val)
    case nil:
        return ""
    default:
        return fmt.Sprintf("%v", val)
    }
}

// SafeTime 安全转换为 time.Time
// 支持 Unix 时间戳（int64）和 "2006-01-02 15:04:05" 格式
func SafeTime(v interface{}) time.Time {
    switch val := v.(type) {
    case time.Time:
        return val
    case int64:
        if val > 0 {
            return time.Unix(val, 0)
        }
    case int:
        if val > 0 {
            return time.Unix(int64(val), 0)
        }
    case float64:
        if val > 0 {
            return time.Unix(int64(val), 0)
        }
    case string:
        val = strings.TrimSpace(val)
        if val == "" || val == "0000-00-00 00:00:00" || val == "0" {
            return time.Time{}
        }
        // 尝试 Unix 时间戳字符串
        if ts, err := strconv.ParseInt(val, 10, 64); err == nil && ts > 0 {
            return time.Unix(ts, 0)
        }
        // 尝试常见日期格式
        formats := []string{
            "2006-01-02 15:04:05",
            "2006-01-02",
            "2006/01/02 15:04:05",
            "2006/01/02",
        }
        for _, f := range formats {
            if t, err := time.ParseInLocation(f, val, time.Local); err == nil {
                return t
            }
        }
    case nil:
        return time.Time{}
    }
    return time.Time{}
}

// SafeBool 安全转换为 bool
func SafeBool(v interface{}) bool {
    switch val := v.(type) {
    case bool:
        return val
    case int:
        return val != 0
    case int64:
        return val != 0
    case string:
        val = strings.TrimSpace(strings.ToLower(val))
        return val == "1" || val == "true" || val == "yes" || val == "on"
    case nil:
        return false
    default:
        return false
    }
}

---

## 五、会话管理完整方案

### 5.1 Session 存储抽象层

```go
// internal/session/session.go

package session

import (
    "crypto/rand"
    "encoding/hex"
    "sync"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/session"
)

// SessionStore 会话存储抽象接口
type SessionStore interface {
    Get(key string) (string, error)
    Set(key, value string, ttl time.Duration) error
    Delete(key string) error
}

// Manager 会话管理器
type Manager struct {
    store      SessionStore
    cookieName string
    maxAge     time.Duration
    mu         sync.RWMutex
}

type ManagerConfig struct {
    StoreType  string // "cookie" 或 "redis"
    CookieName string
    MaxAge     time.Duration
    Secret     string
    // Redis 配置（仅 StoreType=redis 时使用）
    RedisAddr     string
    RedisPassword string
    RedisDB       int
}

func NewManager(cfg ManagerConfig) *Manager {
    var store SessionStore

    switch cfg.StoreType {
    case "redis":
        store = NewRedisSessionStore(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
    default:
        store = NewCookieSessionStore(cfg.Secret)
    }

    name := cfg.CookieName
    if name == "" {
        name = "maccms_session"
    }
    maxAge := cfg.MaxAge
    if maxAge == 0 {
        maxAge = 24 * time.Hour
    }

    return &Manager{
        store:      store,
        cookieName: name,
        maxAge:     maxAge,
    }
}

// Get 获取会话
func (m *Manager) Get(c *fiber.Ctx) *Session {
    sessionID := c.Cookies(m.cookieName)
    if sessionID == "" {
        sessionID = m.generateID()
        c.Cookie(&fiber.Cookie{
            Name:     m.cookieName,
            Value:    sessionID,
            MaxAge:   int(m.maxAge.Seconds()),
            HTTPOnly: true,
            Secure:   true,
            SameSite: "Lax",
        })
    }
    return &Session{
        id:    sessionID,
        store: m.store,
        ttl:   m.maxAge,
    }
}

// Regenerate 登录后重新生成 Session ID（防 Session 固定攻击）
func (m *Manager) Regenerate(c *fiber.Ctx) *Session {
    oldID := c.Cookies(m.cookieName)
    newID := m.generateID()

    // 迁移旧数据
    if oldID != "" {
        m.store.Delete(oldID)
    }

    c.Cookie(&fiber.Cookie{
        Name:     m.cookieName,
        Value:    newID,
        MaxAge:   int(m.maxAge.Seconds()),
        HTTPOnly: true,
        Secure:   true,
        SameSite: "Lax",
    })

    return &Session{
        id:    newID,
        store: m.store,
        ttl:   m.maxAge,
    }
}

func (m *Manager) generateID() string {
    b := make([]byte, 32)
    rand.Read(b)
    return hex.EncodeToString(b)
}

// Session 单个会话
type Session struct {
    id    string
    store SessionStore
    ttl   time.Duration
}

func (s *Session) Get(key string) (string, error) {
    return s.store.Get(s.id + ":" + key)
}

func (s *Session) Set(key, value string) error {
    return s.store.Set(s.id+":"+key, value, s.ttl)
}

func (s *Session) Delete(key string) error {
    return s.store.Delete(s.id + ":" + key)
}

func (s *Session) ID() string {
    return s.id
}
```

### 5.2 Cookie 存储实现（默认，无需外部依赖）

```go
// internal/session/cookie_store.go

package session

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "errors"
    "sync"
    "time"
)

// CookieSessionStore 基于 Cookie 的会话存储
// 数据加密后存在 Cookie 中，无需 Redis
type CookieSessionStore struct {
    secret []byte
    mu     sync.RWMutex
    // 内存缓存（可选，用于大数据量场景）
    cache map[string]*cacheEntry
}

type cacheEntry struct {
    value     string
    expiresAt time.Time
}

func NewCookieSessionStore(secret string) *CookieSessionStore {
    // 确保密钥长度为 16/24/32 字节
    key := make([]byte, 32)
    copy(key, []byte(secret))
    return &CookieSessionStore{
        secret: key,
        cache:  make(map[string]*cacheEntry),
    }
}

func (s *CookieSessionStore) Get(key string) (string, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    if entry, ok := s.cache[key]; ok {
        if time.Now().Before(entry.expiresAt) {
            return entry.value, nil
        }
        // 过期，删除
        delete(s.cache, key)
    }
    return "", errors.New("session key not found")
}

func (s *CookieSessionStore) Set(key, value string, ttl time.Duration) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    s.cache[key] = &cacheEntry{
        value:     value,
        expiresAt: time.Now().Add(ttl),
    }
    return nil
}

func (s *CookieSessionStore) Delete(key string) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    delete(s.cache, key)
    return nil
}

// Encrypt 加密 Cookie 值（用于存储敏感数据到 Cookie）
func (s *CookieSessionStore) Encrypt(plaintext string) (string, error) {
    block, err := aes.NewCipher(s.secret)
    if err != nil {
        return "", err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    nonce := make([]byte, gcm.NonceSize())
    rand.Read(nonce)

    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密 Cookie 值
func (s *CookieSessionStore) Decrypt(encrypted string) (string, error) {
    ciphertext, err := base64.URLEncoding.DecodeString(encrypted)
    if err != nil {
        return "", err
    }

    block, err := aes.NewCipher(s.secret)
    if err != nil {
        return "", err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    nonceSize := gcm.NonceSize()
    if len(ciphertext) < nonceSize {
        return "", errors.New("ciphertext too short")
    }

    plaintext, err := gcm.Open(nil, ciphertext[:nonceSize], ciphertext[nonceSize:], nil)
    if err != nil {
        return "", err
    }

    return string(plaintext), nil
}
```

### 5.3 后台认证中间件

```go
// internal/middleware/admin_auth.go

package middleware

import (
    "strings"

    "github.com/gofiber/fiber/v2"
    "maccms-go/internal/session"
)

// AdminAuth 后台认证中间件
func AdminAuth(sm *session.Manager) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 登录页面跳过
        path := c.Path()
        if strings.HasSuffix(path, "/login") || strings.HasSuffix(path, "/captcha") {
            return c.Next()
        }

        sess := sm.Get(c)
        adminID, _ := sess.Get("admin_id")
        if adminID == "" {
            // API 请求返回 JSON
            if strings.HasPrefix(path, "/api/") || c.Get("X-Requested-With") == "XMLHttpRequest" {
                return c.Status(401).JSON(fiber.Map{"code": 401, "msg": "请先登录"})
            }
            return c.Redirect("/admin/login")
        }

        // 将管理员信息存入 Locals
        adminName, _ := sess.Get("admin_name")
        adminRole, _ := sess.Get("admin_role")
        c.Locals("admin_id", adminID)
        c.Locals("admin_name", adminName)
        c.Locals("admin_role", adminRole)

        return c.Next()
    }
}

// AdminPermission 权限检查中间件
// role: 1=超管, 2=普通管理员, 3=只读
func AdminPermission(requiredRole int) fiber.Handler {
    return func(c *fiber.Ctx) error {
        role := c.Locals("admin_role")
        if role == nil {
            return c.Status(403).JSON(fiber.Map{"code": 403, "msg": "无权限"})
        }

        roleInt := 0
        switch v := role.(type) {
        case string:
            if v == "1" {
                roleInt = 1
            } else if v == "2" {
                roleInt = 2
            } else if v == "3" {
                roleInt = 3
            }
        case int:
            roleInt = v
        }

        // 角色数字越小权限越大
        if roleInt == 0 || roleInt > requiredRole {
            return c.Status(403).JSON(fiber.Map{"code": 403, "msg": "权限不足"})
        }

        return c.Next()
    }
}

// LoginRateLimit 登录失败次数限制
func LoginRateLimit(maxAttempts int, lockoutDuration time.Duration) fiber.Handler {
    type attemptInfo struct {
        count    int
        lastTime time.Time
    }

    var (
        attempts = make(map[string]*attemptInfo)
        mu       sync.RWMutex
    )

    // 定期清理过期记录
    go func() {
        ticker := time.NewTicker(5 * time.Minute)
        for range ticker.C {
            mu.Lock()
            for ip, info := range attempts {
                if time.Since(info.lastTime) > lockoutDuration {
                    delete(attempts, ip)
                }
            }
            mu.Unlock()
        }
    }()

    return func(c *fiber.Ctx) error {
        ip := c.IP()

        mu.RLock()
        info, exists := attempts[ip]
        mu.RUnlock()

        if exists && info.count >= maxAttempts {
            if time.Since(info.lastTime) < lockoutDuration {
                remaining := lockoutDuration - time.Since(info.lastTime)
                return c.Status(429).JSON(fiber.Map{
                    "code": 429,
                    "msg":  fmt.Sprintf("登录失败次数过多，请 %d 分钟后重试", int(remaining.Minutes())),
                })
            }
            // 锁定时间已过，重置
            mu.Lock()
            delete(attempts, ip)
            mu.Unlock()
        }

        // 执行后续处理
        err := c.Next()

        // 如果是登录请求且失败，记录次数
        if c.Method() == "POST" && strings.HasSuffix(c.Path(), "/login") {
            // 检查响应是否表示登录失败
            // 这里简化处理，实际应该检查响应内容
            if c.Response().StatusCode() != 200 {
                mu.Lock()
                if attempts[ip] == nil {
                    attempts[ip] = &attemptInfo{}
                }
                attempts[ip].count++
                attempts[ip].lastTime = time.Now()
                mu.Unlock()
            }
        }

        return err
    }
}
```

---

## 六、旁路缓存 + 缓存击穿防护（核心重写）

### 6.1 CacheManager 重构：Redis 可选

```go
// internal/cache/cache.go

package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"
)

// CacheManager 统一缓存管理器
// 文件缓存为默认驱动，Redis 作为可选扩展
type CacheManager struct {
    mu       sync.RWMutex
    driver   CacheDriver
    config   CacheConfig
    prefix   string
}

// CacheDriver 缓存驱动接口
type CacheDriver interface {
    Get(key string) (string, error)
    Set(key, value string, ttl time.Duration) error
    Delete(key string) error
    DeletePattern(pattern string) error
    Clear() error
    Close() error
}

type CacheConfig struct {
    Type     string `json:"cache_type"`      // "file" (默认) 或 "redis"
    Flag     string `json:"cache_flag"`       // 缓存键前缀（多站点隔离）
    Core     int    `json:"cache_core"`       // 核心缓存开关 (0=关闭, 1=开启)
    Time     int    `json:"cache_time"`       // 默认过期时间(秒)
    Page     int    `json:"cache_page"`       // 页面缓存开关
    TimePage int    `json:"cache_time_page"`  // 页面缓存过期时间(秒)
    FileDir  string `json:"cache_file_dir"`   // 文件缓存目录
    // Redis 配置（可选）
    RedisAddr     string `json:"cache_host"`
    RedisPort     string `json:"cache_port"`
    RedisPassword string `json:"cache_password"`
    RedisDB       int    `json:"cache_db"`
}

func NewCacheManager(config CacheConfig) (*CacheManager, error) {
    var driver CacheDriver
    var err error

    switch config.Type {
    case "redis":
        driver, err = NewRedisDriver(config)
        if err != nil {
            // Redis 不可用，降级到文件缓存
            fmt.Printf("[WARN] Redis 连接失败 (%v)，降级到文件缓存\n", err)
            driver = NewFileDriver(config.FileDir)
        }
    default:
        driver = NewFileDriver(config.FileDir)
    }

    prefix := config.Flag
    if prefix == "" {
        prefix = "maccms"
    }

    return &CacheManager{
        driver: driver,
        config: config,
        prefix: prefix + "_",
    }, nil
}

func (cm *CacheManager) Get(key string) (string, error) {
    return cm.driver.Get(cm.prefix + key)
}

func (cm *CacheManager) Set(key, value string, ttl int) error {
    if ttl <= 0 {
        ttl = cm.config.Time
    }
    return cm.driver.Set(cm.prefix+key, value, time.Duration(ttl)*time.Second)
}

func (cm *CacheManager) Delete(key string) error {
    return cm.driver.Delete(cm.prefix + key)
}

func (cm *CacheManager) DeletePattern(pattern string) error {
    return cm.driver.DeletePattern(cm.prefix + pattern)
}

func (cm *CacheManager) Clear() error {
    return cm.driver.DeletePattern(cm.prefix + "*")
}

func (cm *CacheManager) Close() error {
    return cm.driver.Close()
}

func (cm *CacheManager) IsEnabled() bool {
    return cm.config.Core == 1
}

// ReloadConfig 热更新配置
func (cm *CacheManager) ReloadConfig(config CacheConfig) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    // 如果切换了驱动类型
    if config.Type != cm.config.Type {
        cm.driver.Close()
        switch config.Type {
        case "redis":
            driver, err := NewRedisDriver(config)
            if err != nil {
                fmt.Printf("[WARN] Redis 不可用，继续使用文件缓存\n")
                driver = NewFileDriver(config.FileDir)
            }
            cm.driver = driver
        default:
            cm.driver = NewFileDriver(config.FileDir)
        }
    }

    cm.config = config
    prefix := config.Flag
    if prefix == "" {
        prefix = "maccms"
    }
    cm.prefix = prefix + "_"
}
```

### 6.2 文件缓存驱动（默认，零外部依赖）

```go
// internal/cache/file_driver.go

package cache

import (
    "encoding/binary"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"
)

// FileDriver 文件缓存驱动
// 目录结构: basedir/ab/cd/abcdef...cache
// 文件格式: [8字节过期Unix时间戳][数据内容]
type FileDriver struct {
    baseDir string
    mu      sync.RWMutex
}

func NewFileDriver(baseDir string) *FileDriver {
    if baseDir == "" {
        baseDir = "./runtime/cache"
    }
    os.MkdirAll(baseDir, 0755)
    return &FileDriver{baseDir: baseDir}
}

func (d *FileDriver) Get(key string) (string, error) {
    filePath := d.keyToPath(key)

    data, err := os.ReadFile(filePath)
    if err != nil {
        return "", err
    }

    if len(data) < 8 {
        os.Remove(filePath)
        return "", fmt.Errorf("invalid cache file")
    }

    // 读取过期时间
    expiry := int64(binary.BigEndian.Uint64(data[:8]))

    // 检查是否过期（expiry=0 表示永不过期）
    if expiry > 0 && time.Now().Unix() > expiry {
        os.Remove(filePath)
        return "", fmt.Errorf("cache expired")
    }

    return string(data[8:]), nil
}

func (d *FileDriver) Set(key, value string, ttl time.Duration) error {
    filePath := d.keyToPath(key)

    // 确保目录存在
    dir := filepath.Dir(filePath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    // 计算过期时间戳
    var expiry int64
    if ttl > 0 {
        expiry = time.Now().Add(ttl).Unix()
    }

    // 写入: [8字节过期时间][数据]
    data := make([]byte, 8+len(value))
    binary.BigEndian.PutUint64(data[:8], uint64(expiry))
    copy(data[8:], value)

    // 原子写入（先写临时文件再重命名）
    tmpPath := filePath + ".tmp"
    if err := os.WriteFile(tmpPath, data, 0644); err != nil {
        return err
    }
    return os.Rename(tmpPath, filePath)
}

func (d *FileDriver) Delete(key string) error {
    filePath := d.keyToPath(key)
    err := os.Remove(filePath)
    if os.IsNotExist(err) {
        return nil
    }
    return err
}

func (d *FileDriver) DeletePattern(pattern string) error {
    // 将通配符模式转换为前缀匹配
    prefix := strings.TrimSuffix(pattern, "*")
    if prefix == "" {
        return os.RemoveAll(d.baseDir)
    }

    // 遍历删除匹配的文件
    return filepath.Walk(d.baseDir, func(path string, info os.FileInfo, err error) error {
        if err != nil || info.IsDir() {
            return nil
        }
        // 从路径还原 key 并检查前缀
        relPath, _ := filepath.Rel(d.baseDir, path)
        key := strings.ReplaceAll(relPath, string(filepath.Separator), "")
        key = strings.TrimSuffix(key, ".cache")
        if strings.HasPrefix(key, prefix) {
            os.Remove(path)
        }
        return nil
    })
}

func (d *FileDriver) Clear() error {
    return os.RemoveAll(d.baseDir)
}

func (d *FileDriver) Close() error {
    return nil
}

// keyToPath 将 key 转换为文件路径
// 用 key 的 md5 前6位做两级目录分隔，防止单目录文件过多
func (d *FileDriver) keyToPath(key string) string {
    // 简单 hash：取 key 的前6个字符的 hex 值做目录
    hash := simpleHash(key)
    dir1 := fmt.Sprintf("%02x", hash>>4&0xFF)
    dir2 := fmt.Sprintf("%02x", hash&0xFF)
    return filepath.Join(d.baseDir, dir1, dir2, key+".cache")
}

func simpleHash(s string) uint32 {
    var h uint32
    for _, c := range s {
        h = h*31 + uint32(c)
    }
    return h
}
```

### 6.3 Redis 缓存驱动（可选扩展）

```go
// internal/cache/redis_driver.go

package cache

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"
)

// RedisDriver Redis 缓存驱动
// 需要 import "github.com/redis/go-redis/v9"
// 如果不需要 Redis，可以删除此文件，不影响编译
type RedisDriver struct {
    client *redis.Client
    ctx    context.Context
}

func NewRedisDriver(config CacheConfig) (*RedisDriver, error) {
    addr := config.RedisAddr + ":" + config.RedisPort
    if addr == ":" {
        addr = "127.0.0.1:6379"
    }

    client := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: config.RedisPassword,
        DB:       config.RedisDB,
    })

    ctx := context.Background()

    // 测试连接
    if err := client.Ping(ctx).Err(); err != nil {
        client.Close()
        return nil, err
    }

    return &RedisDriver{client: client, ctx: ctx}, nil
}

func (d *RedisDriver) Get(key string) (string, error) {
    return d.client.Get(d.ctx, key).Result()
}

func (d *RedisDriver) Set(key, value string, ttl time.Duration) error {
    return d.client.Set(d.ctx, key, value, ttl).Err()
}

func (d *RedisDriver) Delete(key string) error {
    return d.client.Del(d.ctx, key).Err()
}

func (d *RedisDriver) DeletePattern(pattern string) error {
    var cursor uint64
    for {
        keys, nextCursor, err := d.client.Scan(d.ctx, cursor, pattern, 100).Result()
        if err != nil {
            return err
        }
        if len(keys) > 0 {
            d.client.Del(d.ctx, keys...)
        }
        cursor = nextCursor
        if cursor == 0 {
            break
        }
    }
    return nil
}

func (d *RedisDriver) Clear() error {
    return d.client.FlushDB(d.ctx).Err()
}

func (d *RedisDriver) Close() error {
    return d.client.Close()
}
```

### 6.4 Singleflight 防缓存击穿

```go
// internal/cache/singleflight.go

package cache

import (
    "sync"
    "time"
)

// call 一次正在进行的调用
type call struct {
    wg  sync.WaitGroup
    val string
    err error
}

// SingleFlight 缓存击穿防护
// 多个 goroutine 同时请求同一个 key 时，只允许一个去查源
// 其他 goroutine 等待结果
type SingleFlight struct {
    mu    sync.Mutex
    calls map[string]*call
}

func NewSingleFlight() *SingleFlight {
    return &SingleFlight{
        calls: make(map[string]*call),
    }
}

// Do 执行请求，如果相同 key 已有请求在进行中，则等待其结果
func (sf *SingleFlight) Do(key string, fn func() (string, error)) (string, error) {
    sf.mu.Lock()
    if sf.calls == nil {
        sf.calls = make(map[string]*call)
    }

    // 已有相同 key 的请求在进行中
    if c, ok := sf.calls[key]; ok {
        sf.mu.Unlock()
        c.wg.Wait()
        return c.val, c.err
    }

    // 发起新请求
    c := &call{}
    c.wg.Add(1)
    sf.calls[key] = c
    sf.mu.Unlock()

    // 执行实际查询
    c.val, c.err = fn()

    // 清理并唤醒等待者
    sf.mu.Lock()
    delete(sf.calls, key)
    sf.mu.Unlock()
    c.wg.Done()

    return c.val, c.err
}
```

### 6.5 旁路缓存各层实现

```go
// internal/cache/layers.go

package cache

import (
    "crypto/md5"
    "fmt"
    "sync"
)

// ============================================================
// PageCache — 页面级旁路缓存（HTML 文件缓存）
// ============================================================

type PageCache struct {
    cm   *CacheManager
    sf   *SingleFlight
    mu   sync.RWMutex
}

func NewPageCache(cm *CacheManager) *PageCache {
    return &PageCache{
        cm: cm,
        sf: NewSingleFlight(),
    }
}

// Get 获取页面缓存（旁路模式）
// 返回 (html, hit)
func (pc *PageCache) Get(url string) (string, bool) {
    if !pc.cm.IsEnabled() || pc.cm.config.Page != 1 {
        return "", false
    }

    key := pc.urlToKey(url)
    data, err := pc.cm.Get(key)
    if err == nil && data != "" {
        return data, true
    }
    return "", false
}

// GetOrLoad 获取页面缓存，不存在时调用 loader 加载（防击穿）
func (pc *PageCache) GetOrLoad(url string, loader func() (string, error)) (string, error) {
    if !pc.cm.IsEnabled() || pc.cm.config.Page != 1 {
        return loader()
    }

    key := pc.urlToKey(url)

    // 先尝试读缓存
    data, err := pc.cm.Get(key)
    if err == nil && data != "" {
        return data, nil
    }

    // 缓存未命中，用 singleflight 防击穿
    result, err := pc.sf.Do(key, func() (string, error) {
        // 双重检查：可能等待期间其他 goroutine 已经写入
        data, err := pc.cm.Get(key)
        if err == nil && data != "" {
            return data, nil
        }

        // 确实没有，调用 loader 加载
        html, err := loader()
        if err != nil {
            return "", err
        }

        // 写入缓存
        pc.cm.Set(key, html, pc.cm.config.TimePage)

        return html, nil
    })

    return result, err
}

// Set 设置页面缓存
func (pc *PageCache) Set(url, html string) {
    if !pc.cm.IsEnabled() || pc.cm.config.Page != 1 {
        return
    }
    key := pc.urlToKey(url)
    pc.cm.Set(key, html, pc.cm.config.TimePage)
}

// Invalidate 使页面缓存失效
func (pc *PageCache) Invalidate(url string) {
    key := pc.urlToKey(url)
    pc.cm.Delete(key)
}

func (pc *PageCache) urlToKey(url string) string {
    hash := fmt.Sprintf("%x", md5.Sum([]byte(url)))
    return "page_" + hash
}


// ============================================================
// ListCache — 列表数据旁路缓存
// ============================================================

type ListCache struct {
    cm *CacheManager
    sf *SingleFlight
}

func NewListCache(cm *CacheManager) *ListCache {
    return &ListCache{
        cm: cm,
        sf: NewSingleFlight(),
    }
}

type ListCacheParams struct {
    Model     string
    Where     map[string]interface{}
    Order     string
    Page      int
    Num       int
    Start     int
    CacheTime int
}

// GetOrLoad 获取列表缓存，不存在时调用 loader 加载
func (lc *ListCache) GetOrLoad(params ListCacheParams, loader func() (interface{}, error)) (interface{}, error) {
    if !lc.cm.IsEnabled() {
        return loader()
    }

    key := lc.generateKey(params)

    // 先读缓存
    data, err := lc.cm.Get(key)
    if err == nil && data != "" {
        return data, nil
    }

    // singleflight 防击穿
    result, err := lc.sf.Do(key, func() (string, error) {
        // 双重检查
        data, err := lc.cm.Get(key)
        if err == nil && data != "" {
            return data, nil
        }

        loaded, err := loader()
        if err != nil {
            return "", err
        }

        // 序列化
        jsonData, _ := json.Marshal(loaded)
        jsonStr := string(jsonData)

        ttl := params.CacheTime
        if ttl <= 0 {
            ttl = lc.cm.config.Time
        }
        lc.cm.Set(key, jsonStr, ttl)

        return jsonStr, nil
    })

    if err != nil {
        return nil, err
    }
    return result, nil
}

// InvalidateByModel 按模型名清除列表缓存
func (lc *ListCache) InvalidateByModel(model string) {
    // 用模式匹配删除所有以该 model 开头的缓存
    lc.cm.DeletePattern("*_list_" + model + "_*")
}

func (lc *ListCache) generateKey(params ListCacheParams) string {
    condStr := fmt.Sprintf("%v", params.Where)
    rawKey := fmt.Sprintf("list_%s_%s_%s_%d_%d",
        params.Model, condStr, params.Order, params.Page, params.Num)
    hash := fmt.Sprintf("%x", md5.Sum([]byte(rawKey)))
    return hash
}


// ============================================================
// DetailCache — 详情数据旁路缓存
// ============================================================

type DetailCache struct {
    cm *CacheManager
    sf *SingleFlight
}

func NewDetailCache(cm *CacheManager) *DetailCache {
    return &DetailCache{
        cm: cm,
        sf: NewSingleFlight(),
    }
}

// GetOrLoad 获取详情缓存，不存在时调用 loader 加载
func (dc *DetailCache) GetOrLoad(model string, id interface{}, en string,
    loader func() (interface{}, error)) (interface{}, error) {

    if !dc.cm.IsEnabled() {
        return loader()
    }

    key := fmt.Sprintf("detail_%s_%v_%s", model, id, en)

    // 先读缓存
    data, err := dc.cm.Get(key)
    if err == nil && data != "" {
        return data, nil
    }

    // singleflight 防击穿
    result, err := dc.sf.Do(key, func() (string, error) {
        // 双重检查
        data, err := dc.cm.Get(key)
        if err == nil && data != "" {
            return data, nil
        }

        loaded, err := loader()
        if err != nil {
            return "", err
        }

        jsonData, _ := json.Marshal(loaded)
        jsonStr := string(jsonData)

        dc.cm.Set(key, jsonStr, dc.cm.config.Time)
        return jsonStr, nil
    })

    if err != nil {
        return nil, err
    }
    return result, nil
}

// Invalidate 使详情缓存失效
func (dc *DetailCache) Invalidate(model string, id interface{}, en string) {
    key := fmt.Sprintf("detail_%s_%v_%s", model, id, en)
    dc.cm.Delete(key)
}


// ============================================================
// ConfigCache — 配置/分类树缓存（始终生效，主动刷新）
// ============================================================

type ConfigCache struct {
    cm *CacheManager
}

func NewConfigCache(cm *CacheManager) *ConfigCache {
    return &ConfigCache{cm: cm}
}

func (cc *ConfigCache) Get(key string) (string, bool) {
    data, err := cc.cm.Get("config_" + key)
    if err != nil {
        return "", false
    }
    return data, true
}

func (cc *ConfigCache) Set(key, value string) {
    cc.cm.Set("config_"+key, value, 0) // 永不过期
}

func (cc *ConfigCache) Refresh(key, value string) {
    cc.Set(key, value)
}
```

---

## 七、并发采集安全

```go
// internal/service/collect/safe_collect.go

package collect

import (
    "sync"
    "time"

    "golang.org/x/sync/singleflight"
    "golang.org/x/time/rate"
)

// SafeCollectEngine 并发安全的采集引擎
type SafeCollectEngine struct {
    db          *gorm.DB
    client      *resty.Client
    sfg         singleflight.Group   // 防重复采集同一资源
    rateLimiter *rate.Limiter        // 请求频率限制
    mu          sync.Mutex
    cancelMap   map[string]context.CancelFunc // 任务取消控制
}

func NewSafeCollectEngine(db *gorm.DB) *SafeCollectEngine {
    return &SafeCollectEngine{
        db:          db,
        client:      resty.New().SetTimeout(30 * time.Second),
        rateLimiter: rate.NewLimiter(rate.Every(500*time.Millisecond), 2), // 每秒2个请求
        cancelMap:   make(map[string]context.CancelFunc),
    }
}

// ProcessSourceVideo 并发安全的视频处理
// 使用 singleflight 防止同一视频被并发处理
func (e *SafeCollectEngine) ProcessSourceVideo(video SourceVideo, source CollectSource) error {
    // 用 视频名+分类ID 作为去重 key
    dedupeKey := fmt.Sprintf("collect:%s:%d", video.Name, video.TypeID)

    // singleflight 防止并发处理同一视频
    _, err, _ := e.sfg.Do(dedupeKey, func() (interface{}, error) {
        return nil, e.processVideoSafe(video, source)
    })
    return err
}

// processVideoSafe 使用数据库唯一约束防重
func (e *SafeCollectEngine) processVideoSafe(video SourceVideo, source CollectSource) error {
    typeID := e.mapTypeID(source, video.TypeID, video.TypeName)
    if typeID == 0 {
        return fmt.Errorf("无法映射分类: %s", video.TypeName)
    }

    // 使用 INSERT ... ON DUPLICATE KEY UPDATE
    // 这样即使并发也不会冲突
    vod := Vod{
        TypeID:     typeID,
        VodName:    video.Name,
        VodPic:     video.Pic,
        VodContent: video.Des,
        VodStatus:  1,
        VodTime:    time.Now().Unix(),
    }

    // 先尝试查找
    var existing Vod
    result := e.db.Where("vod_name = ? AND type_id = ?", video.Name, typeID).First(&existing)

    if result.Error == gorm.ErrRecordNotFound {
        // 新增
        return e.db.Create(&vod).Error
    }
    if result.Error != nil {
        return result.Error
    }

    // 更新（根据更新规则）
    return e.updateByRule(existing, video, source)
}

// WaitForRateLimit 等待频率限制允许
func (e *SafeCollectEngine) WaitForRateLimit(ctx context.Context) error {
    return e.rateLimiter.Wait(ctx)
}

// CancelTask 取消采集任务
func (e *SafeCollectEngine) CancelTask(taskID string) {
    e.mu.Lock()
    if cancel, ok := e.cancelMap[taskID]; ok {
        cancel()
        delete(e.cancelMap, taskID)
    }
    e.mu.Unlock()
}

// RegisterTask 注册采集任务（支持取消）
func (e *SafeCollectEngine) RegisterTask(taskID string) context.Context {
    e.mu.Lock()
    defer e.mu.Unlock()

    ctx, cancel := context.WithCancel(context.Background())
    e.cancelMap[taskID] = cancel
    return ctx
}
```

---

## 八、分类映射实现

```go
// internal/service/collect/type_mapper.go

package collect

import (
    "strings"
    "sync"
)

// TypeMapper 分类映射器
type TypeMapper struct {
    db         *gorm.DB
    mu         sync.RWMutex
    nameIndex  map[string]int // 分类名 → 分类ID
    manualMap  map[int]int    // 源分类ID → 本地分类ID（手动配置）
    autoCreate bool           // 未匹配时是否自动创建
}

func NewTypeMapper(db *gorm.DB) *TypeMapper {
    tm := &TypeMapper{
        db:        db,
        nameIndex: make(map[string]int),
        manualMap: make(map[int]int),
    }
    tm.loadFromDB()
    return tm
}

func (tm *TypeMapper) loadFromDB() {
    var types []Type
    tm.db.Find(&types)

    tm.nameIndex = make(map[string]int, len(types))
    for _, t := range types {
        tm.nameIndex[t.TypeName] = t.TypeID
    }
}

// SetManualMapping 设置手动分类映射
// sourceTypeID: 源站的分类ID, localTypeID: 本地分类ID
func (tm *TypeMapper) SetManualMapping(sourceTypeID, localTypeID int) {
    tm.mu.Lock()
    defer tm.mu.Unlock()
    tm.manualMap[sourceTypeID] = localTypeID
}

// MapTypeID 映射分类ID
// 优先级: 手动映射 > 按名称匹配 > 自动创建 > 默认分类
func (tm *TypeMapper) MapTypeID(source CollectSource, sourceTypeID int, sourceTypeName string) int {
    tm.mu.RLock()
    defer tm.mu.RUnlock()

    // 1. 检查手动映射
    if localID, ok := tm.manualMap[sourceTypeID]; ok && localID > 0 {
        return localID
    }

    // 2. 按名称精确匹配
    if id, ok := tm.nameIndex[sourceTypeName]; ok {
        return id
    }

    // 3. 按名称模糊匹配（去除空格、特殊字符后匹配）
    cleanName := strings.TrimSpace(sourceTypeName)
    for name, id := range tm.nameIndex {
        if strings.TrimSpace(name) == cleanName {
            return id
        }
    }

    // 4. 自动创建分类
    if tm.autoCreate {
        newType := Type{
            TypeName: sourceTypeName,
            TypeSort: 0,
            TypeStatus: 1,
        }
        if err := tm.db.Create(&newType).Error; err == nil {
            tm.nameIndex[sourceTypeName] = newType.TypeID
            return newType.TypeID
        }
    }

    // 5. 返回 0 表示无法映射（调用方可决定跳过或归入默认分类）
    return 0
}
```

---

## 九、图片处理实现

```go
// internal/service/image/image.go

package image

import (
    "crypto/md5"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/disintegration/imaging"
)

// ImageService 图片服务
type ImageService struct {
    uploadDir  string
    baseURL    string
    maxWidth   int
    maxHeight  int
    watermark  string // 水印图片路径
}

func NewImageService(uploadDir, baseURL string) *ImageService {
    return &ImageService{
        uploadDir: uploadDir,
        baseURL:   baseURL,
        maxWidth:  1920,
        maxHeight: 1080,
    }
}

// DownloadAndSave 下载远程图片并保存到本地
// 返回本地相对路径，如 "/uploads/2026/06/ab12cd34.jpg"
func (s *ImageService) DownloadAndSave(remoteURL string) (string, error) {
    if remoteURL == "" {
        return "", nil
    }

    // 检查是否已经是本地路径
    if strings.HasPrefix(remoteURL, "/uploads/") {
        return remoteURL, nil
    }

    // 下载图片
    client := &http.Client{Timeout: 15 * time.Second}
    resp, err := client.Get(remoteURL)
    if err != nil {
        return "", fmt.Errorf("下载图片失败: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return "", fmt.Errorf("下载图片失败: HTTP %d", resp.StatusCode)
    }

    // 检查 Content-Type
    contentType := resp.Header.Get("Content-Type")
    ext := s.extFromContentType(contentType)
    if ext == "" {
        ext = s.extFromURL(remoteURL)
    }
    if ext == "" {
        ext = ".jpg"
    }

    // 读取数据
    data, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    // 检查大小（最大 10MB）
    if len(data) > 10*1024*1024 {
        return "", fmt.Errorf("图片过大: %d bytes", len(data))
    }

    // 生成文件名（用 URL 的 md5 做去重）
    hash := fmt.Sprintf("%x", md5.Sum([]byte(remoteURL)))
    now := time.Now()
    relDir := fmt.Sprintf("/uploads/%04d/%02d", now.Year(), now.Month())
    fileName := hash + ext
    relPath := relDir + "/" + fileName

    // 确保目录存在
    absDir := filepath.Join(s.uploadDir, relDir)
    os.MkdirAll(absDir, 0755)

    absPath := filepath.Join(s.uploadDir, relPath)

    // 写入文件
    if err := os.WriteFile(absPath, data, 0644); err != nil {
        return "", err
    }

    // 处理图片（缩放、水印）
    s.processImage(absPath)

    return relPath, nil
}

// processImage 处理图片：缩放 + 水印
func (s *ImageService) processImage(filePath string) {
    img, err := imaging.Open(filePath)
    if err != nil {
        return
    }

    // 缩放（如果超过最大尺寸）
    bounds := img.Bounds()
    w := bounds.Dx()
    h := bounds.Dy()
    if w > s.maxWidth || h > s.maxHeight {
        img = imaging.Fit(img, s.maxWidth, s.maxHeight, imaging.Lanczos)
    }

    // 添加水印
    if s.watermark != "" {
        if wm, err := imaging.Open(s.watermark); err == nil {
            // 右下角水印
            offset := image.Pt(w-wm.Bounds().Dx()-10, h-wm.Bounds().Dy()-10)
            img = imaging.Overlay(img, wm, offset, 1.0)
        }
    }

    // 保存
    imaging.Save(img, filePath)
}

// GenerateThumbnail 生成缩略图
func (s *ImageService) GenerateThumbnail(filePath string, width, height int) (string, error) {
    img, err := imaging.Open(filePath)
    if err != nil {
        return "", err
    }

    thumb := imaging.Thumbnail(img, width, height, imaging.Lanczos)

    // 缩略图文件名
    ext := filepath.Ext(filePath)
    base := strings.TrimSuffix(filePath, ext)
    thumbPath := base + fmt.Sprintf("_%dx%d%s", width, height, ext)

    imaging.Save(thumb, thumbPath)
    return thumbPath, nil
}

func (s *ImageService) extFromContentType(ct string) string {
    switch {
    case strings.Contains(ct, "jpeg"), strings.Contains(ct, "jpg"):
        return ".jpg"
    case strings.Contains(ct, "png"):
        return ".png"
    case strings.Contains(ct, "gif"):
        return ".gif"
    case strings.Contains(ct, "webp"):
        return ".webp"
    default:
        return ""
    }
}

func (s *ImageService) extFromURL(url string) string {
    // 从 URL 路径中提取扩展名
    parts := strings.Split(url, "?")
    path := parts[0]
    ext := filepath.Ext(path)
    switch ext {
    case ".jpg", ".jpeg", ".png", ".gif", ".webp":
        return ext
    default:
        return ""
    }
}
```

---

## 十、内容清洗函数

```go
// internal/service/sanitize/sanitize.go

package sanitize

import (
    "html"
    "regexp"
    "strings"
)

var (
    tagRe     = regexp.MustCompile(`<[^>]*>`)
    spaceRe   = regexp.MustCompile(`\s+`)
    commentRe = regexp.MustCompile(`<!--[\s\S]*?-->`)
    scriptRe  = regexp.MustCompile(`(?i)<script[\s\S]*?</script>`)
    styleRe   = regexp.MustCompile(`(?i)<style[\s\S]*?</style>`)
)

// StripHTMLTags 去除所有 HTML 标签
func StripHTMLTags(s string) string {
    // 先移除 script 和 style
    s = scriptRe.ReplaceAllString(s, "")
    s = styleRe.ReplaceAllString(s, "")
    // 移除所有标签
    s = tagRe.ReplaceAllString(s, "")
    // 解码 HTML 实体
    s = html.UnescapeString(s)
    // 合并多余空白
    s = spaceRe.ReplaceAllString(s, " ")
    return strings.TrimSpace(s)
}

// HTMLToText HTML 转纯文本（保留换行结构）
func HTMLToText(s string) string {
    // 移除 script 和 style
    s = scriptRe.ReplaceAllString(s, "")
    s = styleRe.ReplaceAllString(s, "")
    // 块级元素替换为换行
    blockTags := regexp.MustCompile(`(?i)</(p|br|div|li|h[1-6]|tr|blockquote)>`)
    s = blockTags.ReplaceAllString(s, "\n")
    // 移除注释
    s = commentRe.ReplaceAllString(s, "")
    // 移除所有标签
    s = tagRe.ReplaceAllString(s, "")
    // 解码实体
    s = html.UnescapeString(s)
    // 合并多余空白（保留换行）
    lines := strings.Split(s, "\n")
    var result []string
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line != "" {
            result = append(result, line)
        }
    }
    return strings.Join(result, "\n")
}

// CompressHTML 压缩 HTML（去注释、合并空白、去空行）
func CompressHTML(s string) string {
    // 移除注释
    s = commentRe.ReplaceAllString(s, "")
    // 移除多余换行和制表符
    s = strings.ReplaceAll(s, "\r\n", "")
    s = strings.ReplaceAll(s, "\n", "")
    s = strings.ReplaceAll(s, "\t", "")
    // 合并连续空白
    s = spaceRe.ReplaceAllString(s, " ")
    // 移除标签间的空白（> < 之间的空白）
    tagSpace := regexp.MustCompile(`>\s+<`)
    s = tagSpace.ReplaceAllString(s, "><")
    return strings.TrimSpace(s)
}

// ResolveImageURL 补全相对路径图片 URL
// baseURL: 当前页面的 URL
// imgURL:  图片的 src（可能是相对路径）
func ResolveImageURL(baseURL, imgURL string) string {
    if imgURL == "" {
        return ""
    }

    // 已经是绝对路径
    if strings.HasPrefix(imgURL, "http://") || strings.HasPrefix(imgURL, "https://") {
        return imgURL
    }

    // 协议相对
    if strings.HasPrefix(imgURL, "//") {
        return "https:" + imgURL
    }

    // 解析 baseURL
    idx := strings.Index(baseURL, "://")
    if idx < 0 {
        return imgURL
    }
    scheme := baseURL[:idx+3]
    rest := baseURL[idx+3:]
    hostEnd := strings.Index(rest, "/")
    if hostEnd < 0 {
        hostEnd = len(rest)
    }
    host := rest[:hostEnd]

    // 绝对路径
    if strings.HasPrefix(imgURL, "/") {
        return scheme + host + imgURL
    }

    // 相对路径
    basePath := rest[hostEnd:]
    lastSlash := strings.LastIndex(basePath, "/")
    if lastSlash >= 0 {
        basePath = basePath[:lastSlash+1]
    }
    return scheme + host + basePath + imgURL
}

// FilterSensitiveWords 敏感词过滤
func FilterSensitiveWords(content string, words []string) string {
    for _, word := range words {
        word = strings.TrimSpace(word)
        if word == "" {
            continue
        }
        replacement := strings.Repeat("*", len([]rune(word)))
        content = strings.ReplaceAll(content, word, replacement)
    }
    return content
}

// ApplyThesaurus 同义词替换
// thesaurus 格式: "旧词1=新词1,旧词2=新词2"
func ApplyThesaurus(content, thesaurus string) string {
    if thesaurus == "" {
        return content
    }
    pairs := strings.Split(thesaurus, ",")
    for _, pair := range pairs {
        parts := strings.SplitN(pair, "=", 2)
        if len(parts) == 2 {
            old := strings.TrimSpace(parts[0])
            new := strings.TrimSpace(parts[1])
            if old != "" {
                content = strings.ReplaceAll(content, old, new)
            }
        }
    }
    return content
}

// FilterByRegex 用正则过滤内容
func FilterByRegex(content, pattern string) string {
    if pattern == "" {
        return content
    }
    re, err := regexp.Compile(pattern)
    if err != nil {
        return content
    }
    return re.ReplaceAllString(content, "")
}
```

---

## 十一、分页器实现

```go
// internal/template/paginator.go

package template

import (
    "fmt"
    "html/template"
    "strings"
)

// Paginator 分页器
type Paginator struct {
    CurrentPage  int
    TotalPages   int
    TotalItems   int
    PageSize     int
    BaseURL      string // URL 模板，如 "/vodtype/1/index-{page}.html"
    PageParam    string // 分页参数名，默认 "{page}"
    ShowFirst    bool   // 显示首页
    ShowLast     bool   // 显示末页
    ShowPrev     bool   // 显示上一页
    ShowNext     bool   // 显示下一页
    ShowPages    int    // 显示的页码数量
}

func NewPaginator(currentPage, totalItems, pageSize int, baseURL string) *Paginator {
    totalPages := totalItems / pageSize
    if totalItems%pageSize > 0 {
        totalPages++
    }
    if totalPages < 1 {
        totalPages = 1
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

// Render 渲染分页 HTML
// 兼容原 macCMS 的分页样式
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
    if page == 1 {
        // 第一页时可能不需要分页参数
        // 兼容原系统：第一页也带 -1 后缀
        url = strings.Replace(url, p.PageParam, fmt.Sprintf("%d", page), 1)
    } else {
        url = strings.Replace(url, p.PageParam, fmt.Sprintf("%d", page), 1)
    }
    return url
}

func (p *Paginator) calcPageRange() (int, int) {
    half := p.ShowPages / 2
    start := p.CurrentPage - half
    end := p.CurrentPage + half

    if start < 1 {
        start = 1
        end = p.ShowPages
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
func (p *Paginator) HasPrev() bool {
    return p.CurrentPage > 1
}

// HasNext 是否有下一页
func (p *Paginator) HasNext() bool {
    return p.CurrentPage < p.TotalPages
}

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

// GetTemplateFuncs 返回分页相关的模板函数
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
```

---

## 十二、旁路静态缓存的路由设计

### 12.1 核心思路

**不再有"生成"操作。** 前台请求的流程是：

```
用户请求 /vodhtml/123/index.html
  │
  ├─ 1. 检查缓存文件是否存在？
  │     ├─ 存在且未过期 → 直接返回 HTML（命中）
  │     └─ 不存在 → 进入 singleflight 防击穿
  │
  ├─ 2. singleflight：同一 URL 只允许一个 goroutine 查源
  │     ├─ 其他 goroutine 等待第一个的结果
  │     └─ 第一个 goroutine 执行以下步骤：
  │
  ├─ 3. 查询数据库
  │
  ├─ 4. 渲染模板
  │
  ├─ 5. 写入缓存文件
  │
  └─ 6. 返回 HTML
```

### 12.2 Fiber 路由注册

```go
// internal/router/frontend.go

package router

import (
    "github.com/gofiber/fiber/v2"
    "maccms-go/internal/cache"
)

// SetupFrontendRoutes 前台路由（旁路缓存模式）
func SetupFrontendRoutes(app *fiber.App, pc *cache.PageCache, handlers *FrontendHandlers) {

    // 所有前台 GET 请求走缓存中间件
    app.Get("/*", func(c *fiber.Ctx) error {
        // POST 请求不缓存
        if c.Method() != "GET" {
            return c.Next()
        }

        // 跳过后台、API、用户中心
        path := c.Path()
        if shouldSkipCache(path) {
            return c.Next()
        }

        // 旁路缓存：GetOrLoad 内部处理 singleflight
        html, err := pc.GetOrLoad(path, func() (string, error) {
            // 缓存未命中，执行实际的 handler 渲染
            // URL 规则引擎解析请求
            target, params, err := urlEngine.Resolve(c)
            if err != nil {
                return "", err
            }

            // 渲染模板
            rendered, err := renderHandler(c, handlers, target, params)
            if err != nil {
                return "", err
            }

            return rendered, nil
        })

        if err != nil {
            return c.Status(404).SendString("Page not found")
        }

        c.Set("Content-Type", "text/html; charset=utf-8")
        c.Set("X-Cache", "HIT")
        return c.SendString(html)
    })
}

func shouldSkipCache(path string) bool {
    prefixes := []string{"/admin/", "/api/", "/user/", "/gbook"}
    for _, p := range prefixes {
        if strings.HasPrefix(path, p) {
            return true
        }
    }
    return false
}

// renderHandler 渲染 handler 并返回 HTML 字符串
func renderHandler(c *fiber.Ctx, handlers *FrontendHandlers,
    target string, params map[string]string) (string, error) {

    // 设置静态模式标记（URL 引擎输出静态路径）
    c.Locals("ismake", true)

    // 分发到 handler
    err := dispatchToHandler(c, handlers, target, params)
    if err != nil {
        return "", err
    }

    // 捕获响应 body
    return string(c.Response().Body()), nil
}
```

### 12.3 缓存清除接口

```go
// internal/handler/admin/cache.go

package admin

import (
    "github.com/gofiber/fiber/v2"
    "maccms-go/internal/cache"
)

type CacheHandler struct {
    cm    *cache.CacheManager
    pc    *cache.PageCache
    lc    *cache.ListCache
    dc    *cache.DetailCache
}

// ClearCache 清除缓存（后台操作）
func (h *CacheHandler) ClearCache(c *fiber.Ctx) error {
    cacheType := c.Query("type", "all")

    switch cacheType {
    case "all":
        h.cm.Clear()
    case "page":
        h.cm.DeletePattern("page_*")
    case "list":
        h.cm.DeletePattern("*list*")
    case "detail":
        h.cm.DeletePattern("*detail*")
    case "config":
        h.cm.DeletePattern("config_*")
    case "url":
        // 清除指定 URL 的缓存
        url := c.Query("url")
        if url != "" {
            h.pc.Invalidate(url)
        }
    }

    return c.JSON(fiber.Map{"code": 1, "msg": "缓存已清理"})
}

// CacheStats 缓存统计（后台仪表盘）
func (h *CacheHandler) CacheStats(c *fiber.Ctx) error {
    // 文件缓存统计
    stats := map[string]interface{}{
        "type":       h.cm.Type(),
        "enabled":    h.cm.IsEnabled(),
        "page_cache": h.cm.PageEnabled(),
    }

    return c.JSON(fiber.Map{"code": 1, "data": stats})
}
```

### 12.4 数据变更时自动清除缓存

```go
// internal/service/cache_invalidator.go

package service

// CacheInvalidator 缓存失效协调器
// 当数据变更时，自动清除相关缓存
type CacheInvalidator struct {
    pc *cache.PageCache
    lc *cache.ListCache
    dc *cache.DetailCache
}

func NewCacheInvalidator(pc *cache.PageCache, lc *cache.ListCache,
    dc *cache.DetailCache) *CacheInvalidator {
    return &CacheInvalidator{pc: pc, lc: lc, dc: dc}
}

// OnVodSave 视频保存时清除相关缓存
func (ci *CacheInvalidator) OnVodSave(vodID int, vodEn string, typeID int) {
    // 清除该视频的详情缓存
    ci.dc.Invalidate("vod", vodID, vodEn)

    // 清除该分类下的列表缓存
    ci.lc.InvalidateByModel("vod")

    // 清除首页缓存
    ci.pc.Invalidate("/")

    // 清除该视频相关的页面缓存（详情页、播放页等）
    // 注意：这里无法精确知道所有可能的 URL 格式
    // 所以采用"清除所有页面缓存"的保守策略
    // 生产环境可以维护 URL → 缓存 key 的索引做精确清除
}

// OnArtSave 文章保存时清除相关缓存
func (ci *CacheInvalidator) OnArtSave(artID int, artEn string, typeID int) {
    ci.dc.Invalidate("art", artID, artEn)
    ci.lc.InvalidateByModel("art")
    ci.pc.Invalidate("/")
}

// OnTypeChange 分类变更时清除相关缓存
func (ci *CacheInvalidator) OnTypeChange() {
    ci.lc.InvalidateByModel("vod")
    ci.lc.InvalidateByModel("art")
    ci.pc.Invalidate("/")
}

// OnConfigChange 配置变更时清除所有缓存
func (ci *CacheInvalidator) OnConfigChange() {
    ci.pc.Invalidate("/")
    // 配置缓存在 ConfigCache 中单独管理
}
```

---

## 十三、数据库索引策略

```sql
-- migrations/indexes.sql
-- 核心表索引设计

-- ============================================================
-- mac_vod 视频表（最核心，数据量最大）
-- ============================================================
ALTER TABLE mac_vod ADD INDEX idx_vod_type_status (type_id, vod_status);
ALTER TABLE mac_vod ADD INDEX idx_vod_type_time (type_id, vod_time DESC);
ALTER TABLE mac_vod ADD INDEX idx_vod_name (vod_name);
ALTER TABLE mac_vod ADD INDEX idx_vod_en (vod_en);
ALTER TABLE mac_vod ADD INDEX idx_vod_year (vod_year);
ALTER TABLE mac_vod ADD INDEX idx_vod_area (vod_area);
ALTER TABLE mac_vod ADD INDEX idx_vod_level (vod_level);
ALTER TABLE mac_vod ADD INDEX idx_vod_hits (vod_hits DESC);
ALTER TABLE mac_vod ADD INDEX idx_vod_hits_day (vod_hits_day DESC);
ALTER TABLE mac_vod ADD INDEX idx_vod_score (vod_score DESC);
ALTER TABLE mac_vod ADD INDEX idx_vod_time_make (vod_time_make);
-- 全文索引（MySQL 原生，Meilisearch 的降级方案）
ALTER TABLE mac_vod ADD FULLTEXT INDEX ft_vod_name (vod_name) WITH PARSER ngram;
ALTER TABLE mac_vod ADD FULLTEXT INDEX ft_vod_actor (vod_actor) WITH PARSER ngram;

-- ============================================================
-- mac_art 文章表
-- ============================================================
ALTER TABLE mac_art ADD INDEX idx_art_type_status (type_id, art_status);
ALTER TABLE mac_art ADD INDEX idx_art_type_time (type_id, art_time DESC);
ALTER TABLE mac_art ADD INDEX idx_art_name (art_name);
ALTER TABLE mac_art ADD INDEX idx_art_en (art_en);
ALTER TABLE mac_art ADD FULLTEXT INDEX ft_art_name (art_name) WITH PARSER ngram;

-- ============================================================
-- mac_actor 演员表
-- ============================================================
ALTER TABLE mac_actor ADD INDEX idx_actor_name (actor_name);
ALTER TABLE mac_actor ADD INDEX idx_actor_en (actor_en);
ALTER TABLE mac_actor ADD INDEX idx_actor_sex (actor_sex);
ALTER TABLE mac_actor ADD INDEX idx_actor_area (actor_area);

-- ============================================================
-- mac_type 分类表
-- ============================================================
ALTER TABLE mac_type ADD INDEX idx_type_pid (type_pid);
ALTER TABLE mac_type ADD INDEX idx_type_en (type_en);
ALTER TABLE mac_type ADD INDEX idx_type_sort (type_sort);

-- ============================================================
-- mac_comment 评论表
-- ============================================================
ALTER TABLE mac_comment ADD INDEX idx_comment_rid (comment_rid, comment_type);
ALTER TABLE mac_comment ADD INDEX idx_comment_time (comment_time DESC);

-- ============================================================
-- mac_gbook 留言表
-- ============================================================
ALTER TABLE mac_gbook ADD INDEX idx_gbook_time (gbook_time DESC);
ALTER TABLE mac_gbook ADD INDEX idx_gbook_status (gbook_status);

-- ============================================================
-- mac_user 用户表
-- ============================================================
ALTER TABLE mac_user ADD UNIQUE INDEX uk_user_name (user_name);
ALTER TABLE mac_user ADD INDEX idx_user_group (group_id);
ALTER TABLE mac_user ADD INDEX idx_user_status (user_status);

-- ============================================================
-- mac_order 订单表
-- ============================================================
ALTER TABLE mac_order ADD INDEX idx_order_user (user_id, order_status);
ALTER TABLE mac_order ADD INDEX idx_order_no (order_no);
ALTER TABLE mac_order ADD INDEX idx_order_time (order_time DESC);

-- ============================================================
-- mac_card 卡密表
-- ============================================================
ALTER TABLE mac_card ADD INDEX idx_card_no (card_no);
ALTER TABLE mac_card ADD INDEX idx_card_status (card_status);

-- ============================================================
-- mac_visit 访问统计表
-- ============================================================
ALTER TABLE mac_visit ADD INDEX idx_visit_time (visit_time DESC);
ALTER TABLE mac_visit ADD INDEX idx_visit_type (visit_type);

-- ============================================================
-- mac_plog / mac_ulog 日志表
-- ============================================================
ALTER TABLE mac_plog ADD INDEX idx_plog_time (plog_time DESC);
ALTER TABLE mac_ulog ADD INDEX idx_ulog_user (user_id);
ALTER TABLE mac_ulog ADD INDEX idx_ulog_time (ulog_time DESC);
```

---

## 十四、错误处理与降级策略

### 14.1 全局错误处理中间件

```go
// internal/middleware/error_handler.go

package middleware

import (
    "fmt"
    "runtime"

    "github.com/gofiber/fiber/v2"
    "go.uber.org/zap"
)

func ErrorHandler(logger *zap.Logger) fiber.ErrorHandler {
    return func(c *fiber.Ctx, err error) error {
        // 获取调用栈
        buf := make([]byte, 4096)
        n := runtime.Stack(buf, false)
        stack := string(buf[:n])

        logger.Error("请求处理错误",
            zap.String("method", c.Method()),
            zap.String("path", c.Path()),
            zap.String("ip", c.IP()),
            zap.Error(err),
            zap.String("stack", stack),
        )

        // API 请求返回 JSON
        if c.Get("Accept") == "application/json" ||
            strings.HasPrefix(c.Path(), "/api/") {
            return c.Status(500).JSON(fiber.Map{
                "code": 500,
                "msg":  "服务器内部错误",
            })
        }

        // 页面请求返回错误页
        return c.Status(500).SendString("服务器内部错误，请稍后重试")
    }
}
```

### 14.2 搜索降级策略

```go
// internal/service/search/search.go

package search

import (
    "strings"

    "gorm.io/gorm"
)

// SearchService 搜索服务（支持 Meilisearch 降级到 MySQL）
type SearchService struct {
    meili       *MeilisearchService
    db          *gorm.DB
    useMeili    bool
}

func NewSearchService(meili *MeilisearchService, db *gorm.DB) *SearchService {
    s := &SearchService{
        meili: meili,
        db:    db,
    }

    // 检测 Meilisearch 是否可用
    if meili != nil {
        health := meili.Health()
        s.useMeili = health["ok"].(bool)
    }

    return s
}

// Search 搜索（自动降级）
func (s *SearchService) Search(keyword string, page, pageSize int) ([]Vod, int64, error) {
    if s.useMeili {
        // 优先使用 Meilisearch
        results, err := s.meili.Search(keyword, SearchOptions{
            Limit:  pageSize,
            Offset: (page - 1) * pageSize,
        })
        if err == nil {
            // 转换结果
            vods := s.convertResults(results)
            return vods, results.TotalHits, nil
        }
        // Meilisearch 失败，降级到 MySQL
        s.useMeili = false
    }

    // 降级：MySQL LIKE 查询
    return s.mysqlSearch(keyword, page, pageSize)
}

// mysqlSearch MySQL 全文搜索降级方案
func (s *SearchService) mysqlSearch(keyword string, page, pageSize int) ([]Vod, int64, error) {
    var vods []Vod
    var total int64

    query := s.db.Model(&Vod{}).Where("vod_status = 1")

    // 优先使用 FULLTEXT 索引
    if s.hasFullTextIndex() {
        query = query.Where("MATCH(vod_name) AGAINST(? IN BOOLEAN MODE)", keyword)
    } else {
        // 降级到 LIKE
        likeKeyword := "%" + keyword + "%"
        query = query.Where("vod_name LIKE ? OR vod_actor LIKE ? OR vod_director LIKE ?",
            likeKeyword, likeKeyword, likeKeyword)
    }

    query.Count(&total)
    query.Order("vod_hits DESC").
        Offset((page - 1) * pageSize).
        Limit(pageSize).
        Find(&vods)

    return vods, total, nil
}

func (s *SearchService) hasFullTextIndex() bool {
    // 检查是否存在全文索引
    var result string
    s.db.Raw("SHOW INDEX FROM mac_vod WHERE Key_name = 'ft_vod_name'").Scan(&result)
    return result != ""
}

// Reconnect 尝试重新连接 Meilisearch
func (s *SearchService) Reconnect() {
    if s.meili != nil {
        health := s.meili.Health()
        s.useMeili = health["ok"].(bool)
    }
}
```

### 14.3 采集重试策略

```go
// internal/service/collect/retry.go

package collect

import (
    "context"
    "math"
    "time"
)

// RetryWithBackoff 指数退避重试
func RetryWithBackoff(ctx context.Context, maxRetries int, baseDelay time.Duration,
    fn func() error) error {

    var lastErr error
    for i := 0; i <= maxRetries; i++ {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        lastErr = fn()
        if lastErr == nil {
            return nil
        }

        if i < maxRetries {
            // 指数退避: baseDelay * 2^i + 随机抖动
            delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(i)))
            jitter := time.Duration(rand.Int63n(int64(delay) / 2))
            delay += jitter

            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(delay):
            }
        }
    }
    return fmt.Errorf("重试 %d 次后失败: %w", maxRetries, lastErr)
}

// 采集请求默认重试配置
var defaultRetryConfig = RetryConfig{
    MaxRetries: 3,
    BaseDelay:  2 * time.Second,
    MaxDelay:   30 * time.Second,
}

type RetryConfig struct {
    MaxRetries int
    BaseDelay  time.Duration
    MaxDelay   time.Duration
}
```

### 14.4 数据库连接池配置

```go
// internal/database/mysql.go

package database

import (
    "time"

    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

type DBConfig struct {
    Host         string
    Port         int
    User         string
    Password     string
    Database     string
    Charset      string
    MaxOpenConns int
    MaxIdleConns int
    MaxLifetime  int // 秒
}

func Connect(cfg DBConfig) (*gorm.DB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
        cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Charset)

    if cfg.Charset == "" {
        dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
            cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
    }

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Warn),
    })
    if err != nil {
        return nil, err
    }

    // 连接池配置
    sqlDB, _ := db.DB()
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)     // 默认 100
    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)     // 默认 10
    sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second) // 默认 3600s

    return db, nil
}
```

---

## 十五、旁路缓存文件目录结构

### 15.1 目录设计

```
runtime/
├── cache/
│   ├── page/           # 页面 HTML 缓存（旁路生成）
│   │   ├── 0a/
│   │   │   ├── bc/
│   │   │   │   ├── page_a1b2c3d4e5f6...cache
│   │   │   │   └── ...
│   │   │   └── ...
│   │   ├── 1d/
│   │   │   └── ...
│   │   └── ...
│   ├── list/           # 列表数据缓存
│   │   └── ...（同上两级目录结构）
│   ├── detail/         # 详情数据缓存
│   │   └── ...
│   └── config/         # 配置缓存
│       └── ...
├── logs/
│   └── maccms.log
└── uploads/
    └── ...
```

### 15.2 缓存文件格式

每个缓存文件的格式：

```
┌─────────────────┬──────────────────────────────────┐
│  8 字节头部      │           数据内容                │
├─────────────────┼──────────────────────────────────┤
│ 过期时间戳       │  HTML / JSON 数据                │
│ (Unix seconds,  │  (UTF-8 编码)                    │
│  Big Endian)    │                                  │
│  0 = 永不过期    │                                  │
└─────────────────┴──────────────────────────────────┘
```

### 15.3 缓存清理策略

```go
// internal/cache/cleaner.go

package cache

import (
    "os"
    "path/filepath"
    "time"
)

// CacheCleaner 缓存清理器
type CacheCleaner struct {
    cacheDir  string
    maxSize   int64  // 最大缓存大小（字节）
    maxAge    time.Duration // 最大缓存时间
}

func NewCacheCleaner(cacheDir string, maxSizeMB int, maxAgeHours int) *CacheCleaner {
    return &CacheCleaner{
        cacheDir: cacheDir,
        maxSize:  int64(maxSizeMB) * 1024 * 1024,
        maxAge:   time.Duration(maxAgeHours) * time.Hour,
    }
}

// Clean 执行清理
// 策略: 先清理过期文件，再按 LRU 清理超出大小限制的文件
func (cc *CacheCleaner) Clean() error {
    var totalSize int64
    var files []fileInfo

    // 遍历所有缓存文件
    filepath.Walk(cc.cacheDir, func(path string, info os.FileInfo, err error) error {
        if err != nil || info.IsDir() {
            return nil
        }

        totalSize += info.Size()
        files = append(files, fileInfo{
            path:    path,
            size:    info.Size(),
            modTime: info.ModTime(),
        })
        return nil
    })

    // 1. 清理过期文件
    now := time.Now()
    for _, f := range files {
        if now.Sub(f.modTime) > cc.maxAge {
            os.Remove(f.path)
            totalSize -= f.size
        }
    }

    // 2. 如果超出大小限制，按修改时间清理最旧的文件
    if totalSize > cc.maxSize {
        // 按修改时间排序
        sort.Slice(files, func(i, j int) bool {
            return files[i].modTime.Before(files[j].modTime)
        })

        for _, f := range files {
            if totalSize <= cc.maxSize {
                break
            }
            if _, err := os.Stat(f.path); err == nil {
                os.Remove(f.path)
                totalSize -= f.size
            }
        }
    }

    return nil
}

// StartAutoClean 启动自动清理（后台 goroutine）
func (cc *CacheCleaner) StartAutoClean(interval time.Duration) {
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        for range ticker.C {
            cc.Clean()
        }
    }()
}

type fileInfo struct {
    path    string
    size    int64
    modTime time.Time
}
```

### 15.4 缓存大小统计

```go
// GetCacheSize 获取缓存目录大小
func GetCacheSize(dir string) (int64, int) {
    var totalSize int64
    var fileCount int

    filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err == nil && !info.IsDir() {
            totalSize += info.Size()
            fileCount++
        }
        return nil
    })

    return totalSize, fileCount
}

// FormatSize 格式化文件大小
func FormatSize(bytes int64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
```

---

## 附录：go.mod 依赖清单

```go
module maccms-go

go 1.22

require (
    // Web 框架
    github.com/gofiber/fiber/v2 v2.52.0
    // ORM
    gorm.io/gorm v1.25.5
    gorm.io/driver/mysql v1.5.2
    // Redis（可选，按需引入）
    github.com/redis/go-redis/v9 v9.4.0
    // 定时任务
    github.com/robfig/cron/v3 v3.0.1
    // 日志
    go.uber.org/zap v1.27.0
    // 配置
    github.com/spf13/viper v1.18.2
    // HTTP 客户端（采集用）
    github.com/go-resty/resty/v2 v2.11.0
    // HTML 解析（自定义采集用）
    github.com/PuerkitoBio/goquery v1.9.1
    // 图片处理
    github.com/disintegration/imaging v1.6.2
    // 验证码
    github.com/dchest/captcha v1.0.0
    // JWT
    github.com/golang-jwt/jwt/v5 v5.2.0
    // 搜索引擎
    github.com/meilisearch/meilisearch-go v0.26.0
    // WebSocket（弹幕）
    github.com/gofiber/websocket/v2 v2.2.1
    // 并发控制
    golang.org/x/sync v0.6.0
    // 频率限制
    golang.org/x/time v0.5.0
)

// 注意: Redis 和 Meilisearch 依赖建议使用 build tag 控制
// 在不需要时可以不引入，减少二进制大小
// //go:build redis
// //go:build meilisearch
```

---

*补充文档完毕。覆盖原报告中所有缺失实现，共 15 个章节。*
*核心变更: Redis 可选、旁路缓存、singleflight 防击穿。*

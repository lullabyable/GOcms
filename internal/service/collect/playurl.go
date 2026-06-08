package collect

import "strings"

// ParsePlayURLs 解析 "名称1$url1#名称2$url2" 格式
func ParsePlayURLs(text string) []PlayURL {
	if text == "" {
		return nil
	}
	var result []PlayURL
	episodes := strings.Split(text, "#")
	for _, ep := range episodes {
		ep = strings.TrimSpace(ep)
		if ep == "" {
			continue
		}
		parts := strings.SplitN(ep, "$", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			url := strings.TrimSpace(parts[1])
			if url != "" {
				result = append(result, PlayURL{Name: name, URL: url})
			}
		} else {
			url := strings.TrimSpace(parts[0])
			if url != "" {
				result = append(result, PlayURL{Name: "", URL: url})
			}
		}
	}
	return result
}

// ParsePlayFromURL 解析 vod_play_from 和 vod_play_url 字段
// from: "量子m3u8$$$非凡m3u8", url: "第1集$url1#第2集$url2$$$第1集$url3#第2集$url4"
func ParsePlayFromURL(fromStr, urlStr string) []PlayGroup {
	if fromStr == "" || urlStr == "" {
		return nil
	}
	froms := strings.Split(fromStr, "$$$")
	urls := strings.Split(urlStr, "$$$")

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
		group := PlayGroup{Flag: flag, URLs: ParsePlayURLs(urlText)}
		if len(group.URLs) > 0 {
			result = append(result, group)
		}
	}
	return result
}

// FormatPlayFrom 将 PlayGroup 格式化回 vod_play_from 字段
func FormatPlayFrom(groups []PlayGroup) string {
	flags := make([]string, 0, len(groups))
	for _, g := range groups {
		flags = append(flags, g.Flag)
	}
	return strings.Join(flags, "$$$")
}

// FormatPlayURL 将 PlayGroup 格式化回 vod_play_url 字段
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
func MergePlayList(existingFrom, existingURL string, newGroups []PlayGroup) (string, string) {
	existing := ParsePlayFromURL(existingFrom, existingURL)
	existingIndex := make(map[string]int)
	for i, g := range existing {
		existingIndex[g.Flag] = i
	}

	for _, newGroup := range newGroups {
		if idx, ok := existingIndex[newGroup.Flag]; ok {
			// 合并URL列表
			nameIndex := make(map[string]bool)
			for _, u := range existing[idx].URLs {
				nameIndex[u.Name] = true
			}
			for _, u := range newGroup.URLs {
				if !nameIndex[u.Name] {
					existing[idx].URLs = append(existing[idx].URLs, u)
				}
			}
		} else {
			existing = append(existing, newGroup)
			existingIndex[newGroup.Flag] = len(existing) - 1
		}
	}

	return FormatPlayFrom(existing), FormatPlayURL(existing)
}

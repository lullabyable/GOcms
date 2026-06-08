package urlpush

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
	"gocms/internal/model"
)

const batchSize = 2000

// Pusher URL推送接口
type Pusher interface {
	PushURLs(urls []string) error
	Name() string
}

// Manager 推送管理器
type Manager struct {
	db       *gorm.DB
	pushers  []Pusher
	client   *http.Client
}

// Config 推送配置
type Config struct {
	BaiduSite   string `json:"baidu_site"`
	BaiduToken  string `json:"baidu_token"`
	ShenmaSite  string `json:"shenma_site"`
	ShenmaToken string `json:"shenma_token"`
	SogouSite   string `json:"sogou_site"`
	SogouToken  string `json:"sogou_token"`
}

func NewManager(db *gorm.DB, cfg Config) *Manager {
	m := &Manager{
		db:     db,
		client: &http.Client{Timeout: 30 * time.Second},
	}

	if cfg.BaiduToken != "" {
		m.pushers = append(m.pushers, &BaiduPusher{
			site:  cfg.BaiduSite,
			token: cfg.BaiduToken,
			client: m.client,
		})
	}
	if cfg.ShenmaToken != "" {
		m.pushers = append(m.pushers, &ShenmaPusher{
			site:  cfg.ShenmaSite,
			token: cfg.ShenmaToken,
			client: m.client,
		})
	}
	if cfg.SogouToken != "" {
		m.pushers = append(m.pushers, &SogouPusher{
			site:  cfg.SogouSite,
			token: cfg.SogouToken,
			client: m.client,
		})
	}
	return m
}

// PushAll 推送到所有平台
func (m *Manager) PushAll(urls []string) []error {
	var errs []error
	for _, p := range m.pushers {
		if err := m.pushBatch(p, urls); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", p.Name(), err))
		}
	}
	return errs
}

// PushTo 推送到指定平台
func (m *Manager) PushTo(platform string, urls []string) error {
	for _, p := range m.pushers {
		if p.Name() == platform {
			return m.pushBatch(p, urls)
		}
	}
	return fmt.Errorf("未知平台: %s", platform)
}

func (m *Manager) pushBatch(p Pusher, urls []string) error {
	start := time.Now()
	var totalErr error
	total := len(urls)

	// 分批推送
	for i := 0; i < len(urls); i += batchSize {
		end := i + batchSize
		if end > len(urls) {
			end = len(urls)
		}
		batch := urls[i:end]
		if err := p.PushURLs(batch); err != nil {
			totalErr = err
		}
	}

	// 记录日志
	log := model.URLPushLog{
		Platform:  p.Name(),
		URLs:      strings.Join(urls, "\n"),
		Total:     total,
		Success:   total,
		CreatedAt: start.Unix(),
	}
	if totalErr != nil {
		log.Failed = total
		log.Success = 0
		log.Error = totalErr.Error()
	}
	m.db.Create(&log)

	return totalErr
}

// GetLogs 获取推送日志
func (m *Manager) GetLogs(platform string, page, pageSize int) ([]model.URLPushLog, int64, error) {
	var logs []model.URLPushLog
	var total int64
	query := m.db.Model(&model.URLPushLog{})
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

// BaiduPusher 百度站长平台推送
type BaiduPusher struct {
	site   string
	token  string
	client *http.Client
}

func (p *BaiduPusher) Name() string { return "baidu" }

func (p *BaiduPusher) PushURLs(urls []string) error {
	data := strings.Join(urls, "\n")
	apiURL := fmt.Sprintf("http://data.zz.baidu.com/urls?site=%s&token=%s", p.site, p.token)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("百度推送失败: %s", string(body))
	}

	var result struct {
		Success int    `json:"success"`
		Remain  int    `json:"remain"`
		Message string `json:"message"`
		Error   int    `json:"error"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}
	if result.Error != 0 {
		return fmt.Errorf("百度推送错误: %s", result.Message)
	}
	return nil
}

// ShenmaPusher 神马搜索推送
type ShenmaPusher struct {
	site   string
	token  string
	client *http.Client
}

func (p *ShenmaPusher) Name() string { return "shenma" }

func (p *ShenmaPusher) PushURLs(urls []string) error {
	data := strings.Join(urls, "\n")
	apiURL := fmt.Sprintf("http://data.zhanzhang.sm.cn/push?site=%s&user_name=%s&resource_name=mip_add", p.site, p.token)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("神马推送失败: %s", string(body))
	}

	var result struct {
		ErrorCode int    `json:"error_code"`
		ErrorMsg  string `json:"error_msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}
	if result.ErrorCode != 0 {
		return fmt.Errorf("神马推送错误: %s", result.ErrorMsg)
	}
	return nil
}

// SogouPusher 搜狗站长平台推送
type SogouPusher struct {
	site   string
	token  string
	client *http.Client
}

func (p *SogouPusher) Name() string { return "sogou" }

func (p *SogouPusher) PushURLs(urls []string) error {
	urlList := strings.Join(urls, "\n")
	apiURL := fmt.Sprintf("http://data.zz.sogou.com/uploads/urls?site=%s&token=%s", p.site, p.token)

	formData := url.Values{}
	formData.Set("urls", urlList)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("搜狗推送失败: %s", string(body))
	}

	var result struct {
		StatusCode int    `json:"status"`
		StatusMsg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}
	if result.StatusCode != 0 {
		return fmt.Errorf("搜狗推送错误: %s", result.StatusMsg)
	}
	return nil
}

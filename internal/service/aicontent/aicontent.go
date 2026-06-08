package aicontent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Config AI 配置
type Config struct {
	Provider string `json:"provider"` // openai/custom
	APIKey   string `json:"api_key"`
	Endpoint string `json:"endpoint"`
	Model    string `json:"model"`
	MaxTokens int   `json:"max_tokens"`
}

// Service AI 内容生成服务
type Service struct {
	config Config
	client *http.Client
}

func NewService(cfg Config) *Service {
	if cfg.Model == "" {
		cfg.Model = "gpt-3.5-turbo"
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 2000
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://api.openai.com/v1/chat/completions"
	}
	return &Service{
		config: cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// GenerateRequest 生成请求
type GenerateRequest struct {
	Prompt  string `json:"prompt"`
	Type    string `json:"type"`    // title/content/summary/tag/description
	MaxLen  int    `json:"max_len"`
	Style   string `json:"style"`   // formal/casual/professional
}

// GenerateResponse 生成响应
type GenerateResponse struct {
	Content string `json:"content"`
	Tokens  int    `json:"tokens"`
}

// Generate 生成内容
func (s *Service) Generate(req GenerateRequest) (*GenerateResponse, error) {
	if s.config.APIKey == "" {
		return nil, fmt.Errorf("AI API Key 未配置")
	}

	systemPrompt := s.buildSystemPrompt(req)
	userPrompt := req.Prompt

	requestBody := map[string]interface{}{
		"model": s.config.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"max_tokens": s.config.MaxTokens,
	}

	body, _ := json.Marshal(requestBody)
	httpReq, err := http.NewRequest("POST", s.config.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("AI 请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI 返回错误 [%d]: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析 AI 响应失败: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("AI 未返回内容")
	}

	return &GenerateResponse{
		Content: result.Choices[0].Message.Content,
		Tokens:  result.Usage.TotalTokens,
	}, nil
}

func (s *Service) buildSystemPrompt(req GenerateRequest) string {
	switch req.Type {
	case "title":
		return "你是一个专业的视频/文章标题生成器。请根据用户提供的关键词，生成吸引人的标题。只返回标题，不要其他内容。"
	case "content":
		return "你是一个专业的内容创作者。请根据用户的要求生成高质量的内容。"
	case "summary":
		return "你是一个专业的内容摘要生成器。请将用户提供的内容精简为简洁的摘要。"
	case "tag":
		return "你是一个标签生成器。请根据用户提供的内容，生成5-10个相关的标签关键词，用逗号分隔。"
	case "description":
		return "你是一个SEO优化专家。请根据用户提供的视频/文章信息，生成一段适合搜索引擎的描述文本。"
	default:
		return "你是一个专业的AI助手。"
	}
}

// GenerateTitle 生成标题
func (s *Service) GenerateTitle(keyword string) (string, error) {
	resp, err := s.Generate(GenerateRequest{
		Prompt: fmt.Sprintf("关键词: %s", keyword),
		Type:   "title",
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// GenerateSummary 生成摘要
func (s *Service) GenerateSummary(content string) (string, error) {
	resp, err := s.Generate(GenerateRequest{
		Prompt: content,
		Type:   "summary",
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// GenerateTags 生成标签
func (s *Service) GenerateTags(content string) (string, error) {
	resp, err := s.Generate(GenerateRequest{
		Prompt: content,
		Type:   "tag",
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// GenerateDescription 生成描述
func (s *Service) GenerateDescription(title, content string) (string, error) {
	resp, err := s.Generate(GenerateRequest{
		Prompt: fmt.Sprintf("标题: %s\n内容: %s", title, content),
		Type:   "description",
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// IsConfigured 是否已配置
func (s *Service) IsConfigured() bool {
	return s.config.APIKey != ""
}

package admin

import (
	"github.com/gofiber/fiber/v2"
	"gocms/internal/service/aicontent"
)

// AIContentHandler AI 内容生成处理器
type AIContentHandler struct {
	ai *aicontent.Service
}

func NewAIContentHandler(ai *aicontent.Service) *AIContentHandler {
	return &AIContentHandler{ai: ai}
}

// Generate 生成内容
func (h *AIContentHandler) Generate(c *fiber.Ctx) error {
	if !h.ai.IsConfigured() {
		return c.JSON(fiber.Map{"code": 0, "msg": "AI 服务未配置"})
	}

	prompt := c.FormValue("prompt")
	contentType := c.FormValue("type", "content")

	if prompt == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "请输入提示词"})
	}

	resp, err := h.ai.Generate(aicontent.GenerateRequest{
		Prompt: prompt,
		Type:   contentType,
	})
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": err.Error()})
	}

	return c.JSON(fiber.Map{"code": 1, "data": resp})
}

// GenerateTitle 生成标题
func (h *AIContentHandler) GenerateTitle(c *fiber.Ctx) error {
	keyword := c.FormValue("keyword")
	if keyword == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "请输入关键词"})
	}

	title, err := h.ai.GenerateTitle(keyword)
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 1, "data": title})
}

// GenerateSummary 生成摘要
func (h *AIContentHandler) GenerateSummary(c *fiber.Ctx) error {
	content := c.FormValue("content")
	if content == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "请输入内容"})
	}

	summary, err := h.ai.GenerateSummary(content)
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 1, "data": summary})
}

// GenerateTags 生成标签
func (h *AIContentHandler) GenerateTags(c *fiber.Ctx) error {
	content := c.FormValue("content")
	if content == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "请输入内容"})
	}

	tags, err := h.ai.GenerateTags(content)
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 1, "data": tags})
}

// Config 获取/保存 AI 配置
func (h *AIContentHandler) Config(c *fiber.Ctx) error {
	if c.Method() == "GET" {
		return c.JSON(fiber.Map{"code": 1, "data": fiber.Map{
			"configured": h.ai.IsConfigured(),
		}})
	}
	// POST: 配置通过 config.yaml 管理，此处提示
	return c.JSON(fiber.Map{"code": 1, "msg": "请通过配置文件修改 AI 设置"})
}

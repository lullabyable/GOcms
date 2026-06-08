package frontend

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"gocms/internal/service/chat"
)

// ChatHandler 前台聊天处理器
type ChatHandler struct {
	chat *chat.Service
}

func NewChatHandler(svc *chat.Service) *ChatHandler {
	return &ChatHandler{chat: svc}
}

// WebSocketUpgrade WebSocket 升级检查
func (h *ChatHandler) WebSocketUpgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// WebSocket WebSocket 连接
func (h *ChatHandler) WebSocket(c *websocket.Conn) {
	h.chat.HandleWebSocket(c)
}

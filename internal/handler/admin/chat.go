package admin

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/model"
	"gocms/internal/service/chat"
)

// ChatHandler 聊天管理处理器
type ChatHandler struct {
	chat *chat.Service
}

func NewChatHandler(svc *chat.Service) *ChatHandler {
	return &ChatHandler{chat: svc}
}

// RoomList 聊天室列表
func (h *ChatHandler) RoomList(c *fiber.Ctx) error {
	rooms, err := h.chat.GetRooms()
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "获取失败"})
	}
	return c.JSON(fiber.Map{"code": 1, "data": rooms})
}

// RoomCreate 创建聊天室
func (h *ChatHandler) RoomCreate(c *fiber.Ctx) error {
	room := &model.ChatRoom{
		Name:     c.FormValue("name"),
		Desc:     c.FormValue("desc"),
		MaxUsers: 500,
		Status:   1,
	}
	if m, err := strconv.Atoi(c.FormValue("max_users")); err == nil && m > 0 {
		room.MaxUsers = m
	}
	if err := h.chat.CreateRoom(room); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "创建失败"})
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "创建成功", "data": room})
}

// RoomUpdate 更新聊天室
func (h *ChatHandler) RoomUpdate(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.FormValue("room_id"))
	room := &model.ChatRoom{
		RoomID:   id,
		Name:     c.FormValue("name"),
		Desc:     c.FormValue("desc"),
		MaxUsers: 500,
		Status:   1,
	}
	if m, err := strconv.Atoi(c.FormValue("max_users")); err == nil {
		room.MaxUsers = m
	}
	if s, err := strconv.Atoi(c.FormValue("status")); err == nil {
		room.Status = s
	}
	if err := h.chat.UpdateRoom(room); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "更新失败"})
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "更新成功"})
}

// RoomDelete 删除聊天室
func (h *ChatHandler) RoomDelete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.chat.DeleteRoom(id); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "删除失败"})
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "删除成功"})
}

// History 聊天记录
func (h *ChatHandler) History(c *fiber.Ctx) error {
	roomID, _ := strconv.Atoi(c.Query("room_id", "1"))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "50"))

	messages, total, err := h.chat.GetHistory(roomID, page, pageSize)
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "查询失败"})
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":      messages,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// OnlineCount 在线人数
func (h *ChatHandler) OnlineCount(c *fiber.Ctx) error {
	roomID, _ := strconv.Atoi(c.Params("id"))
	count := h.chat.GetOnlineCount(roomID)
	return c.JSON(fiber.Map{"code": 1, "data": fiber.Map{"count": count}})
}

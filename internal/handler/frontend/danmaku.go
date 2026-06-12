package frontend

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"gocms/internal/model"
	"gorm.io/gorm"
)

// DanmakuHandler 弹幕处理器
type DanmakuHandler struct {
	db    *gorm.DB
	rooms map[int]*Room
	mu    sync.RWMutex
}

// Room 弹幕房间（按视频ID分组）
type Room struct {
	vodID     int
	clients   map[*websocket.Conn]string
	mu        sync.RWMutex
	broadcast chan *DanmakuMessage
}

// DanmakuMessage 弹幕消息
type DanmakuMessage struct {
	Action  string  `json:"action"`
	VodID   int     `json:"vod_id"`
	UserID  int     `json:"user_id"`
	Content string  `json:"content"`
	Time    float64 `json:"time"`
	Type    int     `json:"type"`
	Color   string  `json:"color"`
	Created int64   `json:"created"`
}

func NewDanmakuHandler(db *gorm.DB) *DanmakuHandler {
	h := &DanmakuHandler{
		db:    db,
		rooms: make(map[int]*Room),
	}
	return h
}

// getOrCreateRoom 获取或创建房间
func (h *DanmakuHandler) getOrCreateRoom(vodID int) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[vodID]; ok {
		return room
	}

	room := &Room{
		vodID:     vodID,
		clients:   make(map[*websocket.Conn]string),
		broadcast: make(chan *DanmakuMessage, 256),
	}
	h.rooms[vodID] = room

	// 启动广播协程
	go room.run()

	return room
}

// run 房间广播循环
func (r *Room) run() {
	for msg := range r.broadcast {
		data, _ := json.Marshal(msg)
		r.mu.RLock()
		for conn := range r.clients {
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				conn.Close()
				r.mu.RUnlock()
				r.mu.Lock()
				delete(r.clients, conn)
				r.mu.Unlock()
				r.mu.RLock()
			}
		}
		r.mu.RUnlock()
	}
}

// addClient 添加客户端
func (r *Room) addClient(conn *websocket.Conn, userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[conn] = userID
}

// removeClient 移除客户端
func (r *Room) removeClient(conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, conn)
}

// clientCount 客户端数量
func (r *Room) clientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

// WebSocketUpgrade WebSocket 升级检查
func (h *DanmakuHandler) WebSocketUpgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// WebSocket WebSocket 连接处理
func (h *DanmakuHandler) WebSocket(c *websocket.Conn) {
	vodIDStr := c.Params("vod_id")
	vodID, err := strconv.Atoi(vodIDStr)
	if err != nil {
		c.Close()
		return
	}

	userID := c.Query("user_id", "0")
	room := h.getOrCreateRoom(vodID)
	room.addClient(c, userID)

	defer func() {
		room.removeClient(c)
		c.Close()
		// 如果房间空了，清理
		if room.clientCount() == 0 {
			h.mu.Lock()
			delete(h.rooms, vodID)
			h.mu.Unlock()
		}
	}()

	log.Printf("[DANMAKU] 用户 %s 加入视频 %d 的弹幕房间", userID, vodID)

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			break
		}

		var dm DanmakuMessage
		if err := json.Unmarshal(msg, &dm); err != nil {
			continue
		}

		// 设置默认值
		if dm.Type == 0 {
			dm.Type = 0 // 滚动
		}
		if dm.Color == "" {
			dm.Color = "#ffffff"
		}
		dm.Content = html.EscapeString(dm.Content)
		dm.VodID = vodID
		dm.Created = time.Now().Unix()

		// 持久化到数据库
		danmaku := model.Danmaku{
			VodID:     vodID,
			UserID:    dm.UserID,
			Content:   dm.Content,
			Time:      dm.Time,
			Type:      dm.Type,
			Color:     dm.Color,
			CreatedAt: dm.Created,
		}
		h.db.Create(&danmaku)

		// 广播到房间
		room.broadcast <- &dm
	}
}

// History 弹幕历史查询 API
func (h *DanmakuHandler) History(c *fiber.Ctx) error {
	vodID, err := strconv.Atoi(c.Params("vod_id"))
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "无效的视频ID"})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "100"))
	if pageSize > 500 {
		pageSize = 500
	}

	var danmakus []model.Danmaku
	var total int64

	query := h.db.Model(&model.Danmaku{}).Where("vod_id = ?", vodID)
	query.Count(&total)
	query.Order("time ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&danmakus)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":      danmakus,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// Send 发送弹幕（HTTP API，非 WebSocket）
func (h *DanmakuHandler) Send(c *fiber.Ctx) error {
	vodID, err := strconv.Atoi(c.Params("vod_id"))
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "无效的视频ID"})
	}

	content := html.EscapeString(c.FormValue("content"))
	if content == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "弹幕内容不能为空"})
	}

	userID, _ := strconv.Atoi(c.FormValue("user_id", "0"))
	timeVal, _ := strconv.ParseFloat(c.FormValue("time"), 64)
	dmType, _ := strconv.Atoi(c.FormValue("type", "0"))
	color := c.FormValue("color", "#ffffff")

	danmaku := model.Danmaku{
		VodID:     vodID,
		UserID:    userID,
		Content:   content,
		Time:      timeVal,
		Type:      dmType,
		Color:     color,
		CreatedAt: time.Now().Unix(),
	}

	if err := h.db.Create(&danmaku).Error; err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "发送失败"})
	}

	// 广播到 WebSocket 房间
	room := h.getOrCreateRoom(vodID)
	msg := &DanmakuMessage{
		Action:  "send",
		VodID:   vodID,
		UserID:  userID,
		Content: content,
		Time:    timeVal,
		Type:    dmType,
		Color:   color,
		Created: danmaku.CreatedAt,
	}

	select {
	case room.broadcast <- msg:
	default:
		// 队列满了，丢弃
	}

	return c.JSON(fiber.Map{"code": 1, "msg": "发送成功", "data": danmaku})
}

// OnlineCount 获取在线人数
func (h *DanmakuHandler) OnlineCount(c *fiber.Ctx) error {
	vodID, err := strconv.Atoi(c.Params("vod_id"))
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "无效的视频ID"})
	}

	h.mu.RLock()
	room, ok := h.rooms[vodID]
	h.mu.RUnlock()

	count := 0
	if ok {
		count = room.clientCount()
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"vod_id": vodID,
			"count":  count,
		},
	})
}

// AdminList 后台弹幕列表
func (h *DanmakuHandler) AdminList(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	vodID, _ := strconv.Atoi(c.Query("vod_id", "0"))

	var danmakus []model.Danmaku
	var total int64

	query := h.db.Model(&model.Danmaku{})
	if vodID > 0 {
		query = query.Where("vod_id = ?", vodID)
	}
	query.Count(&total)
	query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&danmakus)

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":      danmakus,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// AdminDelete 删除弹幕
func (h *DanmakuHandler) AdminDelete(c *fiber.Ctx) error {
	ids := c.FormValue("ids")
	if ids == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "请选择要删除的弹幕"})
	}

	var idList []int
	for _, idStr := range splitStr(ids, ",") {
		if id, err := strconv.Atoi(idStr); err == nil {
			idList = append(idList, id)
		}
	}

	if len(idList) == 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "无效的ID"})
	}

	h.db.Where("danmaku_id IN ?", idList).Delete(&model.Danmaku{})
	return c.JSON(fiber.Map{"code": 1, "msg": fmt.Sprintf("已删除 %d 条弹幕", len(idList))})
}

func splitStr(s, sep string) []string {
	var result []string
	for _, part := range splitByChar(s, sep) {
		if trimmed := trimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitByChar(s string, sep string) []string {
	if len(sep) != 1 {
		return []string{s}
	}
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n') {
		end--
	}
	return s[start:end]
}

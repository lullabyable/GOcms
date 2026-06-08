package chat

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

// Message 聊天消息
type Message struct {
	Action    string `json:"action"`
	RoomID    int    `json:"room_id"`
	UserID    int    `json:"user_id"`
	UserName  string `json:"user_name"`
	Content   string `json:"content"`
	MsgType   int    `json:"msg_type"`
	CreatedAt int64  `json:"created_at"`
}

// Room 聊天室
type Room struct {
	RoomID  int
	clients map[*websocket.Conn]*ClientInfo
	mu      sync.RWMutex
	broadcast chan *Message
}

// ClientInfo 客户端信息
type ClientInfo struct {
	UserID   int
	UserName string
}

// Service 聊天服务
type Service struct {
	db    *gorm.DB
	rooms map[int]*Room
	mu    sync.RWMutex
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db:    db,
		rooms: make(map[int]*Room),
	}
}

// getOrCreateRoom 获取或创建房间
func (s *Service) getOrCreateRoom(roomID int) *Room {
	s.mu.Lock()
	defer s.mu.Unlock()

	if room, ok := s.rooms[roomID]; ok {
		return room
	}

	room := &Room{
		RoomID:    roomID,
		clients:   make(map[*websocket.Conn]*ClientInfo),
		broadcast: make(chan *Message, 256),
	}
	s.rooms[roomID] = room
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

// HandleWebSocket WebSocket 处理
func (s *Service) HandleWebSocket(c *websocket.Conn) {
	roomID := 0
	userID := 0
	userName := ""

	// 从查询参数获取信息
	roomIDStr := c.Query("room_id", "1")
	userIDStr := c.Query("user_id", "0")
	userName = c.Query("user_name", "匿名用户")

	// 简单解析
	for _, ch := range roomIDStr {
		roomID = roomID*10 + int(ch-'0')
	}
	for _, ch := range userIDStr {
		userID = userID*10 + int(ch-'0')
	}

	room := s.getOrCreateRoom(roomID)
	clientInfo := &ClientInfo{UserID: userID, UserName: userName}

	room.mu.Lock()
	room.clients[c] = clientInfo
	room.mu.Unlock()

	// 发送欢迎消息
	welcome := &Message{
		Action:    "system",
		RoomID:    roomID,
		Content:   userName + " 加入了聊天室",
		MsgType:   2,
		CreatedAt: time.Now().Unix(),
	}
	room.broadcast <- welcome

	// 保存系统消息
	s.saveMessage(roomID, 0, "系统", welcome.Content, 2)

	defer func() {
		room.mu.Lock()
		delete(room.clients, c)
		room.mu.Unlock()
		c.Close()

		// 发送离开消息
		leave := &Message{
			Action:    "system",
			RoomID:    roomID,
			Content:   userName + " 离开了聊天室",
			MsgType:   2,
			CreatedAt: time.Now().Unix(),
		}
		room.broadcast <- leave
		s.saveMessage(roomID, 0, "系统", leave.Content, 2)

		log.Printf("[CHAT] 用户 %s 离开房间 %d", userName, roomID)
	}()

	log.Printf("[CHAT] 用户 %s 加入房间 %d", userName, roomID)

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			break
		}

		var chatMsg Message
		if err := json.Unmarshal(msg, &chatMsg); err != nil {
			continue
		}

		chatMsg.Action = "message"
		chatMsg.RoomID = roomID
		chatMsg.UserID = userID
		chatMsg.UserName = userName
		chatMsg.MsgType = 0
		chatMsg.CreatedAt = time.Now().Unix()

		// 保存消息
		s.saveMessage(roomID, userID, userName, chatMsg.Content, 0)

		// 广播
		room.broadcast <- &chatMsg
	}
}

func (s *Service) saveMessage(roomID, userID int, userName, content string, msgType int) {
	msg := model.ChatMessage{
		RoomID:    roomID,
		UserID:    userID,
		UserName:  userName,
		Content:   content,
		MsgType:   msgType,
		CreatedAt: time.Now().Unix(),
	}
	s.db.Create(&msg)
}

// GetOnlineCount 获取在线人数
func (s *Service) GetOnlineCount(roomID int) int {
	s.mu.RLock()
	room, ok := s.rooms[roomID]
	s.mu.RUnlock()

	if !ok {
		return 0
	}
	room.mu.RLock()
	defer room.mu.RUnlock()
	return len(room.clients)
}

// GetHistory 获取历史消息
func (s *Service) GetHistory(roomID, page, pageSize int) ([]model.ChatMessage, int64, error) {
	var messages []model.ChatMessage
	var total int64
	query := s.db.Model(&model.ChatMessage{}).Where("room_id = ?", roomID)
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&messages).Error
	return messages, total, err
}

// GetRooms 获取聊天室列表
func (s *Service) GetRooms() ([]model.ChatRoom, error) {
	var rooms []model.ChatRoom
	err := s.db.Where("status = 1").Find(&rooms).Error
	return rooms, err
}

// CreateRoom 创建聊天室
func (s *Service) CreateRoom(room *model.ChatRoom) error {
	room.CreatedAt = time.Now().Unix()
	return s.db.Create(room).Error
}

// UpdateRoom 更新聊天室
func (s *Service) UpdateRoom(room *model.ChatRoom) error {
	return s.db.Save(room).Error
}

// DeleteRoom 删除聊天室
func (s *Service) DeleteRoom(roomID int) error {
	return s.db.Delete(&model.ChatRoom{}, roomID).Error
}

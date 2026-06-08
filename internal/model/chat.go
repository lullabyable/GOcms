package model

// ChatMessage 聊天消息模型
type ChatMessage struct {
	MsgID     int    `gorm:"primaryKey;column:msg_id" json:"msg_id"`
	RoomID    int    `gorm:"column:room_id;index" json:"room_id"`
	UserID    int    `gorm:"column:user_id" json:"user_id"`
	UserName  string `gorm:"column:user_name;size:50" json:"user_name"`
	Content   string `gorm:"column:content;type:text" json:"content"`
	MsgType   int    `gorm:"column:msg_type" json:"msg_type"` // 0=文本 1=图片 2=系统
	CreatedAt int64  `gorm:"column:created_at" json:"created_at"`
}

func (ChatMessage) TableName() string { return "mac_chat_msg" }

// ChatRoom 聊天室模型
type ChatRoom struct {
	RoomID    int    `gorm:"primaryKey;column:room_id" json:"room_id"`
	Name      string `gorm:"column:name;size:100" json:"name"`
	Desc      string `gorm:"column:desc;type:text" json:"desc"`
	MaxUsers  int    `gorm:"column:max_users" json:"max_users"`
	Status    int    `gorm:"column:status" json:"status"` // 0=关闭 1=开启
	CreatedAt int64  `gorm:"column:created_at" json:"created_at"`
}

func (ChatRoom) TableName() string { return "mac_chat_room" }

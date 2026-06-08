package model

// Live 直播模型
type Live struct {
	LiveID      int    `gorm:"primaryKey;column:live_id" json:"live_id"`
	TypeID      int    `gorm:"column:type_id" json:"type_id"`
	LiveName    string `gorm:"column:live_name" json:"live_name"`
	LiveEn      string `gorm:"column:live_en" json:"live_en"`
	LiveTime    string `gorm:"column:live_time" json:"live_time"`
	LivePic     string `gorm:"column:live_pic" json:"live_pic"`
	LiveURL     string `gorm:"column:live_url" json:"live_url"`
	LiveFrom    string `gorm:"column:live_from" json:"live_from"`
	LiveSort    int    `gorm:"column:live_sort" json:"live_sort"`
	LiveLevel   int    `gorm:"column:live_level" json:"live_level"`
	LiveStatus  int    `gorm:"column:live_status" json:"live_status"`
}

func (Live) TableName() string { return "mac_live" }

// Danmaku 弹幕模型
type Danmaku struct {
	DanmakuID  int     `gorm:"primaryKey;column:danmaku_id" json:"danmaku_id"`
	VodID      int     `gorm:"column:vod_id;index" json:"vod_id"`
	UserID     int     `gorm:"column:user_id" json:"user_id"`
	Content    string  `gorm:"column:content;type:text" json:"content"`
	Time       float64 `gorm:"column:time" json:"time"`
	Type       int     `gorm:"column:type" json:"type"`
	Color      string  `gorm:"column:color" json:"color"`
	CreatedAt  int64   `gorm:"column:created_at" json:"created_at"`
}

func (Danmaku) TableName() string { return "mac_danmaku" }

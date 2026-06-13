package model

// Plog 操作日志模型
type Plog struct {
	PlogID      int    `gorm:"primaryKey;column:plog_id" json:"plog_id"`
	AdminID     int    `gorm:"column:admin_id" json:"admin_id"`
	AdminName   string `gorm:"column:admin_name" json:"admin_name"`
	PlogIP      string `gorm:"column:plog_ip" json:"plog_ip"`
	PlogURL     string `gorm:"column:plog_url" json:"plog_url"`
	PlogTime    int64  `gorm:"column:plog_time" json:"plog_time"`
	PlogContent string `gorm:"column:plog_content" json:"plog_content"`
}

func (Plog) TableName() string { return "mac_plog" }

package model

// LoginBan 登录封禁记录
type LoginBan struct {
	ID       int    `gorm:"primaryKey;autoIncrement" json:"id"`
	IP       string `gorm:"column:ip;size:64;uniqueIndex" json:"ip"`
	Failures int    `gorm:"column:failures;default:0" json:"failures"`
	BanUntil int64  `gorm:"column:ban_until;default:0" json:"ban_until"` // Unix timestamp
	LastFail int64  `gorm:"column:last_fail;default:0" json:"last_fail"`
}

func (LoginBan) TableName() string { return "mac_login_ban" }

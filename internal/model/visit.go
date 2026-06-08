package model

// Visit 访问记录模型
type Visit struct {
	VisitID   int    `gorm:"primaryKey;column:visit_id" json:"visit_id"`
	URL       string `gorm:"column:url;size:500" json:"url"`
	IP        string `gorm:"column:ip;size:45" json:"ip"`
	UserAgent string `gorm:"column:user_agent;size:500" json:"user_agent"`
	Referer   string `gorm:"column:referer;size:500" json:"referer"`
	VisitTime int64  `gorm:"column:visit_time;index" json:"visit_time"`
	Date      string `gorm:"column:date;size:10;index" json:"date"`
	IsUnique  int    `gorm:"column:is_unique" json:"is_unique"`
}

func (Visit) TableName() string { return "mac_visit" }

// VisitStat 访问统计汇总模型
type VisitStat struct {
	StatID   int    `gorm:"primaryKey;column:stat_id" json:"stat_id"`
	Date     string `gorm:"column:date;size:10;uniqueIndex" json:"date"`
	PV       int    `gorm:"column:pv" json:"pv"`
	UV       int    `gorm:"column:uv" json:"uv"`
	IP       int    `gorm:"column:ip" json:"ip"`
	NewUsers int    `gorm:"column:new_users" json:"new_users"`
}

func (VisitStat) TableName() string { return "mac_visit_stat" }

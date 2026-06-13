package model

// Link 友情链接模型
type Link struct {
	LinkID      int    `gorm:"primaryKey;column:link_id" json:"link_id"`
	LinkType    int    `gorm:"column:link_type" json:"link_type"`
	LinkName    string `gorm:"column:link_name" json:"link_name"`
	LinkSort    int    `gorm:"column:link_sort" json:"link_sort"`
	LinkAddTime int64  `gorm:"column:link_add_time" json:"link_add_time"`
	LinkTime    int64  `gorm:"column:link_time" json:"link_time"`
	LinkURL     string `gorm:"column:link_url" json:"link_url"`
	LinkLogo    string `gorm:"column:link_logo" json:"link_logo"`
}

func (Link) TableName() string { return "mac_link" }

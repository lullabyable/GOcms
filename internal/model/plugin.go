package model

// Plugin 插件模型
type Plugin struct {
	PluginID  int    `gorm:"primaryKey;column:plugin_id" json:"plugin_id"`
	Name      string `gorm:"column:name;size:100;uniqueIndex" json:"name"`
	Title     string `gorm:"column:title;size:200" json:"title"`
	Version   string `gorm:"column:version;size:20" json:"version"`
	Author    string `gorm:"column:author;size:100" json:"author"`
	Desc      string `gorm:"column:desc;type:text" json:"desc"`
	Config    string `gorm:"column:config;type:text" json:"config"`
	Status    int    `gorm:"column:status" json:"status"` // 0=禁用 1=启用
	InstalledAt int64 `gorm:"column:installed_at" json:"installed_at"`
	UpdatedAt   int64 `gorm:"column:updated_at" json:"updated_at"`
}

func (Plugin) TableName() string { return "mac_plugin" }

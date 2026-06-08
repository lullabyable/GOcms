package model

// Type 分类模型
type Type struct {
	TypeID       int    `gorm:"primaryKey;column:type_id" json:"type_id"`
	TypeName     string `gorm:"column:type_name" json:"type_name"`
	TypeEn       string `gorm:"column:type_en" json:"type_en"`
	TypePID      int    `gorm:"column:type_pid" json:"type_pid"`
	TypeSort     int    `gorm:"column:type_sort" json:"type_sort"`
	TypeMid      int    `gorm:"column:type_mid" json:"type_mid"`
	TypeLetter   string `gorm:"column:type_letter" json:"type_letter"`
	TypeColor    string `gorm:"column:type_color" json:"type_color"`
	TypeTplList  string `gorm:"column:type_tpl_list" json:"type_tpl_list"`
	TypeTplDetail string `gorm:"column:type_tpl_detail" json:"type_tpl_detail"`
	TypeTplPlay  string `gorm:"column:type_tpl_play" json:"type_tpl_play"`
	TypeTplDown  string `gorm:"column:type_tpl_down" json:"type_tpl_down"`
	TypeKey      string `gorm:"column:type_key" json:"type_key"`
	TypeDes      string `gorm:"column:type_des" json:"type_des"`
	TypeTitle    string `gorm:"column:type_title" json:"type_title"`
	TypeJumpurl  string `gorm:"column:type_jumpurl" json:"type_jumpurl"`
	TypePic      string `gorm:"column:type_pic" json:"type_pic"`
	TypeStatus   int    `gorm:"column:type_status" json:"type_status"`
	TypeExtend   string `gorm:"column:type_extend;type:text" json:"type_extend"`
	TypeShowTpl  string `gorm:"column:type_show_tpl" json:"type_show_tpl"`
	TypeReadTpl  string `gorm:"column:type_read_tpl" json:"type_read_tpl"`
}

func (Type) TableName() string { return "mac_type" }

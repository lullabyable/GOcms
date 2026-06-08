package model

// Admin 管理员模型
type Admin struct {
	AdminID    int    `gorm:"primaryKey;column:admin_id" json:"admin_id"`
	AdminName  string `gorm:"column:admin_name" json:"admin_name"`
	AdminPwd   string `gorm:"column:admin_pwd" json:"admin_pwd"`
	AdminRole  int    `gorm:"column:admin_role" json:"admin_role"` // 1=超管 2=普通 3=只读
	AdminStatus int   `gorm:"column:admin_status" json:"admin_status"`
	AdminLastTime int64 `gorm:"column:admin_last_time" json:"admin_last_time"`
	AdminLastIP string `gorm:"column:admin_last_ip" json:"admin_last_ip"`
	AdminLoginNum int  `gorm:"column:admin_login_num" json:"admin_login_num"`
}

func (Admin) TableName() string { return "mac_admin" }

// Comment 评论模型
type Comment struct {
	CommentID     int    `gorm:"primaryKey;column:comment_id" json:"comment_id"`
	CommentRID    int    `gorm:"column:comment_rid" json:"comment_rid"`
	CommentType   int    `gorm:"column:comment_type" json:"comment_type"`
	UserID        int    `gorm:"column:user_id" json:"user_id"`
	CommentContent string `gorm:"column:comment_content;type:text" json:"comment_content"`
	CommentTime   int64  `gorm:"column:comment_time" json:"comment_time"`
	CommentStatus int    `gorm:"column:comment_status" json:"comment_status"`
}

func (Comment) TableName() string { return "mac_comment" }

// Gbook 留言模型
type Gbook struct {
	GbookID      int    `gorm:"primaryKey;column:gbook_id" json:"gbook_id"`
	UserID       int    `gorm:"column:user_id" json:"user_id"`
	GbookContent string `gorm:"column:gbook_content;type:text" json:"gbook_content"`
	GbookTime    int64  `gorm:"column:gbook_time" json:"gbook_time"`
	GbookStatus  int    `gorm:"column:gbook_status" json:"gbook_status"`
	GbookReply   string `gorm:"column:gbook_reply;type:text" json:"gbook_reply"`
	GbookReplyTime int64 `gorm:"column:gbook_reply_time" json:"gbook_reply_time"`
}

func (Gbook) TableName() string { return "mac_gbook" }

// Config 系统配置模型
type Config struct {
	ID    int    `gorm:"primaryKey;column:config_id" json:"config_id"`
	Type  string `gorm:"column:type" json:"type"`
	Name  string `gorm:"column:name" json:"name"`
	Value string `gorm:"column:value;type:text" json:"value"`
}

func (Config) TableName() string { return "mac_config" }

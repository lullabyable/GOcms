package model

// User 用户模型
type User struct {
	UserID        int    `gorm:"primaryKey;column:user_id" json:"user_id"`
	GroupID       int    `gorm:"column:group_id" json:"group_id"`
	UserName      string `gorm:"column:user_name" json:"user_name"`
	UserPwd       string `gorm:"column:user_pwd" json:"-"`
	UserNickName  string `gorm:"column:user_nick_name" json:"user_nick_name"`
	UserEmail     string `gorm:"column:user_email" json:"user_email"`
	UserPhone     string `gorm:"column:user_phone" json:"user_phone"`
	UserPortrait  string `gorm:"column:user_portrait" json:"user_portrait"`
	UserPoints    int    `gorm:"column:user_points" json:"user_points"`
	UserPointsDay int    `gorm:"column:user_points_day" json:"user_points_day"`
	UserStatus    int    `gorm:"column:user_status" json:"user_status"`
	UserRegTime   int64  `gorm:"column:user_reg_time" json:"user_reg_time"`
	UserRegIP     string `gorm:"column:user_reg_ip" json:"user_reg_ip"`
	UserLoginTime int64  `gorm:"column:user_login_time" json:"user_login_time"`
	UserLoginIP   string `gorm:"column:user_login_ip" json:"user_login_ip"`
	UserLastTime  int64  `gorm:"column:user_last_time" json:"user_last_time"`
	UserLastIP    string `gorm:"column:user_last_ip" json:"user_last_ip"`
	UserLoginNum  int    `gorm:"column:user_login_num" json:"user_login_num"`
	ExpiryTime    int64  `gorm:"column:expiry_time" json:"expiry_time"`
}

func (User) TableName() string { return "mac_user" }

// Group 用户组模型
type Group struct {
	GroupID     int    `gorm:"primaryKey;column:group_id" json:"group_id"`
	GroupName   string `gorm:"column:group_name" json:"group_name"`
	GroupType   int    `gorm:"column:group_type" json:"group_type"`
	GroupPoints int    `gorm:"column:group_points" json:"group_points"`
	GroupState  int    `gorm:"column:group_state" json:"group_state"`
}

func (Group) TableName() string { return "mac_group" }

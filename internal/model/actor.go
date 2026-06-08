package model

// Actor 演员模型
type Actor struct {
	ActorID         int    `gorm:"primaryKey;column:actor_id" json:"actor_id"`
	ActorName       string `gorm:"column:actor_name" json:"actor_name"`
	ActorEn         string `gorm:"column:actor_en" json:"actor_en"`
	ActorSex        int    `gorm:"column:actor_sex" json:"actor_sex"`
	ActorArea       string `gorm:"column:actor_area" json:"actor_area"`
	ActorBirthday   string `gorm:"column:actor_birthday" json:"actor_birthday"`
	ActorBirtharea  string `gorm:"column:actor_birtharea" json:"actor_birtharea"`
	ActorStar       string `gorm:"column:actor_star" json:"actor_star"`
	ActorHeight     string `gorm:"column:actor_height" json:"actor_height"`
	ActorWeight     string `gorm:"column:actor_weight" json:"actor_weight"`
	ActorPic        string `gorm:"column:actor_pic" json:"actor_pic"`
	ActorBlurb      string `gorm:"column:actor_blurb" json:"actor_blurb"`
	ActorContent    string `gorm:"column:actor_content;type:text" json:"actor_content"`
	ActorTag        string `gorm:"column:actor_tag" json:"actor_tag"`
	ActorLevel      int    `gorm:"column:actor_level" json:"actor_level"`
	ActorLock       int    `gorm:"column:actor_lock" json:"actor_lock"`
	ActorTime       string `gorm:"column:actor_time" json:"actor_time"`
	ActorHits       int    `gorm:"column:actor_hits" json:"actor_hits"`
	ActorHitsDay    int    `gorm:"column:actor_hits_day" json:"actor_hits_day"`
	ActorHitsWeek   int    `gorm:"column:actor_hits_week" json:"actor_hits_week"`
	ActorHitsMonth  int    `gorm:"column:actor_hits_month" json:"actor_hits_month"`
	ActorStatus     int    `gorm:"column:actor_status" json:"actor_status"`
}

func (Actor) TableName() string { return "mac_actor" }

// Role 角色模型
type Role struct {
	RoleID      int    `gorm:"primaryKey;column:role_id" json:"role_id"`
	RoleName    string `gorm:"column:role_name" json:"role_name"`
	RoleEn      string `gorm:"column:role_en" json:"role_en"`
	RoleSex     int    `gorm:"column:role_sex" json:"role_sex"`
	RolePic     string `gorm:"column:role_pic" json:"role_pic"`
	RoleActor   string `gorm:"column:role_actor" json:"role_actor"`
	RoleBlurb   string `gorm:"column:role_blurb" json:"role_blurb"`
	RoleContent string `gorm:"column:role_content;type:text" json:"role_content"`
	RoleSort    int    `gorm:"column:role_sort" json:"role_sort"`
	RoleLevel   int    `gorm:"column:role_level" json:"role_level"`
	RoleLock    int    `gorm:"column:role_lock" json:"role_lock"`
	RoleStatus  int    `gorm:"column:role_status" json:"role_status"`
}

func (Role) TableName() string { return "mac_role" }

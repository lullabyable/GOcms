package model

// Art 文章模型
type Art struct {
	ID           int    `gorm:"primaryKey;column:art_id" json:"art_id"`
	TypeID       int    `gorm:"column:type_id" json:"type_id"`
	TypeID1      int    `gorm:"column:type_id_1" json:"type_id_1"`
	ArtName      string `gorm:"column:art_name" json:"art_name"`
	ArtSub       string `gorm:"column:art_sub" json:"art_sub"`
	ArtEn        string `gorm:"column:art_en" json:"art_en"`
	ArtTime      string `gorm:"column:art_time" json:"art_time"`
	ArtLetter    string `gorm:"column:art_letter" json:"art_letter"`
	ArtColor     string `gorm:"column:art_color" json:"art_color"`
	ArtFrom      string `gorm:"column:art_from" json:"art_from"`
	ArtAuthor    string `gorm:"column:art_author" json:"art_author"`
	ArtTag       string `gorm:"column:art_tag" json:"art_tag"`
	ArtClass     string `gorm:"column:art_class" json:"art_class"`
	ArtPic       string `gorm:"column:art_pic" json:"art_pic"`
	ArtPicThumb  string `gorm:"column:art_pic_thumb" json:"art_pic_thumb"`
	ArtPicSlide  string `gorm:"column:art_pic_slide" json:"art_pic_slide"`
	ArtContent   string `gorm:"column:art_content;type:text" json:"art_content"`
	ArtBlurb     string `gorm:"column:art_blurb" json:"art_blurb"`
	ArtRemarks   string `gorm:"column:art_remarks" json:"art_remarks"`
	ArtJumpurl   string `gorm:"column:art_jumpurl" json:"art_jumpurl"`
	ArtLock      int    `gorm:"column:art_lock" json:"art_lock"`
	ArtLevel     int    `gorm:"column:art_level" json:"art_level"`
	ArtPoints    int    `gorm:"column:art_points" json:"art_points"`
	ArtHits      int    `gorm:"column:art_hits" json:"art_hits"`
	ArtHitsDay   int    `gorm:"column:art_hits_day" json:"art_hits_day"`
	ArtHitsWeek  int    `gorm:"column:art_hits_week" json:"art_hits_week"`
	ArtHitsMonth int    `gorm:"column:art_hits_month" json:"art_hits_month"`
	ArtUp        int    `gorm:"column:art_up" json:"art_up"`
	ArtDown      int    `gorm:"column:art_down" json:"art_down"`
	ArtStatus    int    `gorm:"column:art_status" json:"art_status"`
	ArtTimeAdd   int64  `gorm:"column:art_time_add" json:"art_time_add"`
	ArtTimeHits  int64  `gorm:"column:art_time_hits" json:"art_time_hits"`
	ArtTimeMake  int64  `gorm:"column:art_time_make" json:"art_time_make"`
	ArtNote      string `gorm:"column:art_note" json:"art_note"`
}

func (Art) TableName() string { return "mac_art" }

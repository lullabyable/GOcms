package model

// Manga 漫画模型
type Manga struct {
	ID              int    `gorm:"primaryKey;column:manga_id" json:"manga_id"`
	TypeID          int    `gorm:"column:type_id" json:"type_id"`
	TypeID1         int    `gorm:"column:type_id_1" json:"type_id_1"`
	MangaName       string `gorm:"column:manga_name" json:"manga_name"`
	MangaSub        string `gorm:"column:manga_sub" json:"manga_sub"`
	MangaEn         string `gorm:"column:manga_en" json:"manga_en"`
	MangaTime       string `gorm:"column:manga_time" json:"manga_time"`
	MangaClass      string `gorm:"column:manga_class" json:"manga_class"`
	MangaTag        string `gorm:"column:manga_tag" json:"manga_tag"`
	MangaPic        string `gorm:"column:manga_pic" json:"manga_pic"`
	MangaPicThumb   string `gorm:"column:manga_pic_thumb" json:"manga_pic_thumb"`
	MangaAuthor     string `gorm:"column:manga_author" json:"manga_author"`
	MangaBlurb      string `gorm:"column:manga_blurb" json:"manga_blurb"`
	MangaRemarks    string `gorm:"column:manga_remarks" json:"manga_remarks"`
	MangaArea       string `gorm:"column:manga_area" json:"manga_area"`
	MangaLang       string `gorm:"column:manga_lang" json:"manga_lang"`
	MangaYear       string `gorm:"column:manga_year" json:"manga_year"`
	MangaContent    string `gorm:"column:manga_content;type:text" json:"manga_content"`
	MangaPlayFrom   string `gorm:"column:manga_play_from" json:"manga_play_from"`
	MangaPlayURL    string `gorm:"column:manga_play_url;type:text" json:"manga_play_url"`
	MangaDownFrom   string `gorm:"column:manga_down_from" json:"manga_down_from"`
	MangaDownURL    string `gorm:"column:manga_down_url;type:text" json:"manga_down_url"`
	MangaHits       int    `gorm:"column:manga_hits" json:"manga_hits"`
	MangaHitsDay    int    `gorm:"column:manga_hits_day" json:"manga_hits_day"`
	MangaScore      string `gorm:"column:manga_score" json:"manga_score"`
	MangaStatus     int    `gorm:"column:manga_status" json:"manga_status"`
	MangaTimeAdd    int64  `gorm:"column:manga_time_add" json:"manga_time_add"`
	MangaTimeHits   int64  `gorm:"column:manga_time_hits" json:"manga_time_hits"`
}

func (Manga) TableName() string { return "mac_manga" }

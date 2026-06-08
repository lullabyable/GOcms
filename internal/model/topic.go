package model

import "time"

type Topic struct {
	TopicID       int       `gorm:"primaryKey;column:topic_id" json:"topic_id"`
	TopicName     string    `gorm:"column:topic_name" json:"topic_name"`
	TopicEn       string    `gorm:"column:topic_en" json:"topic_en"`
	TopicLetter   string    `gorm:"column:topic_letter" json:"topic_letter"`
	TopicColor    string    `gorm:"column:topic_color" json:"topic_color"`
	TopicPic      string    `gorm:"column:topic_pic" json:"topic_pic"`
	TopicPicThumb string    `gorm:"column:topic_pic_thumb" json:"topic_pic_thumb"`
	TopicPicSlide string    `gorm:"column:topic_pic_slide" json:"topic_pic_slide"`
	TopicKey      string    `gorm:"column:topic_key" json:"topic_key"`
	TopicDesc     string    `gorm:"column:topic_desc" json:"topic_desc"`
	TopicContent  string    `gorm:"column:topic_content;type:text" json:"topic_content"`
	TopicVodID    string    `gorm:"column:topic_vod_id;type:text" json:"topic_vod_id"`
	TopicArtID    string    `gorm:"column:topic_art_id;type:text" json:"topic_art_id"`
	TopicSort     int       `gorm:"column:topic_sort" json:"topic_sort"`
	TopicLevel    int       `gorm:"column:topic_level" json:"topic_level"`
	TopicHits     int       `gorm:"column:topic_hits" json:"topic_hits"`
	TopicHitsDay  int       `gorm:"column:topic_hits_day" json:"topic_hits_day"`
	TopicHitsWeek int       `gorm:"column:topic_hits_week" json:"topic_hits_week"`
	TopicStatus   int       `gorm:"column:topic_status" json:"topic_status"`
	TopicTime     time.Time `gorm:"column:topic_time" json:"topic_time"`
	TopicTimeMake time.Time `gorm:"column:topic_time_make" json:"topic_time_make"`
}

func (Topic) TableName() string { return "mac_topic" }

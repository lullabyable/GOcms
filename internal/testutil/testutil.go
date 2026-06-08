package testutil

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gocms/internal/model"
)

// TestDB 创建测试用内存数据库
func TestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}

	// 自动迁移所有模型
	err = db.AutoMigrate(
		&model.Vod{},
		&model.Art{},
		&model.Type{},
		&model.User{},
		&model.Group{},
		&model.Admin{},
		&model.Comment{},
		&model.Gbook{},
		&model.Config{},
		&model.Actor{},
		&model.Live{},
		&model.Danmaku{},
		&model.Visit{},
		&model.VisitStat{},
		&model.Task{},
		&model.URLPushLog{},
		&model.Plugin{},
		&model.CardKey{},
		&model.Order{},
		&model.Payment{},
		&model.ChatMessage{},
		&model.ChatRoom{},
	)
	if err != nil {
		t.Fatalf("自动迁移失败: %v", err)
	}

	return db
}

// SeedTestData 填充测试数据
func SeedTestData(db *gorm.DB) {
	// 分类
	db.Create(&model.Type{TypeID: 1, TypeName: "电影", TypeSort: 1})
	db.Create(&model.Type{TypeID: 2, TypeName: "电视剧", TypeSort: 2})

	// 视频
	db.Create(&model.Vod{
		ID:        1,
		TypeID:    1,
		VodName:   "测试视频1",
		VodStatus: 1,
		VodHits:   100,
		VodTime:   "2026-01-01 00:00:00",
	})
	db.Create(&model.Vod{
		ID:        2,
		TypeID:    2,
		VodName:   "测试视频2",
		VodStatus: 1,
		VodHits:   200,
		VodTime:   "2026-01-02 00:00:00",
	})

	// 文章
	db.Create(&model.Art{
		ID:        1,
		TypeID:    1,
		ArtName:   "测试文章1",
		ArtStatus: 1,
		ArtHits:   50,
	})

	// 用户
	db.Create(&model.User{
		UserID:   1,
		UserName: "testuser",
		UserPwd:  "123456",
		GroupID:  1,
	})

	// 管理员
	db.Create(&model.Admin{
		AdminID:     1,
		AdminName:   "admin",
		AdminPwd:    "admin123",
		AdminRole:   1,
		AdminStatus: 1,
	})
}

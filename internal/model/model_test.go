package model_test

import (
	"testing"

	"gocms/internal/testutil"
	"gocms/internal/model"
)

func TestVodTableName(t *testing.T) {
	v := model.Vod{}
	if v.TableName() != "mac_vod" {
		t.Errorf("expected mac_vod, got %s", v.TableName())
	}
}

func TestArtTableName(t *testing.T) {
	a := model.Art{}
	if a.TableName() != "mac_art" {
		t.Errorf("expected mac_art, got %s", a.TableName())
	}
}

func TestUserTableName(t *testing.T) {
	u := model.User{}
	if u.TableName() != "mac_user" {
		t.Errorf("expected mac_user, got %s", u.TableName())
	}
}

func TestAdminTableName(t *testing.T) {
	a := model.Admin{}
	if a.TableName() != "mac_admin" {
		t.Errorf("expected mac_admin, got %s", a.TableName())
	}
}

func TestDanmakuTableName(t *testing.T) {
	d := model.Danmaku{}
	if d.TableName() != "mac_danmaku" {
		t.Errorf("expected mac_danmaku, got %s", d.TableName())
	}
}

func TestVisitTableName(t *testing.T) {
	v := model.Visit{}
	if v.TableName() != "mac_visit" {
		t.Errorf("expected mac_visit, got %s", v.TableName())
	}
}

func TestTaskTableName(t *testing.T) {
	tk := model.Task{}
	if tk.TableName() != "mac_task" {
		t.Errorf("expected mac_task, got %s", tk.TableName())
	}
}

func TestOrderTableName(t *testing.T) {
	o := model.Order{}
	if o.TableName() != "mac_order" {
		t.Errorf("expected mac_order, got %s", o.TableName())
	}
}

func TestCardKeyTableName(t *testing.T) {
	c := model.CardKey{}
	if c.TableName() != "mac_card" {
		t.Errorf("expected mac_card, got %s", c.TableName())
	}
}

func TestPluginTableName(t *testing.T) {
	p := model.Plugin{}
	if p.TableName() != "mac_plugin" {
		t.Errorf("expected mac_plugin, got %s", p.TableName())
	}
}

func TestChatMessageTableName(t *testing.T) {
	c := model.ChatMessage{}
	if c.TableName() != "mac_chat_msg" {
		t.Errorf("expected mac_chat_msg, got %s", c.TableName())
	}
}

func TestLiveTableName(t *testing.T) {
	l := model.Live{}
	if l.TableName() != "mac_live" {
		t.Errorf("expected mac_live, got %s", l.TableName())
	}
}

func TestVodCRUD(t *testing.T) {
	db := testutil.TestDB(t)

	// Create
	vod := model.Vod{ID: 100, TypeID: 1, VodName: "CRUD测试", VodStatus: 1}
	if err := db.Create(&vod).Error; err != nil {
		t.Fatalf("创建视频失败: %v", err)
	}

	// Read
	var found model.Vod
	if err := db.First(&found, 100).Error; err != nil {
		t.Fatalf("查询视频失败: %v", err)
	}
	if found.VodName != "CRUD测试" {
		t.Errorf("expected CRUD测试, got %s", found.VodName)
	}

	// Update
	db.Model(&found).Update("vod_name", "更新后")
	var updated model.Vod
	db.First(&updated, 100)
	if updated.VodName != "更新后" {
		t.Errorf("expected 更新后, got %s", updated.VodName)
	}

	// Delete
	db.Delete(&model.Vod{}, 100)
	var count int64
	db.Model(&model.Vod{}).Where("vod_id = ?", 100).Count(&count)
	if count != 0 {
		t.Error("删除失败")
	}
}

func TestUserCRUD(t *testing.T) {
	db := testutil.TestDB(t)

	user := model.User{UserID: 100, UserName: "testcrud", UserPwd: "pwd", GroupID: 1}
	db.Create(&user)

	var found model.User
	db.First(&found, 100)
	if found.UserName != "testcrud" {
		t.Errorf("expected testcrud, got %s", found.UserName)
	}
}

func TestVisitCRUD(t *testing.T) {
	db := testutil.TestDB(t)

	visit := model.Visit{URL: "/test", IP: "127.0.0.1", Date: "2026-01-01", VisitTime: 1000}
	db.Create(&visit)

	var count int64
	db.Model(&model.Visit{}).Where("url = ?", "/test").Count(&count)
	if count != 1 {
		t.Errorf("expected 1 visit, got %d", count)
	}
}

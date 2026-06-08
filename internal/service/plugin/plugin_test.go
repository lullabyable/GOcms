package plugin_test

import (
	"testing"

	"gocms/internal/testutil"
	"gocms/internal/model"
	"gocms/internal/service/plugin"
)

// mockPlugin 测试用插件
type mockPlugin struct {
	name       string
	initialized bool
	enabled    bool
}

func (p *mockPlugin) Name() string        { return p.name }
func (p *mockPlugin) Title() string       { return "Mock Plugin" }
func (p *mockPlugin) Version() string     { return "1.0.0" }
func (p *mockPlugin) Author() string      { return "test" }
func (p *mockPlugin) Description() string { return "test plugin" }
func (p *mockPlugin) Init(cfg map[string]interface{}) error {
	p.initialized = true
	return nil
}
func (p *mockPlugin) Enable() error {
	p.enabled = true
	return nil
}
func (p *mockPlugin) Disable() error {
	p.enabled = false
	return nil
}
func (p *mockPlugin) OnHook(hook string, data interface{}) (interface{}, error) {
	return data, nil
}

func TestPluginInstall(t *testing.T) {
	db := testutil.TestDB(t)
	mgr := plugin.NewManager(db)

	p := &mockPlugin{name: "test_plugin"}
	mgr.Register(p)

	if err := mgr.Install("test_plugin"); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// 重复安装应失败
	if err := mgr.Install("test_plugin"); err == nil {
		t.Error("double install should fail")
	}
}

func TestPluginEnableDisable(t *testing.T) {
	db := testutil.TestDB(t)
	mgr := plugin.NewManager(db)

	p := &mockPlugin{name: "toggle_plugin"}
	mgr.Register(p)
	mgr.Install("toggle_plugin")

	// 启用
	if err := mgr.Enable("toggle_plugin"); err != nil {
		t.Fatalf("Enable failed: %v", err)
	}
	if !p.enabled {
		t.Error("plugin should be enabled")
	}

	// 验证数据库状态
	var plugin model.Plugin
	db.Where("name = ?", "toggle_plugin").First(&plugin)
	if plugin.Status != 1 {
		t.Errorf("expected status=1, got %d", plugin.Status)
	}

	// 禁用
	if err := mgr.Disable("toggle_plugin"); err != nil {
		t.Fatalf("Disable failed: %v", err)
	}
	if p.enabled {
		t.Error("plugin should be disabled")
	}
}

func TestPluginUninstall(t *testing.T) {
	db := testutil.TestDB(t)
	mgr := plugin.NewManager(db)

	p := &mockPlugin{name: "uninstall_plugin"}
	mgr.Register(p)
	mgr.Install("uninstall_plugin")

	if err := mgr.Uninstall("uninstall_plugin"); err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	var count int64
	db.Model(&model.Plugin{}).Where("name = ?", "uninstall_plugin").Count(&count)
	if count != 0 {
		t.Error("plugin should be removed from DB")
	}
}

func TestPluginGetList(t *testing.T) {
	db := testutil.TestDB(t)
	mgr := plugin.NewManager(db)

	mgr.Register(&mockPlugin{name: "p1"})
	mgr.Register(&mockPlugin{name: "p2"})
	mgr.Install("p1")
	mgr.Install("p2")

	list, err := mgr.GetList()
	if err != nil {
		t.Fatalf("GetList failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(list))
	}
}

func TestPluginConfig(t *testing.T) {
	db := testutil.TestDB(t)
	mgr := plugin.NewManager(db)

	mgr.Register(&mockPlugin{name: "cfg_plugin"})
	mgr.Install("cfg_plugin")

	// 保存配置
	cfg := map[string]interface{}{"key1": "value1", "key2": 42}
	if err := mgr.SaveConfig("cfg_plugin", cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// 读取配置
	saved, err := mgr.GetConfig("cfg_plugin")
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}
	if saved["key1"] != "value1" {
		t.Errorf("expected value1, got %v", saved["key1"])
	}
}

func TestPluginNotFound(t *testing.T) {
	db := testutil.TestDB(t)
	mgr := plugin.NewManager(db)

	err := mgr.Enable("nonexistent")
	if err == nil {
		t.Error("should fail for non-existent plugin")
	}
}

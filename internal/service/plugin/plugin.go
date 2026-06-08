package plugin

import (
	"encoding/json"
	"fmt"
	"sync"

	"gorm.io/gorm"
	"gocms/internal/model"
)

// Plugin 插件接口
type Plugin interface {
	Name() string
	Title() string
	Version() string
	Author() string
	Description() string
	Init(cfg map[string]interface{}) error
	Enable() error
	Disable() error
	OnHook(hook string, data interface{}) (interface{}, error)
}

// Manager 插件管理器
type Manager struct {
	db       *gorm.DB
	plugins  map[string]Plugin
	configs  map[string]map[string]interface{}
	mu       sync.RWMutex
}

func NewManager(db *gorm.DB) *Manager {
	return &Manager{
		db:      db,
		plugins: make(map[string]Plugin),
		configs: make(map[string]map[string]interface{}),
	}
}

// Register 注册插件
func (m *Manager) Register(p Plugin) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.plugins[p.Name()] = p
}

// Install 安装插件
func (m *Manager) Install(name string) error {
	m.mu.RLock()
	p, ok := m.plugins[name]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("插件 %s 未注册", name)
	}

	cfgJSON, _ := json.Marshal(m.configs[name])
	plugin := model.Plugin{
		Name:        p.Name(),
		Title:       p.Title(),
		Version:     p.Version(),
		Author:      p.Author(),
		Desc:        p.Description(),
		Config:      string(cfgJSON),
		Status:      0,
		InstalledAt: 0,
	}

	var existing model.Plugin
	if err := m.db.Where("name = ?", name).First(&existing).Error; err == nil {
		return fmt.Errorf("插件 %s 已安装", name)
	}

	return m.db.Create(&plugin).Error
}

// Uninstall 卸载插件
func (m *Manager) Uninstall(name string) error {
	m.mu.RLock()
	p, ok := m.plugins[name]
	m.mu.RUnlock()

	if ok {
		p.Disable()
	}

	return m.db.Where("name = ?", name).Delete(&model.Plugin{}).Error
}

// Enable 启用插件
func (m *Manager) Enable(name string) error {
	m.mu.RLock()
	p, ok := m.plugins[name]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("插件 %s 未注册", name)
	}

	if err := p.Enable(); err != nil {
		return err
	}

	return m.db.Model(&model.Plugin{}).Where("name = ?", name).Update("status", 1).Error
}

// Disable 禁用插件
func (m *Manager) Disable(name string) error {
	m.mu.RLock()
	p, ok := m.plugins[name]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("插件 %s 未注册", name)
	}

	if err := p.Disable(); err != nil {
		return err
	}

	return m.db.Model(&model.Plugin{}).Where("name = ?", name).Update("status", 0).Error
}

// TriggerHook 触发钩子
func (m *Manager) TriggerHook(hook string, data interface{}) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result interface{}
	for _, p := range m.plugins {
		r, err := p.OnHook(hook, data)
		if err != nil {
			continue
		}
		if r != nil {
			result = r
		}
	}
	return result, nil
}

// GetList 获取插件列表
func (m *Manager) GetList() ([]model.Plugin, error) {
	var plugins []model.Plugin
	err := m.db.Find(&plugins).Error
	return plugins, err
}

// GetConfig 获取插件配置
func (m *Manager) GetConfig(name string) (map[string]interface{}, error) {
	var plugin model.Plugin
	if err := m.db.Where("name = ?", name).First(&plugin).Error; err != nil {
		return nil, err
	}

	var cfg map[string]interface{}
	if plugin.Config != "" {
		json.Unmarshal([]byte(plugin.Config), &cfg)
	}
	return cfg, nil
}

// SaveConfig 保存插件配置
func (m *Manager) SaveConfig(name string, cfg map[string]interface{}) error {
	cfgJSON, _ := json.Marshal(cfg)
	return m.db.Model(&model.Plugin{}).Where("name = ?", name).Update("config", string(cfgJSON)).Error
}

// LoadEnabled 加载已启用的插件
func (m *Manager) LoadEnabled() {
	var plugins []model.Plugin
	m.db.Where("status = ?", 1).Find(&plugins)

	for _, p := range plugins {
		m.mu.RLock()
		plugin, ok := m.plugins[p.Name]
		m.mu.RUnlock()

		if ok {
			var cfg map[string]interface{}
			if p.Config != "" {
				json.Unmarshal([]byte(p.Config), &cfg)
			}
			plugin.Init(cfg)
			plugin.Enable()
		}
	}
}

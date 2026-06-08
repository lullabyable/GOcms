package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	DB       DBConfig       `mapstructure:"database"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Session  SessionConfig  `mapstructure:"session"`
	Log      LogConfig      `mapstructure:"log"`
	Upload   UploadConfig   `mapstructure:"upload"`
	Template TemplateConfig `mapstructure:"template"`
}

type TemplateConfig struct {
	Dir   string `mapstructure:"dir"`   // 模板根目录
	Theme string `mapstructure:"theme"` // 当前主题名
}

type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type DBConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	Charset      string `mapstructure:"charset"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxLifetime  int    `mapstructure:"max_lifetime"`
}

type CacheConfig struct {
	Type     string `mapstructure:"type"`
	Flag     string `mapstructure:"flag"`
	Core     int    `mapstructure:"core"`
	Time     int    `mapstructure:"time"`
	Page     int    `mapstructure:"page"`
	TimePage int    `mapstructure:"time_page"`
	FileDir  string `mapstructure:"file_dir"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type SessionConfig struct {
	Type    string `mapstructure:"type"`
	Name    string `mapstructure:"name"`
	MaxAge  int    `mapstructure:"max_age"`
	Secret  string `mapstructure:"secret"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

type UploadConfig struct {
	Dir        string `mapstructure:"dir"`
	MaxSize    int    `mapstructure:"max_size"`
	AllowedExt string `mapstructure:"allowed_ext"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = "./config/config.yaml"
	}

	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", path)
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 设置默认值
	setDefaults(&cfg)

	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.DB.Port == 0 {
		cfg.DB.Port = 3306
	}
	if cfg.DB.Charset == "" {
		cfg.DB.Charset = "utf8mb4"
	}
	if cfg.DB.MaxOpenConns == 0 {
		cfg.DB.MaxOpenConns = 100
	}
	if cfg.DB.MaxIdleConns == 0 {
		cfg.DB.MaxIdleConns = 10
	}
	if cfg.Cache.Type == "" {
		cfg.Cache.Type = "file"
	}
	if cfg.Cache.Flag == "" {
		cfg.Cache.Flag = "gocms"
	}
	if cfg.Cache.Time == 0 {
		cfg.Cache.Time = 3600
	}
	if cfg.Cache.FileDir == "" {
		cfg.Cache.FileDir = "./runtime/cache"
	}
	if cfg.Session.Name == "" {
		cfg.Session.Name = "gocms_session"
	}
	if cfg.Session.MaxAge == 0 {
		cfg.Session.MaxAge = 86400
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Upload.Dir == "" {
		cfg.Upload.Dir = "./web/uploads"
	}
	if cfg.Template.Dir == "" {
		cfg.Template.Dir = "./web/templates"
	}
	if cfg.Template.Theme == "" {
		cfg.Template.Theme = "default"
	}
}

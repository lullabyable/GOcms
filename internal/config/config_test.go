package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// 创建临时配置文件
	content := `
server:
  port: 9090
database:
  host: "localhost"
  database: "test_db"
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(content)
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("期望端口 9090，实际 %d", cfg.Server.Port)
	}
	if cfg.DB.Database != "test_db" {
		t.Errorf("期望数据库 test_db，实际 %s", cfg.DB.Database)
	}
	if cfg.Cache.Type != "file" {
		t.Errorf("期望缓存类型 file，实际 %s", cfg.Cache.Type)
	}
	if cfg.DB.MaxOpenConns != 100 {
		t.Errorf("期望 MaxOpenConns 100，实际 %d", cfg.DB.MaxOpenConns)
	}
}

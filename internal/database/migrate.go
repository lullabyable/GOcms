package database

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

//go:embed sql/*.sql
var migrationFS embed.FS

type Migrator struct {
	db     *gorm.DB
	driver string // "mysql" 或 "sqlite"
}

func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db, driver: "mysql"}
}

func (m *Migrator) Migrate() error {
	if m.isNewInstall() {
		return m.freshInstall()
	}
	return m.upgrade()
}

func (m *Migrator) isNewInstall() bool {
	return !m.db.Migrator().HasTable("mac_config")
}

func (m *Migrator) freshInstall() error {
	sql, err := migrationFS.ReadFile("sql/install.sql")
	if err != nil {
		return fmt.Errorf("找不到安装SQL: %w", err)
	}
	return m.execSQL(string(sql))
}

func (m *Migrator) upgrade() error {
	current := m.getVersion()

	files, err := migrationFS.ReadDir("sql")
	if err != nil {
		return err
	}

	var upgrades []string
	for _, f := range files {
		name := f.Name()
		if strings.HasPrefix(name, "upgrade_") && strings.HasSuffix(name, ".sql") {
			upgrades = append(upgrades, name)
		}
	}
	sort.Strings(upgrades)

	for _, file := range upgrades {
		var ver int
		fmt.Sscanf(strings.TrimPrefix(strings.TrimSuffix(file, ".sql"), "upgrade_"), "%d", &ver)
		if ver <= current {
			continue
		}

		sql, err := migrationFS.ReadFile("sql/" + file)
		if err != nil {
			continue
		}
		if err := m.execSQL(string(sql)); err != nil {
			return fmt.Errorf("升级到 v%d 失败: %w", ver, err)
		}
		m.setVersion(ver)
		fmt.Printf("[MIGRATE] 升级到 v%d 成功\n", ver)
	}
	return nil
}

func (m *Migrator) RunFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return m.execSQL(string(data))
}

func (m *Migrator) RunDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		if err := m.RunFile(filepath.Join(dir, e.Name())); err != nil {
			return fmt.Errorf("执行 %s 失败: %w", e.Name(), err)
		}
	}
	return nil
}

func (m *Migrator) execSQL(sql string) error {
	for _, stmt := range strings.Split(sql, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		if err := m.db.Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}

func (m *Migrator) getVersion() int {
	m.db.Exec("CREATE TABLE IF NOT EXISTS mac_migrations (version INTEGER)")
	var ver int
	m.db.Raw("SELECT COALESCE(MAX(version), 0) FROM mac_migrations").Scan(&ver)
	return ver
}

func (m *Migrator) setVersion(ver int) {
	m.db.Exec("DELETE FROM mac_migrations")
	m.db.Exec("INSERT INTO mac_migrations (version) VALUES (?)", ver)
}

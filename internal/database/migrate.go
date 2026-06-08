package database

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	// 从 gorm.DB 检测驱动类型
	driver := "sqlite"
	if db.Dialector.Name() == "mysql" {
		driver = "mysql"
	}
	return &Migrator{db: db, driver: driver}
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
	// SQLite 兼容处理
	if m.driver == "sqlite" {
		sql = m.convertToSQLite(sql)
	}

	for _, stmt := range strings.Split(sql, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		if err := m.db.Exec(stmt).Error; err != nil {
			// SQLite 忽略已存在表的错误
			if m.driver == "sqlite" && strings.Contains(err.Error(), "already exists") {
				continue
			}
			return err
		}
	}
	return nil
}

// convertToSQLite 将 MySQL DDL 转换为 SQLite 兼容语法
func (m *Migrator) convertToSQLite(sql string) string {
	// 移除 ENGINE=InnoDB
	sql = strings.ReplaceAll(sql, " ENGINE=InnoDB", "")
	sql = strings.ReplaceAll(sql, " ENGINE=MyISAM", "")

	// 移除 DEFAULT CHARSET 和 COLLATE
	sql = regexpReplace(sql, `(?i)\s+DEFAULT\s+CHARSET=\w+`, "")
	sql = regexpReplace(sql, `(?i)\s+COLLATE\s+\w+`, "")

	// int unsigned → INTEGER
	sql = strings.ReplaceAll(sql, "int unsigned", "INTEGER")
	sql = strings.ReplaceAll(sql, "smallint unsigned", "INTEGER")
	sql = strings.ReplaceAll(sql, "tinyint unsigned", "INTEGER")
	sql = strings.ReplaceAll(sql, "mediumint unsigned", "INTEGER")
	sql = strings.ReplaceAll(sql, "bigint unsigned", "INTEGER")
	sql = strings.ReplaceAll(sql, "smallint", "INTEGER")
	sql = strings.ReplaceAll(sql, "mediumint", "INTEGER")
	sql = strings.ReplaceAll(sql, "bigint", "INTEGER")

	// AUTO_INCREMENT → AUTOINCREMENT
	sql = strings.ReplaceAll(sql, "AUTO_INCREMENT", "AUTOINCREMENT")

	// SQLite AUTOINCREMENT 必须紧跟 PRIMARY KEY
	// 修复 "INTEGER NOT NULL AUTOINCREMENT" → "INTEGER PRIMARY KEY AUTOINCREMENT"
	sql = regexpReplace(sql, `INTEGER\s+NOT\s+NULL\s+AUTOINCREMENT`, "INTEGER PRIMARY KEY AUTOINCREMENT")
	// 如果已经有 PRIMARY KEY 但不是 AUTOINCREMENT 模式，也要处理
	sql = regexpReplace(sql, `INTEGER\s+AUTOINCREMENT`, "INTEGER PRIMARY KEY AUTOINCREMENT")

	// 移除 KEY 定义行（保留 PRIMARY KEY）
	// 同时检查是否有列已经带了 PRIMARY KEY AUTOINCREMENT
	hasInlinePK := strings.Contains(sql, "PRIMARY KEY AUTOINCREMENT")

	lines := strings.Split(sql, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		upper := strings.ToUpper(trimmed)

		// 跳过非 PRIMARY KEY 的 KEY 定义
		if (strings.HasPrefix(upper, "KEY ") ||
			strings.HasPrefix(upper, "UNIQUE KEY") ||
			strings.HasPrefix(upper, "INDEX ")) &&
			!strings.Contains(upper, "PRIMARY") {
			if len(result) > 0 {
				last := result[len(result)-1]
				if strings.HasSuffix(last, ",") {
					result[len(result)-1] = strings.TrimSuffix(last, ",")
				}
			}
			continue
		}

		// 如果列已经有 PRIMARY KEY AUTOINCREMENT，移除表级 PRIMARY KEY 定义
		if hasInlinePK && strings.HasPrefix(upper, "PRIMARY KEY") {
			if len(result) > 0 {
				last := result[len(result)-1]
				if strings.HasSuffix(last, ",") {
					result[len(result)-1] = strings.TrimSuffix(last, ",")
				}
			}
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
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

// regexpReplace 正则替换
func regexpReplace(s, pattern, replacement string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(s, replacement)
}

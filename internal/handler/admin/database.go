package admin

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// DatabaseHandler 数据库管理处理器
type DatabaseHandler struct {
	db *gorm.DB
}

func NewDatabaseHandler(db *gorm.DB) *DatabaseHandler {
	return &DatabaseHandler{db: db}
}

// TableInfo 表信息
type TableInfo struct {
	Name    string `json:"name"`
	Engine  string `json:"engine"`
	Rows    int64  `json:"rows"`
	Size    string `json:"size"`
	DataLen int64  `json:"data_length"`
	IdxLen  int64  `json:"index_length"`
	Comment string `json:"comment"`
}

// List 列出所有数据表
func (h *DatabaseHandler) List(c *fiber.Ctx) error {
	var tables []TableInfo

	rows, err := h.db.Raw("SHOW TABLE STATUS").Rows()
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "查询失败"})
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	for rows.Next() {
		// 动态扫描列，取我们需要的字段
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)

		colMap := make(map[string]interface{})
		for i, col := range cols {
			colMap[col] = values[i]
		}

		ti := TableInfo{}
		if v, ok := colMap["Name"]; ok && v != nil {
			ti.Name = string(v.([]byte))
		}
		if v, ok := colMap["Engine"]; ok && v != nil {
			ti.Engine = string(v.([]byte))
		}
		if v, ok := colMap["Rows"]; ok && v != nil {
			switch vv := v.(type) {
			case int64:
				ti.Rows = vv
			}
		}
		if v, ok := colMap["Data_length"]; ok && v != nil {
			switch vv := v.(type) {
			case int64:
				ti.DataLen = vv
			}
		}
		if v, ok := colMap["Index_length"]; ok && v != nil {
			switch vv := v.(type) {
			case int64:
				ti.IdxLen = vv
			}
		}
		if v, ok := colMap["Comment"]; ok && v != nil {
			ti.Comment = string(v.([]byte))
		}

		total := ti.DataLen + ti.IdxLen
		ti.Size = formatFileSize(total)
		tables = append(tables, ti)
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{"list": tables},
	})
}

// Optimize 优化表
func (h *DatabaseHandler) Optimize(c *fiber.Ctx) error {
	var req struct {
		Tables []string `json:"tables"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "参数错误"})
	}
	if len(req.Tables) == 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "请选择要优化的表"})
	}

	// 防注入：只允许字母、数字、下划线
	for _, t := range req.Tables {
		if !isValidTableName(t) {
			return c.JSON(fiber.Map{"code": 0, "msg": "非法表名: " + t})
		}
	}

	tables := strings.Join(req.Tables, "`, `")
	sql := fmt.Sprintf("OPTIMIZE TABLE `%s`", tables)
	h.db.Exec(sql)
	return c.JSON(fiber.Map{"code": 1, "msg": "优化完成"})
}

// Repair 修复表
func (h *DatabaseHandler) Repair(c *fiber.Ctx) error {
	var req struct {
		Tables []string `json:"tables"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "参数错误"})
	}
	if len(req.Tables) == 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "请选择要修复的表"})
	}

	for _, t := range req.Tables {
		if !isValidTableName(t) {
			return c.JSON(fiber.Map{"code": 0, "msg": "非法表名: " + t})
		}
	}

	tables := strings.Join(req.Tables, "`, `")
	sql := fmt.Sprintf("REPAIR TABLE `%s`", tables)
	h.db.Exec(sql)
	return c.JSON(fiber.Map{"code": 1, "msg": "修复完成"})
}

// Backup 备份数据库（导出 SQL）
func (h *DatabaseHandler) Backup(c *fiber.Ctx) error {
	var req struct {
		Tables []string `json:"tables"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "参数错误"})
	}

	backupDir := "./runtime/backup"
	os.MkdirAll(backupDir, 0755)

	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s.sql", timestamp)
	filepath := filepath.Join(backupDir, filename)

	var sqlContent strings.Builder
	sqlContent.WriteString("-- GoCMS Database Backup\n")
	sqlContent.WriteString(fmt.Sprintf("-- Date: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	sqlContent.WriteString("SET FOREIGN_KEY_CHECKS=0;\n\n")

	// 如果未指定表，备份所有表
	if len(req.Tables) == 0 {
		rows, err := h.db.Raw("SHOW TABLES").Rows()
		if err != nil {
			return c.JSON(fiber.Map{"code": 0, "msg": "查询表失败"})
		}
		defer rows.Close()
		for rows.Next() {
			var name string
			rows.Scan(&name)
			req.Tables = append(req.Tables, name)
		}
	}

	for _, tableName := range req.Tables {
		if !isValidTableName(tableName) {
			continue
		}

		// 获取建表语句
		var createSQL string
		var tableNameResult string
		row := h.db.Raw(fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName)).Row()
		row.Scan(&tableNameResult, &createSQL)
		if createSQL != "" {
			sqlContent.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS `%s`;\n", tableName))
			sqlContent.WriteString(createSQL + ";\n\n")
		}

		// 导出数据
		rows, err := h.db.Raw(fmt.Sprintf("SELECT * FROM `%s`", tableName)).Rows()
		if err != nil {
			continue
		}
		cols, _ := rows.Columns()
		for rows.Next() {
			values := make([]interface{}, len(cols))
			valuePtrs := make([]interface{}, len(cols))
			for i := range values {
				valuePtrs[i] = &values[i]
			}
			rows.Scan(valuePtrs...)

			var vals []string
			for _, v := range values {
				if v == nil {
					vals = append(vals, "NULL")
				} else {
					switch vv := v.(type) {
					case []byte:
						vals = append(vals, fmt.Sprintf("'%s'", strings.ReplaceAll(string(vv), "'", "\\'")))
					case int64:
						vals = append(vals, strconv.FormatInt(vv, 10))
					case float64:
						vals = append(vals, strconv.FormatFloat(vv, 'f', -1, 64))
					default:
						vals = append(vals, fmt.Sprintf("'%v'", v))
					}
				}
			}
			sqlContent.WriteString(fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s);\n",
				tableName,
				"`"+strings.Join(cols, "`, `")+"`",
				strings.Join(vals, ", ")))
		}
		rows.Close()
		sqlContent.WriteString("\n")
	}

	sqlContent.WriteString("SET FOREIGN_KEY_CHECKS=1;\n")

	if err := os.WriteFile(filepath, []byte(sqlContent.String()), 0644); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "备份写入失败"})
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"msg":  "备份成功",
		"data": fiber.Map{"file": filename},
	})
}

// BackupInfo 备份文件信息
type BackupInfo struct {
	Name    string `json:"name"`
	Size    string `json:"size"`
	SizeRaw int64  `json:"size_raw"`
	Time    int64  `json:"time"`
}

// Backups 列出备份文件
func (h *DatabaseHandler) Backups(c *fiber.Ctx) error {
	backupDir := "./runtime/backup"
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return c.JSON(fiber.Map{"code": 1, "data": fiber.Map{"list": []BackupInfo{}}})
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		backups = append(backups, BackupInfo{
			Name:    entry.Name(),
			Size:    formatFileSize(info.Size()),
			SizeRaw: info.Size(),
			Time:    info.ModTime().Unix(),
		})
	}

	// 按时间倒序
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Time > backups[j].Time
	})

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{"list": backups},
	})
}

// Restore 从备份恢复
func (h *DatabaseHandler) Restore(c *fiber.Ctx) error {
	file := c.FormValue("file")
	if file == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "请选择备份文件"})
	}

	// 安全检查：防止路径穿越
	if strings.Contains(file, "..") || strings.Contains(file, "/") || strings.Contains(file, "\\") {
		return c.JSON(fiber.Map{"code": 0, "msg": "非法文件名"})
	}

	filePath := filepath.Join("./runtime/backup", file)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "备份文件不存在"})
	}

	// 按分号分割 SQL 语句并逐条执行
	statements := strings.Split(string(data), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		h.db.Exec(stmt)
	}

	return c.JSON(fiber.Map{"code": 1, "msg": "恢复完成"})
}

// SQL 执行原始 SQL（仅超级管理员）
func (h *DatabaseHandler) SQL(c *fiber.Ctx) error {
	sql := strings.TrimSpace(c.FormValue("sql"))
	if sql == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "SQL 不能为空"})
	}

	// 简单的安全检查：禁止危险操作
	upperSQL := strings.ToUpper(sql)
	dangerous := []string{"DROP DATABASE", "TRUNCATE", "GRANT", "REVOKE"}
	for _, kw := range dangerous {
		if strings.Contains(upperSQL, kw) {
			return c.JSON(fiber.Map{"code": 0, "msg": "禁止执行危险 SQL: " + kw})
		}
	}

	result := h.db.Exec(sql)
	if result.Error != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": result.Error.Error()})
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"msg":  "执行成功",
		"data": fiber.Map{"affected": result.RowsAffected},
	})
}

func formatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func isValidTableName(name string) bool {
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return len(name) > 0 && len(name) < 128
}

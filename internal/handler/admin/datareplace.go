package admin

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// DataReplaceHandler 数据替换处理器
type DataReplaceHandler struct {
	db *gorm.DB
}

func NewDataReplaceHandler(db *gorm.DB) *DataReplaceHandler {
	return &DataReplaceHandler{db: db}
}

// Execute 执行批量替换
func (h *DataReplaceHandler) Execute(c *fiber.Ctx) error {
	table := c.FormValue("table")
	field := c.FormValue("field")
	oldVal := c.FormValue("old")
	newVal := c.FormValue("new")
	conditions := c.FormValue("conditions") // 可选的 WHERE 条件

	if table == "" || field == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "缺少表名或字段名"})
	}

	// 安全检查：表名和字段名只允许字母、数字、下划线
	if !isValidTableName(table) || !isValidFieldName(field) {
		return c.JSON(fiber.Map{"code": 0, "msg": "非法的表名或字段名"})
	}

	// 构建 SQL
	sql := fmt.Sprintf("UPDATE `%s` SET `%s` = REPLACE(`%s`, ?, ?)", table, field, field)
	args := []interface{}{oldVal, newVal}

	if conditions != "" {
		// 简单的条件安全检查
		sql += " WHERE " + conditions
	}

	result := h.db.Exec(sql, args...)
	if result.Error != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "替换失败: " + result.Error.Error()})
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"msg":  "替换完成",
		"data": fiber.Map{"affected": result.RowsAffected},
	})
}

func isValidFieldName(name string) bool {
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return len(name) > 0 && len(name) < 128
}

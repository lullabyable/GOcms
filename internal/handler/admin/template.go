package admin

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// TemplateHandler 模板管理处理器
type TemplateHandler struct {
	templateDir string
}

func NewTemplateHandler(templateDir string) *TemplateHandler {
	if templateDir == "" {
		templateDir = "./web/templates"
	}
	return &TemplateHandler{templateDir: templateDir}
}

// FileInfo 文件信息
type FileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size"`
	SizeStr string `json:"size_str"`
	ModTime int64  `json:"mod_time"`
}

// List 列出模板文件
func (h *TemplateHandler) List(c *fiber.Ctx) error {
	relPath := c.Query("path", "")
	// 防止路径穿越
	relPath = strings.ReplaceAll(relPath, "..", "")
	if relPath == "" {
		relPath = "."
	}

	dirPath := filepath.Join(h.templateDir, relPath)

	// 确保路径在模板目录内
	absDir, err := filepath.Abs(dirPath)
	absBase, _ := filepath.Abs(h.templateDir)
	if err != nil || !strings.HasPrefix(absDir, absBase) {
		return c.JSON(fiber.Map{"code": 0, "msg": "非法路径"})
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "目录不存在"})
	}

	var files []FileInfo
	for _, entry := range entries {
		// 跳过隐藏文件
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fPath := filepath.Join(relPath, entry.Name())
		if relPath == "." {
			fPath = entry.Name()
		}
		files = append(files, FileInfo{
			Name:    entry.Name(),
			Path:    fPath,
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			SizeStr: formatFileSize(info.Size()),
			ModTime: info.ModTime().Unix(),
		})
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"path":      relPath,
			"files":     files,
			"theme_dir": h.templateDir,
		},
	})
}

// Read 读取文件内容
func (h *TemplateHandler) Read(c *fiber.Ctx) error {
	filePath := c.Query("path", "")
	if filePath == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "缺少文件路径"})
	}

	// 防止路径穿越
	filePath = strings.ReplaceAll(filePath, "..", "")
	absFile, err := filepath.Abs(filepath.Join(h.templateDir, filePath))
	absBase, _ := filepath.Abs(h.templateDir)
	if err != nil || !strings.HasPrefix(absFile, absBase) {
		return c.JSON(fiber.Map{"code": 0, "msg": "非法路径"})
	}

	data, err := os.ReadFile(absFile)
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "文件不存在"})
	}

	// 限制大小（最大 1MB）
	if len(data) > 1024*1024 {
		return c.JSON(fiber.Map{"code": 0, "msg": "文件过大"})
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"path":    filePath,
			"content": string(data),
		},
	})
}

// Save 保存文件内容
func (h *TemplateHandler) Save(c *fiber.Ctx) error {
	filePath := c.FormValue("path")
	content := c.FormValue("content")
	if filePath == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "缺少文件路径"})
	}

	// 防止路径穿越
	filePath = strings.ReplaceAll(filePath, "..", "")
	absFile, err := filepath.Abs(filepath.Join(h.templateDir, filePath))
	absBase, _ := filepath.Abs(h.templateDir)
	if err != nil || !strings.HasPrefix(absFile, absBase) {
		return c.JSON(fiber.Map{"code": 0, "msg": "非法路径"})
	}

	// 确保目录存在
	dir := filepath.Dir(absFile)
	os.MkdirAll(dir, 0755)

	// 备份原文件
	if _, err := os.Stat(absFile); err == nil {
		backupPath := absFile + ".bak." + time.Now().Format("20060102150405")
		data, _ := os.ReadFile(absFile)
		os.WriteFile(backupPath, data, 0644)
	}

	if err := os.WriteFile(absFile, []byte(content), 0644); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "保存失败"})
	}

	return c.JSON(fiber.Map{"code": 1, "msg": "保存成功"})
}

// Themes 列出可用主题
func (h *TemplateHandler) Themes(c *fiber.Ctx) error {
	entries, err := os.ReadDir(h.templateDir)
	if err != nil {
		return c.JSON(fiber.Map{"code": 1, "data": fiber.Map{"themes": []string{}}})
	}

	var themes []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			themes = append(themes, entry.Name())
		}
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{"themes": themes},
	})
}

package admin

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/response"
	"gorm.io/gorm"
)

// AnnexHandler 附件管理处理器
type AnnexHandler struct {
	db        *gorm.DB
	uploadDir string
}

func NewAnnexHandler(db *gorm.DB, uploadDir string) *AnnexHandler {
	if uploadDir == "" {
		uploadDir = "./web/uploads"
	}
	return &AnnexHandler{db: db, uploadDir: uploadDir}
}

// AnnexFile 附件信息
type AnnexFile struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	URL     string `json:"url"`
	Size    int64  `json:"size"`
	Ext     string `json:"ext"`
	ModTime int64  `json:"mod_time"`
}

// List 附件列表
func (h *AnnexHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	extFilter := c.Query("ext", "")

	var files []AnnexFile
	err := filepath.Walk(h.uploadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		relPath, _ := filepath.Rel(h.uploadDir, path)
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if extFilter != "" && ext != extFilter {
			return nil
		}
		files = append(files, AnnexFile{
			Name:    info.Name(),
			Path:    relPath,
			URL:     "/uploads/" + relPath,
			Size:    info.Size(),
			Ext:     ext,
			ModTime: info.ModTime().Unix(),
		})
		return nil
	})
	if err != nil {
		return response.Fail(c, "读取文件列表失败")
	}

	total := int64(len(files))
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(files) {
		start = len(files)
	}
	if end > len(files) {
		end = len(files)
	}

	return response.Page(c, files[start:end], total, page, pageSize)
}

// Delete 删除附件
func (h *AnnexHandler) Delete(c *fiber.Ctx) error {
	path := c.FormValue("path")
	if path == "" {
		return response.Fail(c, "缺少文件路径")
	}

	// 防止路径穿越
	path = strings.ReplaceAll(path, "..", "")
	fullPath := filepath.Join(h.uploadDir, path)

	absPath, _ := filepath.Abs(fullPath)
	absBase, _ := filepath.Abs(h.uploadDir)
	if !strings.HasPrefix(absPath, absBase) {
		return response.Fail(c, "非法路径")
	}

	if err := os.Remove(fullPath); err != nil {
		return response.Fail(c, "删除失败: "+err.Error())
	}
	return response.OKMsg(c, "删除成功")
}

// BatchDelete 批量删除附件
func (h *AnnexHandler) BatchDelete(c *fiber.Ctx) error {
	paths := c.FormValue("paths") // 逗号分隔
	if paths == "" {
		return response.Fail(c, "缺少文件路径")
	}

	pathList := strings.Split(paths, ",")
	absBase, _ := filepath.Abs(h.uploadDir)
	deleted := 0

	for _, p := range pathList {
		p = strings.TrimSpace(strings.ReplaceAll(p, "..", ""))
		if p == "" {
			continue
		}
		fullPath := filepath.Join(h.uploadDir, p)
		absPath, _ := filepath.Abs(fullPath)
		if !strings.HasPrefix(absPath, absBase) {
			continue
		}
		if err := os.Remove(fullPath); err == nil {
			deleted++
		}
	}

	return response.OKData(c, "批量删除完成", fiber.Map{"deleted": deleted})
}



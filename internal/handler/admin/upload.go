package admin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	uploadDir  string
	maxSize    int64
	allowedExt map[string]bool
}

func NewUploadHandler(uploadDir string, maxSize int64) *UploadHandler {
	if uploadDir == "" {
		uploadDir = "./web/uploads"
	}
	if maxSize == 0 {
		maxSize = 10 * 1024 * 1024 // 10MB
	}
	return &UploadHandler{
		uploadDir: uploadDir,
		maxSize:   maxSize,
		allowedExt: map[string]bool{
			".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
			".webp": true, ".bmp": true, ".svg": true,
			".mp4": true, ".mp3": true, ".zip": true, ".rar": true,
			".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
			".txt": true, ".json": true, ".xml": true, ".csv": true,
		},
	}
}

// File 上传文件
func (h *UploadHandler) File(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "请选择文件"})
	}

	// 大小检查
	if file.Size > h.maxSize {
		return c.JSON(fiber.Map{"code": 0, "msg": fmt.Sprintf("文件大小超过限制 (最大 %d MB)", h.maxSize/1024/1024)})
	}

	// 扩展名检查
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !h.allowedExt[ext] {
		return c.JSON(fiber.Map{"code": 0, "msg": "不支持的文件类型: " + ext})
	}

	// 生成保存路径：uploads/2006/01/filename
	dateDir := time.Now().Format("2006/01")
	saveDir := filepath.Join(h.uploadDir, dateDir)
	os.MkdirAll(saveDir, 0755)

	// 生成唯一文件名
	newName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(saveDir, newName)

	if err := c.SaveFile(file, savePath); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "上传失败"})
	}

	// 返回相对URL
	url := fmt.Sprintf("/uploads/%s/%s", dateDir, newName)

	return c.JSON(fiber.Map{
		"code": 1,
		"msg":  "上传成功",
		"data": fiber.Map{
			"url":      url,
			"name":     file.Filename,
			"new_name": newName,
			"size":     file.Size,
		},
	})
}

// Image 上传图片（仅限图片类型）
func (h *UploadHandler) Image(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "请选择图片"})
	}

	// 大小检查
	if file.Size > h.maxSize {
		return c.JSON(fiber.Map{"code": 0, "msg": fmt.Sprintf("图片大小超过限制 (最大 %d MB)", h.maxSize/1024/1024)})
	}

	// 仅限图片扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".webp": true, ".bmp": true,
	}
	if !imageExts[ext] {
		return c.JSON(fiber.Map{"code": 0, "msg": "仅支持图片格式 (jpg/png/gif/webp/bmp)"})
	}

	// 生成保存路径
	dateDir := time.Now().Format("2006/01")
	saveDir := filepath.Join(h.uploadDir, dateDir)
	os.MkdirAll(saveDir, 0755)

	newName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(saveDir, newName)

	if err := c.SaveFile(file, savePath); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "上传失败"})
	}

	url := fmt.Sprintf("/uploads/%s/%s", dateDir, newName)

	return c.JSON(fiber.Map{
		"code": 1,
		"msg":  "上传成功",
		"data": fiber.Map{
			"url":      url,
			"name":     file.Filename,
			"new_name": newName,
			"size":     file.Size,
		},
	})
}

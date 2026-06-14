package admin

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/response"
	"gorm.io/gorm"
)

// MakeHandler 静态生成处理器
type MakeHandler struct {
	db     *gorm.DB
	status *MakeStatus
	mu     sync.RWMutex
}

func NewMakeHandler(db *gorm.DB) *MakeHandler {
	return &MakeHandler{
		db: db,
		status: &MakeStatus{
			Status: "idle",
		},
	}
}

// MakeStatus 生成状态
type MakeStatus struct {
	Status    string `json:"status"` // idle, running, done, error
	Total     int    `json:"total"`
	Current   int    `json:"current"`
	Message   string `json:"message"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
}

// Start 开始静态生成
func (h *MakeHandler) Start(c *fiber.Ctx) error {
	makeType := c.FormValue("type", "all") // all, vod, art, topic, home

	h.mu.Lock()
	if h.status.Status == "running" {
		h.mu.Unlock()
		return response.Fail(c, "正在生成中，请等待完成")
	}
	h.status = &MakeStatus{
		Status:    "running",
		StartTime: time.Now().Unix(),
		Message:   "开始生成: " + makeType,
	}
	h.mu.Unlock()

	// 异步执行生成
	go func() {
		// TODO: 实际的静态页面生成逻辑
		time.Sleep(2 * time.Second) // 模拟生成过程

		h.mu.Lock()
		h.status.Status = "done"
		h.status.Total = 100
		h.status.Current = 100
		h.status.EndTime = time.Now().Unix()
		h.status.Message = "生成完成"
		h.mu.Unlock()
	}()

	return response.OKMsg(c, "已开始生成")
}

// Status 获取生成状态
func (h *MakeHandler) Status(c *fiber.Ctx) error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return response.OK(c, h.status)
}

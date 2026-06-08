package scheduler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"gocms/internal/model"
)

// TaskFunc 任务执行函数
type TaskFunc func() error

// Scheduler 定时任务调度器
type Scheduler struct {
	db       *gorm.DB
	cron     *cron.Cron
	tasks    map[string]cron.EntryID
	funcs    map[string]TaskFunc
	mu       sync.RWMutex
	running  bool
}

// TaskInfo 任务信息
type TaskInfo struct {
	TaskID    int    `json:"task_id"`
	Name      string `json:"name"`
	Schedule  string `json:"schedule"`
	Status    int    `json:"status"`
	LastRun   int64  `json:"last_run"`
	NextRun   int64  `json:"next_run"`
	RunCount  int    `json:"run_count"`
	LastError string `json:"last_error"`
	Running   bool   `json:"running"`
}

func NewScheduler(db *gorm.DB) *Scheduler {
	return &Scheduler{
		db:    db,
		cron:  cron.New(),
		tasks: make(map[string]cron.EntryID),
		funcs: make(map[string]TaskFunc),
	}
}

// Register 注册内置任务函数
func (s *Scheduler) Register(name string, fn TaskFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.funcs[name] = fn
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("调度器已在运行")
	}

	// 从数据库加载已启用的任务
	var tasks []model.Task
	if err := s.db.Where("status = ?", 1).Find(&tasks).Error; err != nil {
		return fmt.Errorf("加载任务失败: %w", err)
	}

	for _, task := range tasks {
		if err := s.addTask(task); err != nil {
			log.Printf("[SCHEDULER] 注册任务 %s 失败: %v", task.Name, err)
		}
	}

	s.cron.Start()
	s.running = true
	log.Printf("[SCHEDULER] 调度器启动，已加载 %d 个任务", len(tasks))
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.cron.Stop()
	s.running = false
	log.Println("[SCHEDULER] 调度器已停止")
}

// AddTask 添加/更新定时任务
func (s *Scheduler) AddTask(task model.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 移除旧任务
	if entryID, ok := s.tasks[fmt.Sprintf("%d", task.TaskID)]; ok {
		s.cron.Remove(entryID)
		delete(s.tasks, fmt.Sprintf("%d", task.TaskID))
	}

	if task.Status == 1 {
		return s.addTask(task)
	}
	return nil
}

func (s *Scheduler) addTask(task model.Task) error {
	entryID, err := s.cron.AddFunc(task.Schedule, func() {
		s.runTask(task)
	})
	if err != nil {
		return err
	}
	s.tasks[fmt.Sprintf("%d", task.TaskID)] = entryID

	// 更新下次运行时间
	entry := s.cron.Entry(entryID)
	if !entry.Next.IsZero() {
		s.db.Model(&model.Task{}).Where("task_id = ?", task.TaskID).
			Update("next_run", entry.Next.Unix())
	}
	return nil
}

func (s *Scheduler) runTask(task model.Task) {
	s.mu.RLock()
	fn, ok := s.funcs[task.Command]
	s.mu.RUnlock()

	start := time.Now()
	var lastErr string

	if !ok {
		lastErr = fmt.Sprintf("未注册的任务命令: %s", task.Command)
		log.Printf("[SCHEDULER] 任务 %s 执行失败: %s", task.Name, lastErr)
	} else {
		log.Printf("[SCHEDULER] 开始执行任务: %s", task.Name)
		if err := fn(); err != nil {
			lastErr = err.Error()
			log.Printf("[SCHEDULER] 任务 %s 执行出错: %v", task.Name, err)
		} else {
			log.Printf("[SCHEDULER] 任务 %s 执行完成，耗时: %v", task.Name, time.Since(start))
		}
	}

	// 更新任务状态
	s.db.Model(&model.Task{}).Where("task_id = ?", task.TaskID).Updates(map[string]interface{}{
		"last_run":   start.Unix(),
		"run_count":  gorm.Expr("run_count + 1"),
		"last_error": lastErr,
	})
}

// RemoveTask 移除任务
func (s *Scheduler) RemoveTask(taskID int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%d", taskID)
	if entryID, ok := s.tasks[key]; ok {
		s.cron.Remove(entryID)
		delete(s.tasks, key)
	}
}

// GetTaskList 获取任务列表
func (s *Scheduler) GetTaskList() ([]TaskInfo, error) {
	var tasks []model.Task
	if err := s.db.Find(&tasks).Error; err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []TaskInfo
	for _, t := range tasks {
		info := TaskInfo{
			TaskID:    t.TaskID,
			Name:      t.Name,
			Schedule:  t.Schedule,
			Status:    t.Status,
			LastRun:   t.LastRun,
			NextRun:   t.NextRun,
			RunCount:  t.RunCount,
			LastError: t.LastError,
		}

		// 获取下次运行时间
		key := fmt.Sprintf("%d", t.TaskID)
		if entryID, ok := s.tasks[key]; ok {
			entry := s.cron.Entry(entryID)
			if !entry.Next.IsZero() {
				info.NextRun = entry.Next.Unix()
			}
		}

		result = append(result, info)
	}
	return result, nil
}

// TriggerTask 手动触发任务
func (s *Scheduler) TriggerTask(taskID int) error {
	var task model.Task
	if err := s.db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		return fmt.Errorf("任务不存在")
	}

	s.mu.RLock()
	fn, ok := s.funcs[task.Command]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("未注册的任务命令: %s", task.Command)
	}

	start := time.Now()
	var lastErr string
	if err := fn(); err != nil {
		lastErr = err.Error()
	}

	s.db.Model(&model.Task{}).Where("task_id = ?", taskID).Updates(map[string]interface{}{
		"last_run":   start.Unix(),
		"run_count":  gorm.Expr("run_count + 1"),
		"last_error": lastErr,
	})

	if lastErr != "" {
		return fmt.Errorf("%s", lastErr)
	}
	return nil
}

// ToggleTask 启用/禁用任务
func (s *Scheduler) ToggleTask(taskID int, status int) error {
	var task model.Task
	if err := s.db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		return fmt.Errorf("任务不存在")
	}

	s.db.Model(&task).Update("status", status)
	task.Status = status

	return s.AddTask(task)
}

// CreateTask 创建任务
func (s *Scheduler) CreateTask(task *model.Task) error {
	task.CreatedAt = time.Now().Unix()
	task.UpdatedAt = time.Now().Unix()
	if err := s.db.Create(task).Error; err != nil {
		return err
	}
	if task.Status == 1 {
		return s.AddTask(*task)
	}
	return nil
}

// UpdateTask 更新任务
func (s *Scheduler) UpdateTask(task *model.Task) error {
	task.UpdatedAt = time.Now().Unix()
	if err := s.db.Save(task).Error; err != nil {
		return err
	}
	return s.AddTask(*task)
}

// DeleteTask 删除任务
func (s *Scheduler) DeleteTask(taskID int) error {
	s.RemoveTask(taskID)
	return s.db.Delete(&model.Task{}, taskID).Error
}

// InitBuiltinTasks 初始化内置任务
func (s *Scheduler) InitBuiltinTasks() {
	builtinTasks := []model.Task{
		{Name: "访问统计汇总", Schedule: "0 2 * * *", Command: "aggregate_daily", Status: 1},
		{Name: "缓存清理", Schedule: "0 3 * * *", Command: "cache_clean", Status: 1},
		{Name: "数据库优化", Schedule: "0 4 * * 0", Command: "db_optimize", Status: 1},
		{Name: "URL推送", Schedule: "0 6 * * *", Command: "url_push", Status: 0},
	}

	for _, bt := range builtinTasks {
		var existing model.Task
		if err := s.db.Where("command = ?", bt.Command).First(&existing).Error; err != nil {
			// 任务不存在，创建
			s.CreateTask(&bt)
		}
	}
}

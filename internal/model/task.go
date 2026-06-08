package model

// Task 定时任务模型
type Task struct {
	TaskID    int    `gorm:"primaryKey;column:task_id" json:"task_id"`
	Name      string `gorm:"column:name;size:100" json:"name"`
	Schedule  string `gorm:"column:schedule;size:50" json:"schedule"`
	Command   string `gorm:"column:command;size:200" json:"command"`
	Status    int    `gorm:"column:status" json:"status"` // 0=禁用 1=启用
	LastRun   int64  `gorm:"column:last_run" json:"last_run"`
	NextRun   int64  `gorm:"column:next_run" json:"next_run"`
	RunCount  int    `gorm:"column:run_count" json:"run_count"`
	LastError string `gorm:"column:last_error;type:text" json:"last_error"`
	CreatedAt int64  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt int64  `gorm:"column:updated_at" json:"updated_at"`
}

func (Task) TableName() string { return "mac_task" }

// URLPushLog URL推送日志模型
type URLPushLog struct {
	LogID     int    `gorm:"primaryKey;column:log_id" json:"log_id"`
	Platform  string `gorm:"column:platform;size:20" json:"platform"`
	URLs      string `gorm:"column:urls;type:text" json:"urls"`
	Total     int    `gorm:"column:total" json:"total"`
	Success   int    `gorm:"column:success" json:"success"`
	Failed    int    `gorm:"column:failed" json:"failed"`
	Error     string `gorm:"column:error;type:text" json:"error"`
	CreatedAt int64  `gorm:"column:created_at" json:"created_at"`
}

func (URLPushLog) TableName() string { return "mac_url_push_log" }

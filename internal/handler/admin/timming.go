package admin

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/model"
	"gocms/internal/service/scheduler"
)

// TimmingHandler 定时任务管理处理器
type TimmingHandler struct {
	scheduler *scheduler.Scheduler
}

func NewTimmingHandler(s *scheduler.Scheduler) *TimmingHandler {
	return &TimmingHandler{scheduler: s}
}

// List 任务列表
func (h *TimmingHandler) List(c *fiber.Ctx) error {
	tasks, err := h.scheduler.GetTaskList()
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "获取任务列表失败"})
	}
	return c.JSON(fiber.Map{"code": 1, "data": tasks})
}

// Create 创建任务
func (h *TimmingHandler) Create(c *fiber.Ctx) error {
	name := c.FormValue("name")
	schedule := c.FormValue("schedule")
	command := c.FormValue("command")
	status, _ := strconv.Atoi(c.FormValue("status", "1"))

	if name == "" || schedule == "" || command == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "任务名称、调度表达式和命令不能为空"})
	}

	task := &model.Task{
		Name:     name,
		Schedule: schedule,
		Command:  command,
		Status:   status,
	}

	if err := h.scheduler.CreateTask(task); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "创建失败: " + err.Error()})
	}

	return c.JSON(fiber.Map{"code": 1, "msg": "创建成功", "data": task})
}

// Update 更新任务
func (h *TimmingHandler) Update(c *fiber.Ctx) error {
	taskID, _ := strconv.Atoi(c.FormValue("task_id"))
	if taskID == 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "任务ID不能为空"})
	}

	task := &model.Task{
		TaskID:   taskID,
		Name:     c.FormValue("name"),
		Schedule: c.FormValue("schedule"),
		Command:  c.FormValue("command"),
		Status:   1,
	}

	if s, err := strconv.Atoi(c.FormValue("status")); err == nil {
		task.Status = s
	}

	if err := h.scheduler.UpdateTask(task); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "更新失败: " + err.Error()})
	}

	return c.JSON(fiber.Map{"code": 1, "msg": "更新成功"})
}

// Delete 删除任务
func (h *TimmingHandler) Delete(c *fiber.Ctx) error {
	taskID, _ := strconv.Atoi(c.Params("id"))
	if taskID == 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "任务ID不能为空"})
	}

	if err := h.scheduler.DeleteTask(taskID); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "删除失败: " + err.Error()})
	}

	return c.JSON(fiber.Map{"code": 1, "msg": "删除成功"})
}

// Toggle 启用/禁用任务
func (h *TimmingHandler) Toggle(c *fiber.Ctx) error {
	taskID, _ := strconv.Atoi(c.Params("id"))
	status, _ := strconv.Atoi(c.FormValue("status", "0"))

	if taskID == 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "任务ID不能为空"})
	}

	if err := h.scheduler.ToggleTask(taskID, status); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "操作失败: " + err.Error()})
	}

	return c.JSON(fiber.Map{"code": 1, "msg": "操作成功"})
}

// Trigger 手动触发任务
func (h *TimmingHandler) Trigger(c *fiber.Ctx) error {
	taskID, _ := strconv.Atoi(c.Params("id"))
	if taskID == 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "任务ID不能为空"})
	}

	if err := h.scheduler.TriggerTask(taskID); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "执行失败: " + err.Error()})
	}

	return c.JSON(fiber.Map{"code": 1, "msg": "任务已触发"})
}

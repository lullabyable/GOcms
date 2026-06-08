package scheduler_test

import (
	"testing"

	"gocms/internal/testutil"
	"gocms/internal/model"
	"gocms/internal/service/scheduler"
)

func TestCreateTask(t *testing.T) {
	db := testutil.TestDB(t)
	s := scheduler.NewScheduler(db)

	task := &model.Task{
		Name:     "测试任务",
		Schedule: "*/5 * * * *",
		Command:  "test_cmd",
		Status:   1,
	}

	if err := s.CreateTask(task); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if task.TaskID == 0 {
		t.Error("task ID should be set after create")
	}
}

func TestGetTaskList(t *testing.T) {
	db := testutil.TestDB(t)
	s := scheduler.NewScheduler(db)

	s.CreateTask(&model.Task{Name: "A", Schedule: "* * * * *", Command: "a", Status: 1})
	s.CreateTask(&model.Task{Name: "B", Schedule: "* * * * *", Command: "b", Status: 0})

	tasks, err := s.GetTaskList()
	if err != nil {
		t.Fatalf("GetTaskList failed: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestToggleTask(t *testing.T) {
	db := testutil.TestDB(t)
	s := scheduler.NewScheduler(db)

	s.Register("test_cmd", func() error { return nil })

	task := &model.Task{Name: "Toggle测试", Schedule: "* * * * *", Command: "test_cmd", Status: 1}
	s.CreateTask(task)

	// 禁用
	if err := s.ToggleTask(task.TaskID, 0); err != nil {
		t.Fatalf("ToggleTask disable failed: %v", err)
	}

	var updated model.Task
	db.First(&updated, task.TaskID)
	if updated.Status != 0 {
		t.Errorf("expected status=0, got %d", updated.Status)
	}

	// 启用
	if err := s.ToggleTask(task.TaskID, 1); err != nil {
		t.Fatalf("ToggleTask enable failed: %v", err)
	}
}

func TestDeleteTask(t *testing.T) {
	db := testutil.TestDB(t)
	s := scheduler.NewScheduler(db)

	task := &model.Task{Name: "Delete测试", Schedule: "* * * * *", Command: "del", Status: 0}
	s.CreateTask(task)

	if err := s.DeleteTask(task.TaskID); err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}

	var count int64
	db.Model(&model.Task{}).Where("task_id = ?", task.TaskID).Count(&count)
	if count != 0 {
		t.Error("task should be deleted")
	}
}

func TestTriggerTask(t *testing.T) {
	db := testutil.TestDB(t)
	s := scheduler.NewScheduler(db)

	executed := false
	s.Register("trigger_test", func() error {
		executed = true
		return nil
	})

	task := &model.Task{Name: "Trigger测试", Schedule: "* * * * *", Command: "trigger_test", Status: 1}
	s.CreateTask(task)

	if err := s.TriggerTask(task.TaskID); err != nil {
		t.Fatalf("TriggerTask failed: %v", err)
	}

	if !executed {
		t.Error("task should have been executed")
	}

	// 验证运行次数
	var updated model.Task
	db.First(&updated, task.TaskID)
	if updated.RunCount != 1 {
		t.Errorf("expected run_count=1, got %d", updated.RunCount)
	}
}

func TestTriggerTaskNotRegistered(t *testing.T) {
	db := testutil.TestDB(t)
	s := scheduler.NewScheduler(db)

	task := &model.Task{Name: "NoCmd", Schedule: "* * * * *", Command: "nonexistent", Status: 1}
	s.CreateTask(task)

	err := s.TriggerTask(task.TaskID)
	if err == nil {
		t.Error("should fail for unregistered command")
	}
}

func TestUpdateTask(t *testing.T) {
	db := testutil.TestDB(t)
	s := scheduler.NewScheduler(db)

	task := &model.Task{Name: "Original", Schedule: "* * * * *", Command: "cmd", Status: 1}
	s.CreateTask(task)

	task.Name = "Updated"
	task.Schedule = "0 * * * *"
	if err := s.UpdateTask(task); err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	var updated model.Task
	db.First(&updated, task.TaskID)
	if updated.Name != "Updated" {
		t.Errorf("expected Updated, got %s", updated.Name)
	}
}

func TestStartStop(t *testing.T) {
	db := testutil.TestDB(t)
	s := scheduler.NewScheduler(db)

	s.Register("noop", func() error { return nil })

	if err := s.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// 重复启动应失败
	if err := s.Start(); err == nil {
		t.Error("double start should fail")
	}

	s.Stop()

	// 停止后应能重新启动
	if err := s.Start(); err != nil {
		t.Fatalf("Restart failed: %v", err)
	}
	s.Stop()
}

func TestInitBuiltinTasks(t *testing.T) {
	db := testutil.TestDB(t)
	s := scheduler.NewScheduler(db)

	s.InitBuiltinTasks()

	var count int64
	db.Model(&model.Task{}).Count(&count)
	if count < 4 {
		t.Errorf("expected at least 4 builtin tasks, got %d", count)
	}
}

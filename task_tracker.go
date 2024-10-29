package main

import (
	"fmt"
	"log/slog"
	"time"
)

type Task struct {
	ID              int       `db:"id"`
	StartAt         time.Time `db:"start_at"`
	FinishAt        time.Time `db:"finish_at"`
	SecondsDuration int       `db:"duration"`

	// don't save to db; need for app logic
	lastStartAt time.Time
	finished    bool
}

type TaskManager struct {
	storage       *Storage
	logger        *slog.Logger
	currentTask   *Task
	stateTaskChan chan StateChangeEvent
}

func NewTaskManager(logger *slog.Logger, storage *Storage, stateTaskChan chan StateChangeEvent) *TaskManager {
	return &TaskManager{
		storage:       storage,
		logger:        logger,
		stateTaskChan: stateTaskChan,
		currentTask:   nil,
	}
}

func (tm *TaskManager) HandleTaskStateChanges() {
	for event := range tm.stateTaskChan {
		if event.TimerType == FocusTimer {
			switch event.NewState {
			case StateActive:
				tm.handleStartTask()
			case StatePaused:
				tm.handlePauseTask()
			case StateFinished:
				tm.handleFinishTask()
			}
		}
	}
}

func (tm *TaskManager) Tasks() ([]*Task, error) {
	return tm.storage.GetTasks()
}

func (tm *TaskManager) TodayTasks() ([]*Task, error) {
	return tm.storage.GetTodayTasks()
}

func (tm *TaskManager) RemoveTask(id int) error {
	return tm.storage.RemoveTask(id)
}

func (tm *TaskManager) CreateNewTask(startAt time.Time, finishAt time.Time, duration int) (*Task, error) {
	task := &Task{
		ID:              0,
		StartAt:         startAt,
		FinishAt:        finishAt,
		SecondsDuration: duration,
		lastStartAt:     time.Now(),
		finished:        false,
	}

	err := tm.storage.CreateTask(task)
	if err != nil {
		return nil, fmt.Errorf("can't create task: %w", err)
	}
	return task, nil
}

func (tm *TaskManager) Hours(tasks []*Task) float64 {
	var result time.Duration
	for _, task := range tasks {
		result += time.Duration(task.SecondsDuration) * time.Second
	}
	return result.Hours()
}

func (tm *TaskManager) CountDays(tasks []*Task) int {
	if len(tasks) == 0 {
		return 0
	}

	count := 1
	for i := 1; i < len(tasks); i++ {
		prev := tasks[i-1]
		cur := tasks[i]

		if prev.StartAt.Day() != cur.StartAt.Day() {
			count++
		}
	}

	return count
}

func (tm *TaskManager) HoursInWeek(tasks []*Task) [7]int {
	if len(tasks) == 0 {
		return [7]int{}
	}

	var weekdayHours [7]int
	firstDayIndx := time.Now().Day() - int(time.Now().Weekday()) + 1
	lastDayIndx := time.Now().Day()
	for _, task := range tasks {
		if task.StartAt.Day() >= firstDayIndx && task.StartAt.Day() <= lastDayIndx {
			weekdayHours[(task.FinishAt.Weekday()+6)%7] += task.SecondsDuration / 3600
		}
	}

	return weekdayHours
}

func (tm *TaskManager) FinishRunningTask() {
	if tm.currentTask != nil && !tm.currentTask.finished {
		tm.handleFinishTask()
	}
}

func (tm *TaskManager) handleStartTask() {
	// предыдущая задача не создана или завершена
	// создаём новую пустую задачу
	if tm.currentTask == nil || tm.currentTask.finished {
		newTask, err := tm.CreateNewTask(time.Now(), time.Now(), 0)
		if err != nil {
			tm.logger.Error("handle start task", slog.Any("error", err))
		}
		tm.currentTask = newTask
		return
	}

	// есть текущая незавершённая задача (запуск после паузы)
	tm.currentTask.lastStartAt = time.Now()
}

func (tm *TaskManager) handlePauseTask() {
	err := tm.updateCurrentTaskDuration()
	if err != nil {
		tm.logger.Error("handle start task", slog.Any("error", err))
	}
}

func (tm *TaskManager) handleFinishTask() {
	tm.currentTask.finished = true
	err := tm.updateCurrentTaskDuration()
	if err != nil {
		tm.logger.Error("handle start task", slog.Any("error", err))
	}
}

func (tm *TaskManager) updateCurrentTaskDuration() error {
	duration := int(time.Since(tm.currentTask.lastStartAt).Seconds())

	tm.currentTask.SecondsDuration += duration
	tm.currentTask.FinishAt = time.Now()

	err := tm.storage.UpdateTask(tm.currentTask)
	if err != nil {
		return fmt.Errorf("can't update task: %w", err)
	}
	return nil
}

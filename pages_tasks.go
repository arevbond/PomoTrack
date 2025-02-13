package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (m *UIManager) NewTasksPage() *Page {
	renderFunc := m.renderAllTasksGrid()
	return NewPageComponent(allTasksPage, true, renderFunc)
}

func (m *UIManager) NewTaskCreationPage() *Page {
	renderFunc := m.renderCreationTaskGrid()
	return NewPageComponent(addNewTaskPage, true, renderFunc)
}

func (m *UIManager) NewTaskDeletionPage() *Page {
	return NewPageComponent(deleteTaskPage, true, m.renderDeletionTaskGrid())
}

func (m *UIManager) renderAllTasksGrid() func() tview.Primitive {
	tasks, err := m.taskTracker.Tasks()
	if err != nil {
		m.logger.Error("can't get tasks", slog.String("func", "renderAllTasls"), slog.Any("error", err))
		return nil
	}
	return func() tview.Primitive {
		return m.allTasksGrid(tasks)
	}
}

func (m *UIManager) allTasksGrid(tasks []*Task) tview.Primitive {
	list := tview.NewList()
	shortCut := '0'
	for _, task := range tasks {
		name := task.Name
		if task.IsActive {
			name = fmt.Sprintf("[::bu]%s[-]", name)
		}
		if task.IsComplete {
			name = fmt.Sprintf("[gray]%s[-]", name)
		}
		list = list.AddItem(name, fmt.Sprintf("%d/%d", task.PomodorosCompleted, task.PomodorosRequired),
			shortCut, m.changeActiveTask(task))
		shortCut++
	}

	grid := tview.NewGrid().
		SetRows(0).
		SetColumns(0, 46, 0)

	grid.AddItem(list, 0, 1, 1, 1, 0, 0, true)
	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlA:
			m.AddPageAndSwitch(m.NewTaskCreationPage())
		case tcell.KeyCtrlD:
			m.AddPageAndSwitch(m.NewTaskDeletionPage())
		}
		return event
	})

	return grid
}

func (m *UIManager) changeActiveTask(task *Task) func() {
	return func() {
		prevActiveTask, err := m.taskTracker.ActiveTask()
		if err != nil {
			m.logger.Error("eror in change active task", slog.Any("error", err))
		}
		// TODO: add transaction
		if prevActiveTask != nil {
			prevActiveTask.IsActive = false
			err = m.taskTracker.UpdateTask(prevActiveTask)
			if err != nil {
				m.logger.Error("can't update field is_active in prev active task", slog.Any("error", err))
				return
			}
			if prevActiveTask.ID == task.ID {
				m.AddPageAndSwitch(m.NewTasksPage())
				return
			}
		}
		task.IsActive = true
		err = m.taskTracker.UpdateTask(task)
		if err != nil {
			m.logger.Error("can't update field is_active in task", slog.Any("error", err))
			return
		}
		m.AddPageAndSwitch(m.NewPausePage(FocusTimer))
		return
	}
}

func (m *UIManager) renderCreationTaskGrid() func() tview.Primitive {
	tasks, err := m.taskTracker.Tasks()
	if err != nil {
		m.logger.Error("can't get tasks", slog.String("func", "renderAllTasls"), slog.Any("error", err))
		return nil
	}
	return func() tview.Primitive {
		form := tview.NewForm().
			AddInputField("Task name", "", 15, nil, nil).
			AddInputField("Pomodoros", "1", 3, tview.InputFieldInteger, nil)
		form.AddButton("Save", m.saveNewTask(form, len(tasks) == 0))

		taskList := m.allTasksGrid(tasks)

		grid := tview.NewGrid().
			SetRows(0, 0).
			SetColumns(0, 46, 0).
			SetBorders(true)

		grid.AddItem(form, 0, 1, 1, 1, 0, 0, true)
		grid.AddItem(taskList, 1, 1, 1, 1, 0, 0, false)
		return grid
	}
}

func (m *UIManager) renderDeletionTaskGrid() func() tview.Primitive {
	tasks, err := m.taskTracker.Tasks()
	if err != nil {
		m.logger.Error("can't get tasks", slog.String("func", "renderAllTasls"), slog.Any("error", err))
		return nil
	}
	return func() tview.Primitive {
		form := tview.NewForm().
			AddInputField("Remove task with index", "", 3, tview.InputFieldInteger, nil)
		form.AddButton("Save", func() {
			indxStr := form.GetFormItem(0).(*tview.InputField).GetText()
			indx, err := strconv.Atoi(indxStr)
			if err != nil {
				m.logger.Error("can't get task index to delete", slog.Any("error", err))
				m.AddPageAndSwitch(m.NewTasksPage())
				return
			}
			if indx < 0 || indx >= len(tasks) {
				m.AddPageAndSwitch(m.NewTasksPage())
				return
			}
			err = m.taskTracker.DeleteTask(tasks[indx].ID)
			if err != nil {
				m.logger.Error("can't delete task", slog.Int("index", indx), slog.Any("error", err))
				m.AddPageAndSwitch(m.NewTasksPage())
				return
			}
			m.AddPageAndSwitch(m.NewTasksPage())
			return
		})
		form.AddButton("Cancel", func() {
			m.AddPageAndSwitch(m.NewTasksPage())
		})

		taskList := m.allTasksGrid(tasks)

		grid := tview.NewGrid().
			SetRows(0, 0).
			SetColumns(0, 46, 0).
			SetBorders(true)

		grid.AddItem(form, 0, 1, 1, 1, 0, 0, true)
		grid.AddItem(taskList, 1, 1, 1, 1, 0, 0, false)
		return grid
	}
}

func (m *UIManager) saveNewTask(form *tview.Form, isFirstTask bool) func() {
	return func() {
		taskName := form.GetFormItem(0).(*tview.InputField).GetText()
		pomodorosRequiredStr := form.GetFormItem(1).(*tview.InputField).GetText()

		pomodorosRequired, err := strconv.Atoi(pomodorosRequiredStr)
		if err != nil {
			m.logger.Error("can't convert pomodoros required string to int", slog.String("value", pomodorosRequiredStr))
			return
		}

		if pomodorosRequired < 0 {
			pomodorosRequired = 0
		}

		err = m.taskTracker.CreateTask(&Task{
			Name:              taskName,
			PomodorosRequired: pomodorosRequired,
			IsActive:          isFirstTask,
		})
		if err != nil {
			m.logger.Error("can't create new task in form", slog.Any("error", err))
			return
		}
		m.AddPageAndSwitch(m.NewTasksPage())
	}
}

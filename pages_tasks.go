package main

import (
	"fmt"
	"log/slog"

	"github.com/rivo/tview"
)

func (m *UIManager) NewTasksPage() *Page {
	tasks, err := m.taskTracker.Tasks()
	if err != nil {
		m.logger.Error("can't get tasks", slog.String("func", "renderAllTasls"), slog.Any("error", err))
		return nil
	}
	renderFunc := m.renderAllTasksPage(tasks)
	return NewPageComponent(allTasksPage, true, false, renderFunc)
}

func (m *UIManager) renderAllTasksPage(args ...any) func() tview.Primitive {
	tasks := args[0].([]*Task)
	return func() tview.Primitive {
		list := tview.NewList()
		for _, task := range tasks {
			list = list.AddItem(task.Name, fmt.Sprintf("%d/%d", task.PomodorosCompleted, task.PomodorosRequired),
				'a', nil)
		}

		grid := tview.NewGrid().
			SetRows(0).
			SetColumns(0, 23, 23, 0)

		grid.AddItem(list, 0, 1, 1, 2, 0, 0, true)

		return grid
	}
}

package main

import (
	"fmt"
	"log/slog"

	"github.com/rivo/tview"
)

func (m *UIManager) NewTasksPage() *Page {
	renderFunc := m.renderAllTasksGrid()
	return NewPageComponent(allTasksPage, true, renderFunc)
}

func (m *UIManager) renderAllTasksGrid() func() tview.Primitive {
	tasks, err := m.taskTracker.Tasks()
	if err != nil {
		m.logger.Error("can't get tasks", slog.String("func", "renderAllTasls"), slog.Any("error", err))
		return nil
	}
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

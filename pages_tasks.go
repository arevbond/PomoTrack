package main

import (
	"fmt"

	"github.com/rivo/tview"
)

const (
	allTasksPage = "All-Tasks-Pages"
)

func (m *UIManager) renderAllTasksPage(tasks []*Task) *PageComponent {
	list := tview.NewList()
	for _, task := range tasks {
		list = list.AddItem(task.Name, fmt.Sprintf("%d/%d", task.PomodorosCompleted, task.PomodorosRequired),
			'a', nil)
	}

	grid := tview.NewGrid().
		SetRows(0).
		SetColumns(0, 23, 23, 0)

	grid.AddItem(list, 0, 1, 1, 2, 0, 0, true)

	return NewPageComponent(allTasksPage, grid, true, false)
}

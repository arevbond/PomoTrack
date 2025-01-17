package main

import (
	"fmt"

	"github.com/rivo/tview"
)

const (
	allTasksPage = "All-Tasks-Pages"
)

func (m *UIManager) renderAllTasksPage(tasks []*Task) {
	list := tview.NewList()
	for _, task := range tasks {
		list = list.AddItem(task.Name, fmt.Sprintf("%d/%d", task.PomodorosCompleted, task.PomodorosRequired),
			'a', nil)
	}

	grid := tview.NewGrid().
		SetRows(20, 0).
		SetColumns(0, 46, 0).
		SetBorders(true)

	grid.AddItem(list, 0, 1, 1, 1, 0, 0, true)

	m.pages.AddPage(allTasksPage, grid, true, false)
}

package main

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (uim *UIManager) pagePause(pageName string, title string, timerType TimerType) {
	pauseText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(title)

	durationText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(uim.stateManager.timeToFinish(timerType).String())

	startButton := tview.NewButton("Start").SetSelectedFunc(func() {
		uim.stateManager.SetState(StateActive, timerType)
	})

	grid := tview.NewGrid().
		SetRows(0, 3, 3, 1, 0).
		SetColumns(0, 30, 0).
		SetBorders(true)

	grid.AddItem(pauseText, 1, 1, 1, 1, 0, 0, false)
	grid.AddItem(durationText, 2, 1, 1, 1, 0, 0, false)
	grid.AddItem(startButton, 3, 1, 1, 1, 0, 0, true)

	uim.pages.AddPage(pageName, grid, true, true)
}

func (uim *UIManager) pageActiveTimer(pageName string, color, title string, timerType TimerType) {
	breakText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("[%s]%s", color, title))

	timerText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
			tview.Print(screen, fmt.Sprintf("[%s]%s[-]", color,
				uim.stateManager.timeToFinish(timerType).String()),
				x, y+height/4, width, tview.AlignCenter, tcell.ColorLime)
			return 0, 0, 0, 0
		})

	pauseButton := tview.NewButton("Pause").SetSelectedFunc(func() {
		uim.stateManager.SetState(StatePaused, timerType)
	})

	toggleButton := tview.NewButton("→").SetSelectedFunc(func() {
		uim.stateManager.SetState(StateFinished, timerType)
	})

	grid := tview.NewGrid().
		SetRows(0, 3, 3, 1, 0).
		SetColumns(0, 25, 5, 0).
		SetBorders(true)

	grid.AddItem(breakText, 1, 1, 1, 2, 0, 0, false)
	grid.AddItem(timerText, 2, 1, 1, 2, 0, 0, false)
	grid.AddItem(pauseButton, 3, 1, 1, 1, 0, 0, true)
	grid.AddItem(toggleButton, 3, 2, 1, 1, 0, 0, false)

	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB, tcell.KeyLeft, tcell.KeyRight:
			if uim.ui.GetFocus() == pauseButton {
				uim.ui.SetFocus(toggleButton)
			} else {
				uim.ui.SetFocus(pauseButton)
			}
		default:
			return event
		}
		return nil
	})

	uim.pages.AddPage(pageName, grid, true, false)
}

func (uim *UIManager) pageStatistics(start, end int, tasks []*Task) {
	table := uim.statisticsTable(start, end, tasks)

	buttons := uim.createStatisticsButtons(start, end, tasks)

	text := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("Total: %s", uim.totalDuration(tasks))).
		SetDynamicColors(true)

	grid := tview.NewGrid().
		SetRows(3, 0, 1, 1).
		SetColumns(0, 25, 25, 0).
		SetBorders(true)

	grid.AddItem(text, 0, 1, 1, 2, 0, 0, false)

	grid.AddItem(table, 1, 1, 1, 2, 0, 0, true)

	for i, button := range buttons {
		grid.AddItem(button, 2, i+1, 1, 1, 0, 0, true)
	}

	grid.SetInputCapture(uim.inputCapturePageStats(table, buttons))

	uim.pages.AddPage(pageNameStatistics, grid, true, false)
}

func (uim *UIManager) inputCapturePageStats(
	table *tview.Table,
	buttons []*tview.Button) func(*tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:
			if uim.anyButtonFocused(buttons) {
				uim.ui.SetFocus(table)
			} else {
				uim.ui.SetFocus(buttons[0])
			}
		case tcell.KeyLeft, tcell.KeyRight:
			prevIndx := indexOf(uim.ui.GetFocus(), buttons)
			newIndx := (prevIndx + 1) % len(buttons)
			uim.ui.SetFocus(buttons[newIndx])
		default:
		}
		return event
	}
}

func (uim *UIManager) anyButtonFocused(buttons []*tview.Button) bool {
	focusedPrimitive := uim.ui.GetFocus()
	for _, button := range buttons {
		if button == focusedPrimitive {
			return true
		}
	}
	return false
}

func indexOf(targetButton tview.Primitive, buttons []*tview.Button) int {
	for i, button := range buttons {
		if targetButton == button {
			return i
		}
	}
	return -1
}

func (uim *UIManager) createStatisticsButtons(start, end int, tasks []*Task) []*tview.Button {
	buttons := make([]*tview.Button, 0)

	const pageSize = 5
	if start >= pageSize {
		prevButton := tview.NewButton("Prev").SetSelectedFunc(func() {
			uim.pageStatistics(start-pageSize, start, tasks)
			uim.pages.SwitchToPage(pageNameStatistics)
		})

		buttons = append(buttons, prevButton)
	}
	if end < len(tasks) {
		nextButton := tview.NewButton("Next").SetSelectedFunc(func() {
			if end+5 < len(tasks) {
				uim.pageStatistics(end, end+pageSize, tasks)
			} else {
				uim.pageStatistics(end, len(tasks), tasks)
			}
			uim.pages.SwitchToPage(pageNameStatistics)
		})

		buttons = append(buttons, nextButton)
	}

	return buttons
}

func (uim *UIManager) statisticsTable(start, end int, tasks []*Task) *tview.Table {
	table := tview.NewTable().SetBorders(true)
	headers := []string{"Date", "Time", "Seconds", "Action"}

	for col, header := range headers {
		table.SetCell(0, col, tview.NewTableCell(header).SetAlign(tview.AlignCenter))
	}

	for i, t := range tasks[start:end] {
		row := i + 1

		dateStr := t.StartAt.Format("02-Jan-2006")
		const timeFormat = "15:04:05"
		timeStr := fmt.Sprintf("%s-%s", t.StartAt.Format(timeFormat), t.FinishAt.Format(timeFormat))

		table.SetCell(row, 0, tview.NewTableCell(dateStr).SetAlign(tview.AlignCenter))
		table.SetCell(row, 1, tview.NewTableCell(timeStr).SetAlign(tview.AlignCenter))
		table.SetCell(row, 2, tview.NewTableCell(strconv.Itoa(t.Duration)).SetAlign(tview.AlignCenter))
		table.SetCell(row, 3, tview.NewTableCell("[red] Delete [-]").SetAlign(tview.AlignCenter).SetSelectable(true))
	}

	table.SetInputCapture(uim.createTableInputCapture(table, tasks))
	return table
}

func (uim *UIManager) createTableInputCapture(table *tview.Table, tasks []*Task) func(*tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		row, col := table.GetSelection()

		switch event.Key() {
		case tcell.KeyEnter:
			uim.handleEnterKey(table, tasks, col)
		case tcell.KeyDown, tcell.KeyUp:
			uim.handleVerticalNavigation(table, row, col, event.Key())
		case tcell.KeyLeft, tcell.KeyRight:
			if col != 3 {
				table.Select(row, 3)
			}
		case tcell.KeyCtrlY:
			if col == 3 && row > 0 {
				uim.removeTask(tasks, row-1)
			}
		case tcell.KeyEscape:
			table.Select(0, 0).SetSelectable(false, false)
		default:
			return event
		}

		return nil
	}
}

func (uim *UIManager) handleEnterKey(table *tview.Table, tasks []*Task, col int) {
	if len(tasks) > 0 && col != 3 {
		table.Select(1, 3).SetSelectable(true, true)
	} else if col == 3 {
		table.Select(0, 0).SetSelectable(false, false)
	}
}

func (uim *UIManager) handleVerticalNavigation(table *tview.Table, row, col int, key tcell.Key) {
	switch key {
	case tcell.KeyDown:
		if row < table.GetRowCount()-1 {
			table.Select(row+1, col)
		} else {
			table.Select(1, col)
		}
	case tcell.KeyUp:
		if row > 1 {
			table.Select(row-1, col)
		} else {
			table.Select(table.GetRowCount()-1, col)
		}
	default:
	}
}

func (uim *UIManager) removeTask(tasks []*Task, indx int) {
	err := uim.taskManager.RemoveTask(tasks[indx].ID)
	if err != nil {
		uim.logger.Error("Can't delete task", slog.Any("id", tasks[indx].ID))
	}

	tasks = append(tasks[:indx], tasks[indx+1:]...)
	if len(tasks) > 5 {
		uim.pageStatistics(0, 5, tasks)
	} else {
		uim.pageStatistics(0, len(tasks), tasks)
	}
	uim.pages.SwitchToPage(pageNameStatistics)
}

func (uim *UIManager) totalDuration(tasks []*Task) string {
	var total int
	for _, t := range tasks {
		total += t.Duration
	}
	res := time.Duration(total) * time.Second
	return res.String()
}

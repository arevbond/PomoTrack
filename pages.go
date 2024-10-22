package main

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const statisticsPageSize = 6

func (uim *UIManager) pageStatistics(start, end int, tasks []*Task) {
	table := uim.statisticsTable(start, end, tasks)

	buttons := uim.createStatisticsButtons(start, end, tasks)

	text := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("Total: %s", uim.totalDuration(tasks))).
		SetDynamicColors(true)

	grid := tview.NewGrid().
		SetRows(1, 0, 1, 1).
		//SetRows(1, 15, 1, 1).
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

	if start >= statisticsPageSize {
		prevButton := tview.NewButton("Prev").SetSelectedFunc(func() {
			uim.pageStatistics(start-statisticsPageSize, start, tasks)
			uim.pages.SwitchToPage(pageNameStatistics)
		})

		buttons = append(buttons, prevButton)
	}
	if end < len(tasks) {
		nextButton := tview.NewButton("Next").SetSelectedFunc(func() {
			if end+5 < len(tasks) {
				uim.pageStatistics(end, end+statisticsPageSize, tasks)
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

func (uim *UIManager) removeTgiask(tasks []*Task, indx int) {
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

func (uim *UIManager) pageInsertStatistics(start, end int, tasks []*Task) {
	table := uim.statisticsTable(start, end, tasks)

	buttons := uim.createStatisticsButtons(start, end, tasks)

	text := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("Total: %s", uim.totalDuration(tasks))).
		SetDynamicColors(true)

	form1 := tview.NewForm().SetHorizontal(true).
		AddInputField("Date", time.Now().Format("02-01-06"), 10, func(textToCheck string, lastChar rune) bool {
			return true
		}, nil).
		AddInputField("Time start", time.Now().Format("15-04"), 10, func(textToCheck string, lastChar rune) bool {
			return true
		}, nil)

	form2 := tview.NewForm().SetHorizontal(true).
		AddInputField("Minutes", "", 10, func(textToCheck string, lastChar rune) bool {
			return true
		}, nil).
		AddButton("Save", nil).AddButton("Cancel", nil)

	grid := tview.NewGrid().
		SetRows(1, 3, 3, 0, 1, 1).
		SetColumns(0, 25, 25, 0).
		SetBorders(true)

	grid.AddItem(text, 0, 1, 1, 2, 0, 0, false)

	grid.AddItem(form1, 1, 1, 1, 2, 0, 0, true)
	grid.AddItem(form2, 2, 1, 1, 2, 0, 0, true)

	grid.AddItem(table, 3, 1, 1, 2, 0, 0, false)

	for i, button := range buttons {
		grid.AddItem(button, 4, i+1, 1, 1, 0, 0, false)
	}

	//grid.SetInputCapture(uim.inputCapturePageStats(table, buttons))

	uim.pages.AddPage(pageNameInsertStatistics, grid, true, false)
}

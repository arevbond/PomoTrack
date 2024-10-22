package main

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	statsPage       = "Statistics"
	insertStatsPage = "Insert-Statistics"
)

const statisticsPageSize = 6

func (uim *UIManager) renderStatsPage(start, end int, tasks []*Task) {
	table := uim.newStatsTable(start, end, tasks)

	buttons := uim.newStatsButtons(start, end, tasks)

	text := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("Total: %s", uim.totalDuration(tasks))).
		SetDynamicColors(true)

	grid := tview.NewGrid().
		SetRows(1, 0, 1, 1).
		SetColumns(0, 25, 25, 0).
		SetBorders(true)

	grid.AddItem(text, 0, 1, 1, 2, 0, 0, false)

	grid.AddItem(table, 1, 1, 1, 2, 0, 0, true)

	for i, button := range buttons {
		grid.AddItem(button, 2, i+1, 1, 1, 0, 0, true)
	}

	grid.SetInputCapture(uim.captureStatsInput(table, buttons))

	uim.pages.AddPage(statsPage, grid, true, false)
}

func (uim *UIManager) captureStatsInput(
	table *tview.Table,
	buttons []*tview.Button) func(*tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:
			if uim.isButtonFocused(buttons) {
				uim.ui.SetFocus(table)
			} else {
				uim.ui.SetFocus(buttons[0])
			}
		case tcell.KeyLeft, tcell.KeyRight:
			if len(buttons) > 0 {
				prevIndx := buttonIndex(uim.ui.GetFocus(), buttons)
				newIndx := (prevIndx + 1) % len(buttons)
				uim.ui.SetFocus(buttons[newIndx])
			}
		default:
		}
		return event
	}
}

func (uim *UIManager) isButtonFocused(buttons []*tview.Button) bool {
	focusedPrimitive := uim.ui.GetFocus()
	for _, button := range buttons {
		if button == focusedPrimitive {
			return true
		}
	}
	return false
}

func buttonIndex(targetButton tview.Primitive, buttons []*tview.Button) int {
	for i, button := range buttons {
		if targetButton == button {
			return i
		}
	}
	return -1
}

func (uim *UIManager) newStatsButtons(start, end int, tasks []*Task) []*tview.Button {
	buttons := make([]*tview.Button, 0)

	if start >= statisticsPageSize {
		prevButton := tview.NewButton("Prev").SetSelectedFunc(func() {
			uim.renderStatsPage(start-statisticsPageSize, start, tasks)
			uim.pages.SwitchToPage(statsPage)
		})

		buttons = append(buttons, prevButton)
	}
	if end < len(tasks) {
		nextButton := tview.NewButton("Next").SetSelectedFunc(func() {
			if end+5 < len(tasks) {
				uim.renderStatsPage(end, end+statisticsPageSize, tasks)
			} else {
				uim.renderStatsPage(end, len(tasks), tasks)
			}
			uim.pages.SwitchToPage(statsPage)
		})

		buttons = append(buttons, nextButton)
	}

	return buttons
}

func (uim *UIManager) newStatsTable(start, end int, tasks []*Task) *tview.Table {
	table := tview.NewTable().SetBorders(true)
	headers := []string{"Date", "Time", "Seconds", "Action"}

	for col, header := range headers {
		table.SetCell(0, col, tview.NewTableCell(header).SetAlign(tview.AlignCenter))
	}

	for i, task := range tasks[start:end] {
		row := i + 1

		dateStr := task.StartAt.Format("02-Jan-2006")
		const timeFormat = "15:04:05"
		timeStr := fmt.Sprintf("%s-%s", task.StartAt.Format(timeFormat), task.FinishAt.Format(timeFormat))

		table.SetCell(row, 0, tview.NewTableCell(dateStr).SetAlign(tview.AlignCenter))
		table.SetCell(row, 1, tview.NewTableCell(timeStr).SetAlign(tview.AlignCenter))
		table.SetCell(row, 2, tview.NewTableCell(strconv.Itoa(task.Duration)).SetAlign(tview.AlignCenter))
		table.SetCell(row, 3, tview.NewTableCell("[red] Delete [-]").SetAlign(tview.AlignCenter).SetSelectable(true))
	}

	table.SetInputCapture(uim.captureTableInput(table, tasks))
	return table
}

func (uim *UIManager) captureTableInput(table *tview.Table, tasks []*Task) func(*tcell.EventKey) *tcell.EventKey {
	handleEnterKey := func(table *tview.Table, tasks []*Task, col int) {
		if len(tasks) > 0 && col != 3 {
			table.Select(1, 3).SetSelectable(true, true)
		} else if col == 3 {
			table.Select(0, 0).SetSelectable(false, false)
		}
	}

	return func(event *tcell.EventKey) *tcell.EventKey {
		row, col := table.GetSelection()
		switch event.Key() {
		case tcell.KeyEnter:
			handleEnterKey(table, tasks, col)
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

func (uim *UIManager) removeTask(tasks []*Task, taskIndex int) {
	err := uim.taskManager.RemoveTask(tasks[taskIndex].ID)
	if err != nil {
		uim.logger.Error("Can't delete task", slog.Any("id", tasks[taskIndex].ID))
	}

	tasks = append(tasks[:taskIndex], tasks[taskIndex+1:]...)
	if len(tasks) > 5 {
		uim.renderStatsPage(0, statisticsPageSize, tasks)
	} else {
		uim.renderStatsPage(0, len(tasks), tasks)
	}
	uim.pages.SwitchToPage(statsPage)
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
	table := uim.newStatsTable(start, end, tasks)

	buttons := uim.newStatsButtons(start, end, tasks)

	text := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("Total: %s", uim.totalDuration(tasks))).
		SetDynamicColors(true)

	form := tview.NewForm().
		SetHorizontal(true).
		AddInputField("Time start", time.Now().Format("15:04"), 7, checkTimeInInput(), nil).
		AddInputField("Seconds", "0", 5, tview.InputFieldInteger, nil)

	form.AddButton("Save", uim.saveTask(form))

	grid := tview.NewGrid().
		SetRows(1, 3, 0, 1, 1).
		SetColumns(0, 25, 25, 0).
		SetBorders(true)

	grid.AddItem(text, 0, 1, 1, 2, 0, 0, false)

	grid.AddItem(form, 1, 1, 1, 2, 0, 0, true)

	grid.AddItem(table, 2, 1, 1, 2, 0, 0, false)

	for i, button := range buttons {
		grid.AddItem(button, 3, i+1, 1, 1, 0, 0, false)
	}

	uim.pages.AddPage(insertStatsPage, grid, true, false)
}

func checkTimeInInput() func(textToCheck string, lastChar rune) bool {
	return func(textToCheck string, lastChar rune) bool {
		switch {
		case len(textToCheck) == 5:
			_, err := time.Parse("15:04", textToCheck)
			return err == nil
		case len(textToCheck) == 4:
			return unicode.IsDigit(lastChar) && lastChar <= '5'
		case len(textToCheck) == 3:
			return lastChar == ':'
		case len(textToCheck) == 2:
			return unicode.IsDigit(lastChar)
		case len(textToCheck) == 1:
			return unicode.IsDigit(lastChar) && lastChar <= '2'
		}

		return false
	}
}

func (uim *UIManager) saveTask(form *tview.Form) func() {
	return func() {
		timeStartStr := form.GetFormItem(0).(*tview.InputField).GetText()
		secondsStr := form.GetFormItem(1).(*tview.InputField).GetText()

		timeStart, err := time.Parse("15:04", timeStartStr)
		if err != nil {
			uim.logger.Error("can't convert str to time", slog.String("input time", timeStartStr))
			return
		}
		seconds, err := strconv.Atoi(secondsStr)
		if err != nil {
			uim.logger.Error("can't convert seconds string to int", slog.String("input seconds", secondsStr))
			return
		}
		dateStart := timeStart.AddDate(time.Now().Year(), int(time.Now().Month())-1, time.Now().Day()-1)

		err = uim.saveTaskFromForm(dateStart, seconds)
		if err != nil {
			uim.logger.Error("can't create task", slog.Any("error", err))
			uim.switchToStatisticsPage(insertStatsPage)
			return
		}
		uim.switchToStatisticsPage(statsPage)
	}
}

func (uim *UIManager) saveTaskFromForm(timeStart time.Time, duration int) error {
	finishTime := timeStart.Add(time.Duration(duration) * time.Second)
	_, err := uim.taskManager.CreateNewTask(timeStart, finishTime, duration)
	return err
}

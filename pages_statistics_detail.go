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

const statisticsPageSize = 7

func (m *UIManager) renderDetailStatsPage(args ...any) func() tview.Primitive {
	start := args[0].(int)
	end := args[1].(int)
	pomodoros := args[2].([]*Pomodoro)

	return func() tview.Primitive {
		table := m.newStatsTable(start, end, pomodoros)

		buttons := m.newStatsButtons(start, end, pomodoros)

		text := tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(fmt.Sprintf("Total: %s", m.totalDuration(pomodoros))).
			SetDynamicColors(true)

		grid := tview.NewGrid().
			SetRows(1, 0, 1).
			SetColumns(0, 23, 23, 0)

		grid.AddItem(text, 0, 1, 1, 2, 0, 0, false)
		grid.AddItem(table, 1, 1, 1, 2, 0, 0, true)

		for i, button := range buttons {
			grid.AddItem(button, 2, i+1, 1, 1, 0, 0, true)
		}

		grid.SetInputCapture(m.captureStatsInput(table, buttons))
		return grid
	}
}

func (m *UIManager) captureStatsInput(
	table *tview.Table,
	buttons []*tview.Button) func(*tcell.EventKey) *tcell.EventKey {

	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:
			if m.isButtonFocused(buttons) {
				m.ui.SetFocus(table)
			} else {
				m.ui.SetFocus(buttons[0])
			}
		case tcell.KeyLeft, tcell.KeyRight:
			if len(buttons) > 0 {
				prevIndx := buttonIndex(m.ui.GetFocus(), buttons)
				newIndx := (prevIndx + 1) % len(buttons)
				m.ui.SetFocus(buttons[newIndx])
			}
		default:
		}
		return event
	}
}

func (m *UIManager) isButtonFocused(buttons []*tview.Button) bool {
	focusedPrimitive := m.ui.GetFocus()
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

func (m *UIManager) newStatsButtons(start, end int, pomodoros []*Pomodoro) []*tview.Button {
	buttons := make([]*tview.Button, 0)

	if start >= statisticsPageSize {
		prevButton := tview.NewButton("Prev").SetSelectedFunc(func() {
			m.AddPageAndSwitch(m.NewDetailStats(start-statisticsPageSize, start))
		})

		buttons = append(buttons, prevButton)
	}
	if end < len(pomodoros) {
		nextButton := tview.NewButton("Next").SetSelectedFunc(func() {
			if end+5 < len(pomodoros) {
				m.AddPageAndSwitch(m.NewDetailStats(end, end+statisticsPageSize))
			} else {
				m.AddPageAndSwitch(m.NewDetailStats(end, len(pomodoros)))
			}
		})

		buttons = append(buttons, nextButton)
	}

	return buttons
}

func (m *UIManager) newStatsTable(start, end int, pomodoros []*Pomodoro) *tview.Table {
	table := tview.NewTable().SetBorders(true)
	headers := []string{"Date", "Time", "Minutes", "Action"}

	for col, header := range headers {
		table.SetCell(0, col, tview.NewTableCell(header).SetAlign(tview.AlignCenter))
	}

	for i, pmdr := range pomodoros[start:end] {
		row := i + 1

		dateStr := pmdr.StartAt.Format("02-Jan-2006")
		const timeFormat = "15:04"
		timeStr := fmt.Sprintf("%s-%s", pmdr.StartAt.Format(timeFormat), pmdr.FinishAt.Format(timeFormat))

		table.SetCell(row, 0, tview.NewTableCell(dateStr).SetAlign(tview.AlignCenter))
		table.SetCell(row, 1, tview.NewTableCell(timeStr).SetAlign(tview.AlignCenter))
		table.SetCell(row, 2, tview.NewTableCell(strconv.Itoa(pmdr.SecondsDuration/60)).SetAlign(tview.AlignCenter))
		table.SetCell(row, 3, tview.NewTableCell("[red] Delete [-]").SetAlign(tview.AlignCenter).SetSelectable(true))
	}

	table.SetInputCapture(m.captureTableInput(table, pomodoros))
	return table
}

func (m *UIManager) captureTableInput(table *tview.Table, pomdoro []*Pomodoro) func(*tcell.EventKey) *tcell.EventKey {
	handleEnterKey := func(table *tview.Table, pomodoro []*Pomodoro, col int) {
		if len(pomodoro) > 0 && col != 3 {
			table.Select(1, 3).SetSelectable(true, true)
		} else if col == 3 {
			table.Select(0, 0).SetSelectable(false, false)
		}
	}

	return func(event *tcell.EventKey) *tcell.EventKey {
		row, col := table.GetSelection()
		switch event.Key() {
		case tcell.KeyEnter:
			handleEnterKey(table, pomdoro, col)
		case tcell.KeyDown, tcell.KeyUp:
			m.handleVerticalNavigation(table, row, col, event.Key())
		case tcell.KeyLeft, tcell.KeyRight:
			if col != 3 {
				table.Select(row, 3)
			}
		case tcell.KeyCtrlY:
			if col == 3 && row > 0 {
				m.removePomodoro(pomdoro, row-1)
			}
		case tcell.KeyEscape:
			table.Select(0, 0).SetSelectable(false, false)
		default:
			return event
		}

		return nil
	}
}

func (m *UIManager) handleVerticalNavigation(table *tview.Table, row, col int, key tcell.Key) {
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

func (m *UIManager) removePomodoro(pomodoros []*Pomodoro, pomodoroIndx int) {
	err := m.pomodoroTracker.RemovePomodoro(pomodoros[pomodoroIndx].ID)
	if err != nil {
		m.logger.Error("Can't delete pomodoro", slog.Any("id", pomodoros[pomodoroIndx].ID))
	}
	m.AddPageAndSwitch(m.NewDetailStats(-1, -1))
}

func (m *UIManager) totalDuration(pomodoros []*Pomodoro) string {
	var total int
	for _, t := range pomodoros {
		total += t.SecondsDuration
	}
	res := time.Duration(total) * time.Second
	return res.String()
}

func (m *UIManager) renderInsertStatsPage(args ...any) func() tview.Primitive {
	start := args[0].(int)
	end := args[1].(int)
	pomodoros := args[2].([]*Pomodoro)

	return func() tview.Primitive {
		table := m.newStatsTable(start, end, pomodoros)

		buttons := m.newStatsButtons(start, end, pomodoros)

		text := tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(fmt.Sprintf("Total: %s", m.totalDuration(pomodoros))).
			SetDynamicColors(true)

		form := tview.NewForm().
			SetHorizontal(true).
			AddInputField("Time start", time.Now().Format("15:04"), 7, checkTimeInInput(), nil).
			AddInputField("Minutes", "0", 5, tview.InputFieldInteger, nil)

		form.AddButton("Save", m.savePomodoro(form))

		grid := tview.NewGrid().
			SetRows(1, 3, 0, 1, 1).
			SetColumns(0, 23, 23, 0).
			SetBorders(true)

		grid.AddItem(text, 0, 1, 1, 2, 0, 0, false)

		grid.AddItem(form, 1, 1, 1, 2, 0, 0, true)

		grid.AddItem(table, 2, 1, 1, 2, 0, 0, false)

		for i, button := range buttons {
			grid.AddItem(button, 3, i+1, 1, 1, 0, 0, false)
		}
		return grid
	}
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

func (m *UIManager) savePomodoro(form *tview.Form) func() {
	return func() {
		timeStartStr := form.GetFormItem(0).(*tview.InputField).GetText() //nolint:errcheck // simple parsing
		minutesStr := form.GetFormItem(1).(*tview.InputField).GetText()   //nolint:errcheck // simple parsing

		timeStart, err := time.Parse("15:04", timeStartStr)
		if err != nil {
			m.logger.Error("can't convert str to time", slog.String("input time", timeStartStr))
			return
		}
		minutes, err := strconv.Atoi(minutesStr)
		if err != nil {
			minutes = 0
			m.logger.Error("can't convert seconds string to int", slog.String("input seconds", minutesStr))
		}
		dateStart := timeStart.AddDate(time.Now().Year(), int(time.Now().Month())-1, time.Now().Day()-1)

		err = m.savePomodoroFromForm(dateStart, minutes)
		if err != nil {
			m.logger.Error("can't create pomodoro", slog.Any("error", err))
			m.AddPageAndSwitch(m.NewInsertDetailPage(-1, -1))
			return
		}
		m.AddPageAndSwitch(m.NewDetailStats(-1, -1))
	}
}

func (m *UIManager) savePomodoroFromForm(timeStart time.Time, minutes int) error {
	finishTime := timeStart.Add(time.Duration(minutes) * time.Minute)
	_, err := m.pomodoroTracker.CreateNewPomodoro(timeStart, finishTime, minutes*60)
	return err
}

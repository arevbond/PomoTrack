package main

import (
	"fmt"
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
	var nextButton *tview.Button
	var prevButton *tview.Button
	table := uim.statisticsTable(start, end, tasks)

	const pageSize = 5
	if start >= pageSize {
		prevButton = tview.NewButton("Prev").SetSelectedFunc(func() {
			uim.pageStatistics(start-pageSize, start, tasks)
			uim.pages.SwitchToPage(pageNameStatistics)
		})
	}
	if end < len(tasks) {
		nextButton = tview.NewButton("Next").SetSelectedFunc(func() {
			if end+5 < len(tasks) {
				uim.pageStatistics(end, end+pageSize, tasks)
			} else {
				uim.pageStatistics(end, len(tasks), tasks)
			}
			uim.pages.SwitchToPage(pageNameStatistics)
		})
	}

	text := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("Total: %s", uim.totalDuration(tasks))).
		SetDynamicColors(true)

	grid := tview.NewGrid().
		SetRows(3, 0, 1, 1).      // Добавляем строки для текста, таблицы и кнопок
		SetColumns(0, 25, 25, 0). // Определяем 4 столбца: 0 — для отступов, 25 — для кнопок
		SetBorders(true)

	grid.AddItem(text, 0, 1, 1, 2, 0, 0, false)

	grid.AddItem(table, 1, 1, 1, 2, 0, 0, false)

	if prevButton != nil {
		grid.AddItem(prevButton, 2, 1, 1, 1, 0, 0, true)
	}
	if nextButton != nil {
		grid.AddItem(nextButton, 2, 2, 1, 1, 0, 0, true)
	}

	if nextButton != nil || prevButton != nil {
		grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyTAB, tcell.KeyLeft, tcell.KeyRight:
				if uim.ui.GetFocus() == prevButton && nextButton != nil {
					uim.ui.SetFocus(nextButton)
				} else if uim.ui.GetFocus() == nextButton && prevButton != nil {
					uim.ui.SetFocus(prevButton)
				}
			default:
			}
			return event
		})
	}

	uim.pages.AddPage(pageNameStatistics, grid, true, false)
}

func (uim *UIManager) statisticsTable(start, end int, tasks []*Task) *tview.Table {
	table := tview.NewTable().
		SetBorders(true)

	table.SetCell(0, 0, tview.NewTableCell("Date").SetAlign(tview.AlignCenter))
	table.SetCell(0, 1, tview.NewTableCell("Time").SetAlign(tview.AlignCenter))
	table.SetCell(0, 2, tview.NewTableCell("Seconds").SetAlign(tview.AlignCenter))

	for i, t := range tasks[start:end] {
		dateStr := t.StartAt.Format("02-Jan-2006")
		table.SetCell(i+1, 0, tview.NewTableCell(dateStr).SetAlign(tview.AlignCenter))

		const timeFormat = "15:04:05"
		timeStr := fmt.Sprintf("%s-%s", t.StartAt.Format(timeFormat), t.FinishAt.Format(timeFormat))
		table.SetCell(i+1, 1, tview.NewTableCell(timeStr).SetAlign(tview.AlignCenter))

		table.SetCell(i+1, 2, tview.NewTableCell(strconv.Itoa(t.Duration)).SetAlign(tview.AlignCenter))
	}

	return table
}

func (uim *UIManager) totalDuration(tasks []*Task) string {
	var total int
	for _, t := range tasks {
		total += t.Duration
	}
	res := time.Duration(total) * time.Second
	return res.String()
}

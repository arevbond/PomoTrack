package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	activeFocusPage = "Active-Focus"
	pauseFocusPage  = "Stop-Focus"

	activeBreakPage = "Active-Break"
	pauseBreakPage  = "Stop-Break"
)

func formatDuration(d time.Duration) string {
	totalSeconds := int(d.Seconds())

	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func (m *UIManager) renderPausePage(pageName string, title string, timerType TimerType) {
	pauseText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(title)

	durationText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(formatDuration(m.stateManager.timeToFinish(timerType)))

	startButton := tview.NewButton("▶ Start").SetSelectedFunc(func() {
		m.stateManager.SetState(StateActive, timerType)
	})

	hotKeysPanel := constructBottomPanel(pageName)

	grid := tview.NewGrid().
		SetRows(0, 1, 3, 1, 0, 1).
		SetColumns(0, 0, 15, 0, 0).
		SetBorders(true)

	grid.AddItem(pauseText, 1, 2, 1, 1, 0, 0, false)
	grid.AddItem(durationText, 2, 2, 1, 1, 0, 0, false)
	grid.AddItem(startButton, 3, 2, 1, 1, 0, 0, true)
	grid.AddItem(hotKeysPanel, 5, 1, 1, 3, 0, 0, false)

	m.pages.AddPage(pageName, grid, true, true)
}

func (m *UIManager) renderActivePage(pageName string, color, title string, timerType TimerType) {
	breakText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("[%s]%s", color, title))

	timerText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
			tview.Print(screen, fmt.Sprintf("[%s]%s[-]", color,
				formatDuration(m.stateManager.timeToFinish(timerType))),
				x, y+height/4, width, tview.AlignCenter, tcell.ColorLime)
			return 0, 0, 0, 0
		})

	pauseButton := tview.NewButton("Pause").SetSelectedFunc(func() {
		m.stateManager.SetState(StatePaused, timerType)
	})

	toggleButton := tview.NewButton("→").SetSelectedFunc(func() {
		m.stateManager.SetState(StateFinished, timerType)
	})

	grid := tview.NewGrid().
		SetRows(0, 1, 3, 1, 0, 1).
		SetColumns(0, 15, 5, 0).
		SetBorders(true)

	grid.AddItem(breakText, 1, 1, 1, 2, 0, 0, false)
	grid.AddItem(timerText, 2, 1, 1, 2, 0, 0, false)
	grid.AddItem(pauseButton, 3, 1, 1, 1, 0, 0, true)
	grid.AddItem(toggleButton, 3, 2, 1, 1, 0, 0, false)

	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB, tcell.KeyLeft, tcell.KeyRight:
			if m.ui.GetFocus() == pauseButton {
				m.ui.SetFocus(toggleButton)
			} else {
				m.ui.SetFocus(pauseButton)
			}
		default:
			return event
		}
		return nil
	})

	m.pages.AddPage(pageName, grid, true, false)
}

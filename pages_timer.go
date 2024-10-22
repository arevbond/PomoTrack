package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	activeFocusPage = "Active-Focus"
	pauseFocusPage  = "Stop-Focus"

	activeBreakPage = "Active-Break"
	pauseBreakPage  = "Stop-Break"
)

func (uim *UIManager) renderPausePage(pageName string, title string, timerType TimerType) {
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

func (uim *UIManager) renderActivePage(pageName string, color, title string, timerType TimerType) {
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

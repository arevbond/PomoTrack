package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func formatDuration(d time.Duration) string {
	totalSeconds := int(d.Seconds())

	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func (m *UIManager) NewActivePage(timerType TimerType, stopSignal chan struct{}) *Page {
	go playClickSound()

	var render func() tview.Primitive
	var pageName PageName
	switch timerType {
	case FocusTimer:
		pageName = activeFocusPage
		render = m.renderActivePage("red", "Pomodoro", FocusTimer)
	case BreakTimer:
		pageName = activeBreakPage
		render = m.renderActivePage("green", "Break", BreakTimer)
	}

	go m.updateUIWithTicker(stopSignal)
	return NewPageComponent(pageName, true, false, render)
}

func (m *UIManager) updateUIWithTicker(quit chan struct{}) {
	tick := time.NewTicker(screenRefreshInterval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			m.ui.Draw()
		case <-quit:
			return
		}
	}
}

func (m *UIManager) NewPausePage(timerType TimerType) *Page {
	var render func() tview.Primitive
	var pageName PageName
	switch timerType {
	case BreakTimer:
		pageName = pauseBreakPage
		render = m.renderPausePage("Break", BreakTimer)
	case FocusTimer:
		pageName = pauseFocusPage
		render = m.renderPausePage("Pomodoro", FocusTimer)
	}
	return NewPageComponent(pageName, true, false, render)
}

func (m *UIManager) handleStatePaused(timerType TimerType) {
	go playClickSound()

	switch timerType {
	case FocusTimer:
		m.ui.QueueUpdateDraw(func() {
			m.AddPageAndSwitch(m.NewPausePage(FocusTimer))
		})
	case BreakTimer:
		m.ui.QueueUpdateDraw(func() {
			m.AddPageAndSwitch(m.NewPausePage(BreakTimer))
		})
	}
}

func (m *UIManager) handleStateFinished(timerType TimerType) {
	go playEndSound()

	switch timerType {
	case FocusTimer:
		m.ui.QueueUpdateDraw(func() {
			m.AddPageAndSwitch(m.NewPausePage(BreakTimer))
		})
	case BreakTimer:
		m.ui.QueueUpdateDraw(func() {
			m.AddPageAndSwitch(m.NewPausePage(FocusTimer))
		})
	}
}

func (m *UIManager) renderPausePage(args ...any) func() tview.Primitive {
	title, ok := args[0].(string)
	if !ok {
		m.logger.Error("can't extract arg", slog.String("func", "renderPausePage"),
			slog.String("expected", "string (title)"))
		return nil
	}
	timerType, ok := args[1].(TimerType)
	if !ok {
		m.logger.Error("can't extract arg", slog.String("func", "renderPausePage"),
			slog.String("expected", "timerType (TimerType)"))
		return nil
	}

	return func() tview.Primitive {
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

		grid := tview.NewGrid().
			SetRows(0, 1, 3, 1, 0, 1).
			SetColumns(0, 0, 15, 0, 0).
			SetBorders(true)

		grid.AddItem(pauseText, 1, 2, 1, 1, 0, 0, false)
		grid.AddItem(durationText, 2, 2, 1, 1, 0, 0, false)
		grid.AddItem(startButton, 3, 2, 1, 1, 0, 0, true)

		return grid
	}
}

func (m *UIManager) renderActivePage(args ...any) func() tview.Primitive {
	color, ok := args[0].(string)
	if !ok {
		m.logger.Error("can't extract arg", slog.String("func", "renderActivePage"),
			slog.String("expected", "color (string)"))
		return nil
	}
	title, ok := args[1].(string)
	if !ok {
		m.logger.Error("can't extract arg", slog.String("func", "renderActivePage"),
			slog.String("expected", "title (string)"))
		return nil
	}
	timerType, ok := args[2].(TimerType)
	if !ok {
		m.logger.Error("can't extract arg", slog.String("func", "renderActivePage"),
			slog.String("expected", "timerType (timerType)"))
		return nil
	}

	return func() tview.Primitive {
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

		hiddenTimerText := tview.NewTextView().
			SetDynamicColors(true).
			SetTextAlign(tview.AlignCenter).
			SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
				tview.Print(screen, fmt.Sprintf("[%s]%s[-]", color,
					"focusing"),
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
		if m.stateManager.IsFocusTimeHidden() && timerType == FocusTimer {
			grid.AddItem(hiddenTimerText, 2, 1, 1, 2, 0, 0, false)
		} else {
			grid.AddItem(timerText, 2, 1, 1, 2, 0, 0, false)
		}
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

		return grid
	}
}

package main

import (
	"log/slog"
	"time"

	"github.com/arevbond/PomoTrack/config"

	"github.com/gdamore/tcell/v2"

	"github.com/rivo/tview"
)

const (
	pageNameActiveFocus = "Active-Focus"
	pageNamePauseFocus  = "Stop-Focus"

	pageNameActiveBreak = "Active-Break"
	pageNamePauseBreak  = "Stop-Break"

	pageNameStatistics = "Statistics"
)

const screenRefreshInterval = 1 * time.Second

type taskMngr interface {
	HandleTaskStateChanges()
	TodayTasks() ([]*Task, error)
	Tasks(limit int) ([]*Task, error)
}

type UIManager struct {
	ui              *tview.Application
	pages           *tview.Pages
	logger          *slog.Logger
	stateManager    *StateManager
	stateChangeChan chan StateChangeEvent

	taskManager   taskMngr
	stateTaskChan chan StateChangeEvent
}

type StateChangeEvent struct {
	TimerType TimerType
	NewState  TimerState
}

func NewUIManager(logger *slog.Logger, cfg *config.Config, events chan StateChangeEvent, tm taskMngr) *UIManager {
	stateChangeChan := make(chan StateChangeEvent)
	focusTimer := NewFocusTimer(cfg.Timer.FocusDuration)
	breakTimer := NewBreakTimer(cfg.Timer.BreakDuration)

	return &UIManager{
		ui:              tview.NewApplication(),
		pages:           tview.NewPages(),
		logger:          logger,
		stateManager:    NewStateManager(logger, focusTimer, breakTimer, stateChangeChan, cfg.Timer),
		stateChangeChan: stateChangeChan,
		taskManager:     tm,
		stateTaskChan:   events,
	}
}

func (uim *UIManager) DefaultTimerPages() {
	uim.pageActiveTimer(pageNameActiveFocus, "purple", "Pomodoro", FocusTimer)
	uim.pageActiveTimer(pageNameActiveBreak, "green", "Break", BreakTimer)
	uim.pagePause(pageNamePauseBreak, "Break", BreakTimer)
	uim.pagePause(pageNamePauseFocus, "Pomodoro", FocusTimer)
}

func (uim *UIManager) setKeyboardEvents() {
	uim.ui.SetInputCapture(uim.keyboardEvents)
}

func (uim *UIManager) keyboardEvents(event *tcell.EventKey) *tcell.EventKey {
	name, _ := uim.pages.GetFrontPage()

	switch event.Key() {
	case tcell.KeyF1:
		if name != pageNameStatistics && name != pageNamePauseBreak {
			return event
		}
		uim.pages.SwitchToPage(pageNamePauseFocus)
	case tcell.KeyF2:
		if name != pageNameStatistics && name != pageNamePauseFocus {
			return event
		}
		uim.pages.SwitchToPage(pageNamePauseBreak)
	case tcell.KeyF3:
		if name == pageNamePauseFocus || name == pageNamePauseBreak {
			tasks, err := uim.taskManager.TodayTasks()
			if err != nil {
				uim.logger.Error("can't get today tasks", slog.Any("error", err))
				return event
			}
			if len(tasks) > 5 {
				uim.pageStatistics(0, 5, tasks)
			} else {
				uim.pageStatistics(0, len(tasks), tasks)
			}
			uim.pages.SwitchToPage(pageNameStatistics)
		}
	default:
		return event
	}
	return nil
}

func (uim *UIManager) HandleStatesAndKeyboard() {
	stopRefreshing := make(chan struct{})
	uim.setKeyboardEvents()
	go uim.handleTimerStateChanges(stopRefreshing)
}

func (uim *UIManager) handleTimerStateChanges(stopRefreshing chan struct{}) {
	for event := range uim.stateChangeChan {
		uim.stateTaskChan <- event

		switch event.NewState {
		case StateActive:
			uim.handleStateActive(event.TimerType, stopRefreshing)

		case StatePaused, StateFinished:
			stopRefreshing <- struct{}{}

			if event.NewState == StatePaused {
				uim.handleStatePaused(event.TimerType)
			} else if event.NewState == StateFinished {
				uim.handleStateFinished(event.TimerType)
			}
		}
	}
}

func (uim *UIManager) handleStateActive(timerType TimerType, quit chan struct{}) {
	go playClickSound()

	switch timerType {
	case FocusTimer:
		uim.showPage(pageNameActiveFocus)
	case BreakTimer:
		uim.showPage(pageNameActiveBreak)
	}

	go uim.updateUIWithTicker(quit)
}

func (uim *UIManager) updateUIWithTicker(quit chan struct{}) {
	tick := time.NewTicker(screenRefreshInterval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			uim.ui.Draw()
		case <-quit:
			return
		}
	}
}

func (uim *UIManager) handleStatePaused(timerType TimerType) {
	go playClickSound()

	switch timerType {
	case FocusTimer:
		uim.updateAndShowPausePage(FocusTimer)
	case BreakTimer:
		uim.updateAndShowPausePage(BreakTimer)
	}
}

func (uim *UIManager) handleStateFinished(timerType TimerType) {
	go playEndSound()

	switch timerType {
	case FocusTimer:
		uim.updateAndShowPausePage(BreakTimer)
	case BreakTimer:
		uim.updateAndShowPausePage(FocusTimer)
	}
}

func (uim *UIManager) updateAndShowPausePage(timerType TimerType) {
	switch timerType {
	case BreakTimer:
		uim.pagePause(pageNamePauseBreak, "Break", BreakTimer)
		uim.showPage(pageNamePauseBreak)
	case FocusTimer:
		uim.pagePause(pageNamePauseFocus, "Pomodoro", FocusTimer)
		uim.showPage(pageNamePauseFocus)
	}
}

func (uim *UIManager) showPage(pageName string) {
	uim.ui.QueueUpdateDraw(func() {
		uim.pages.SwitchToPage(pageName)
	})
}

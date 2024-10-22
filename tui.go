package main

import (
	"log/slog"
	"time"

	"github.com/arevbond/PomoTrack/config"

	"github.com/gdamore/tcell/v2"

	"github.com/rivo/tview"
)

const screenRefreshInterval = 1 * time.Second

type taskTracker interface {
	HandleTaskStateChanges()
	TodayTasks() ([]*Task, error)
	Tasks(limit int) ([]*Task, error)
	RemoveTask(id int) error
	CreateNewTask(startAt time.Time, finishAt time.Time, duration int) (*Task, error)
}

type UIManager struct {
	ui           *tview.Application
	pages        *tview.Pages
	logger       *slog.Logger
	stateManager *StateManager
	stateUpdates chan StateChangeEvent

	taskManager      taskTracker
	stateTaskUpdates chan StateChangeEvent
}

type StateChangeEvent struct {
	TimerType TimerType
	NewState  TimerState
}

func NewUIManager(logger *slog.Logger, cfg *config.Config, events chan StateChangeEvent, tm taskTracker) *UIManager {
	stateChangeChan := make(chan StateChangeEvent)
	focusTimer := NewFocusTimer(cfg.Timer.FocusDuration)
	breakTimer := NewBreakTimer(cfg.Timer.BreakDuration)

	return &UIManager{
		ui:               tview.NewApplication(),
		pages:            tview.NewPages(),
		logger:           logger,
		stateManager:     NewStateManager(logger, focusTimer, breakTimer, stateChangeChan, cfg.Timer),
		stateUpdates:     stateChangeChan,
		taskManager:      tm,
		stateTaskUpdates: events,
	}
}

func (uim *UIManager) DefaultTimerPages() {
	uim.renderActivePage(activeFocusPage, "purple", "Pomodoro", FocusTimer)
	uim.renderActivePage(activeBreakPage, "green", "Break", BreakTimer)
	uim.renderPausePage(pauseBreakPage, "Break", BreakTimer)
	uim.renderPausePage(pauseFocusPage, "Pomodoro", FocusTimer)
}

func (uim *UIManager) setKeyboardEvents() {
	uim.ui.SetInputCapture(uim.keyboardEvents)
}

func (uim *UIManager) keyboardEvents(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyF1:
		availablePages := []string{statsPage, pauseBreakPage}
		if !uim.currentPageIs(availablePages) {
			return event
		}
		uim.pages.SwitchToPage(pauseFocusPage)
	case tcell.KeyF2:
		availablePages := []string{statsPage, pauseFocusPage}
		if !uim.currentPageIs(availablePages) {
			return event
		}
		uim.pages.SwitchToPage(pauseBreakPage)
	case tcell.KeyF3:
		availablePages := []string{pauseFocusPage, pauseBreakPage, insertStatsPage}
		if uim.currentPageIs(availablePages) {
			uim.switchToStatisticsPage(statsPage)
		}
	case tcell.KeyCtrlI:
		availablePages := []string{statsPage}
		if uim.currentPageIs(availablePages) {
			uim.switchToStatisticsPage(insertStatsPage)
		}
	default:
		return event
	}
	return nil
}

func (uim *UIManager) switchToStatisticsPage(pageName string) {
	tasks, err := uim.taskManager.TodayTasks()
	if err != nil {
		uim.logger.Error("can't get today tasks", slog.Any("error", err))
		return
	}
	if len(tasks) > statisticsPageSize {
		uim.pageInsertStatistics(0, statisticsPageSize, tasks)
		uim.renderStatsPage(0, statisticsPageSize, tasks)
	} else {
		uim.pageInsertStatistics(0, len(tasks), tasks)
		uim.renderStatsPage(0, len(tasks), tasks)
	}
	uim.pages.SwitchToPage(pageName)
}

func (uim *UIManager) currentPageIs(pages []string) bool {
	name, _ := uim.pages.GetFrontPage()
	for _, pageName := range pages {
		if name == pageName {
			return true
		}
	}
	return false
}

func (uim *UIManager) InitStateAndKeyboardHandling() {
	stopRefreshing := make(chan struct{})
	uim.setKeyboardEvents()
	go uim.listenToStateChanges(stopRefreshing)
}

func (uim *UIManager) listenToStateChanges(stopRefreshing chan struct{}) {
	for event := range uim.stateUpdates {
		uim.stateTaskUpdates <- event

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

func (uim *UIManager) handleStateActive(timerType TimerType, stopSignal chan struct{}) {
	go playClickSound()

	switch timerType {
	case FocusTimer:
		uim.displayPage(activeFocusPage)
	case BreakTimer:
		uim.displayPage(activeBreakPage)
	}

	go uim.updateUIWithTicker(stopSignal)
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
		uim.renderPausePageForTimer(FocusTimer)
	case BreakTimer:
		uim.renderPausePageForTimer(BreakTimer)
	}
}

func (uim *UIManager) handleStateFinished(timerType TimerType) {
	go playEndSound()

	switch timerType {
	case FocusTimer:
		uim.renderPausePageForTimer(BreakTimer)
	case BreakTimer:
		uim.renderPausePageForTimer(FocusTimer)
	}
}

func (uim *UIManager) renderPausePageForTimer(timerType TimerType) {
	switch timerType {
	case BreakTimer:
		uim.renderPausePage(pauseBreakPage, "Break", BreakTimer)
		uim.displayPage(pauseBreakPage)
	case FocusTimer:
		uim.renderPausePage(pauseFocusPage, "Pomodoro", FocusTimer)
		uim.displayPage(pauseFocusPage)
	}
}

func (uim *UIManager) displayPage(pageName string) {
	uim.ui.QueueUpdateDraw(func() {
		uim.pages.SwitchToPage(pageName)
	})
}

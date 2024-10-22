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

func (m *UIManager) DefaultTimerPages() {
	m.renderActivePage(activeFocusPage, "purple", "Pomodoro", FocusTimer)
	m.renderActivePage(activeBreakPage, "green", "Break", BreakTimer)
	m.renderPausePage(pauseBreakPage, "Break", BreakTimer)
	m.renderPausePage(pauseFocusPage, "Pomodoro", FocusTimer)
}

func (m *UIManager) setKeyboardEvents() {
	m.ui.SetInputCapture(m.keyboardEvents)
}

func (m *UIManager) keyboardEvents(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyF1:
		availablePages := []string{statsPage, pauseBreakPage}
		if !m.currentPageIs(availablePages) {
			return event
		}
		m.pages.SwitchToPage(pauseFocusPage)
	case tcell.KeyF2:
		availablePages := []string{statsPage, pauseFocusPage}
		if !m.currentPageIs(availablePages) {
			return event
		}
		m.pages.SwitchToPage(pauseBreakPage)
	case tcell.KeyF3:
		availablePages := []string{pauseFocusPage, pauseBreakPage, insertStatsPage}
		if m.currentPageIs(availablePages) {
			m.switchToStatisticsPage(statsPage)
		}
	case tcell.KeyCtrlI:
		availablePages := []string{statsPage}
		if m.currentPageIs(availablePages) {
			m.switchToStatisticsPage(insertStatsPage)
		}
	default:
		return event
	}
	return nil
}

func (m *UIManager) switchToStatisticsPage(pageName string) {
	tasks, err := m.taskManager.TodayTasks()
	if err != nil {
		m.logger.Error("can't get today tasks", slog.Any("error", err))
		return
	}
	if len(tasks) > statisticsPageSize {
		m.pageInsertStatistics(0, statisticsPageSize, tasks)
		m.renderStatsPage(0, statisticsPageSize, tasks)
	} else {
		m.pageInsertStatistics(0, len(tasks), tasks)
		m.renderStatsPage(0, len(tasks), tasks)
	}
	m.pages.SwitchToPage(pageName)
}

func (m *UIManager) currentPageIs(pages []string) bool {
	name, _ := m.pages.GetFrontPage()
	for _, pageName := range pages {
		if name == pageName {
			return true
		}
	}
	return false
}

func (m *UIManager) InitStateAndKeyboardHandling() {
	stopRefreshing := make(chan struct{})
	m.setKeyboardEvents()
	go m.listenToStateChanges(stopRefreshing)
}

func (m *UIManager) listenToStateChanges(stopRefreshing chan struct{}) {
	for event := range m.stateUpdates {
		m.stateTaskUpdates <- event

		switch event.NewState {
		case StateActive:
			m.handleStateActive(event.TimerType, stopRefreshing)

		case StatePaused, StateFinished:
			stopRefreshing <- struct{}{}

			if event.NewState == StatePaused {
				m.handleStatePaused(event.TimerType)
			} else if event.NewState == StateFinished {
				m.handleStateFinished(event.TimerType)
			}
		}
	}
}

func (m *UIManager) handleStateActive(timerType TimerType, stopSignal chan struct{}) {
	go playClickSound()

	switch timerType {
	case FocusTimer:
		m.displayPage(activeFocusPage)
	case BreakTimer:
		m.displayPage(activeBreakPage)
	}

	go m.updateUIWithTicker(stopSignal)
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

func (m *UIManager) handleStatePaused(timerType TimerType) {
	go playClickSound()

	switch timerType {
	case FocusTimer:
		m.renderPausePageForTimer(FocusTimer)
	case BreakTimer:
		m.renderPausePageForTimer(BreakTimer)
	}
}

func (m *UIManager) handleStateFinished(timerType TimerType) {
	go playEndSound()

	switch timerType {
	case FocusTimer:
		m.renderPausePageForTimer(BreakTimer)
	case BreakTimer:
		m.renderPausePageForTimer(FocusTimer)
	}
}

func (m *UIManager) renderPausePageForTimer(timerType TimerType) {
	switch timerType {
	case BreakTimer:
		m.renderPausePage(pauseBreakPage, "Break", BreakTimer)
		m.displayPage(pauseBreakPage)
	case FocusTimer:
		m.renderPausePage(pauseFocusPage, "Pomodoro", FocusTimer)
		m.displayPage(pauseFocusPage)
	}
}

func (m *UIManager) displayPage(pageName string) {
	m.ui.QueueUpdateDraw(func() {
		m.pages.SwitchToPage(pageName)
	})
}

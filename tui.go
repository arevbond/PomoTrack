package main

import (
	"log/slog"
	"time"

	"github.com/arevbond/PomoTrack/config"

	"github.com/gdamore/tcell/v2"

	"github.com/rivo/tview"
)

const screenRefreshInterval = 1 * time.Second

type pomodoroTracker interface {
	HandlePomodoroStateChanges()
	TodayPomodoros() ([]*Pomodoro, error)
	Pomodoros() ([]*Pomodoro, error)
	RemovePomodoro(id int) error
	CreateNewPomodoro(startAt time.Time, finishAt time.Time, duration int) (*Pomodoro, error)
	Hours([]*Pomodoro) float64
	CountDays([]*Pomodoro) int
	HoursInWeek([]*Pomodoro) [7]int
	FinishRunningPomodoro()
}

type UIManager struct {
	ui           *tview.Application
	pages        *tview.Pages
	logger       *slog.Logger
	stateManager *StateManager
	stateUpdates chan StateEvent

	pomodoroTracker      pomodoroTracker
	statePomodoroUpdates chan StateEvent

	allowedTransitions map[string][]string
	keyPageMapping     map[tcell.Key]string
}

type StateEvent struct {
	TimerType TimerType
	NewState  TimerState
}

func NewUIManager(logger *slog.Logger, cfg *config.Config, events chan StateEvent, tm pomodoroTracker) *UIManager {
	stateChangeChan := make(chan StateEvent)
	focusTimer := NewFocusTimer(cfg.Timer.FocusDuration)
	breakTimer := NewBreakTimer(cfg.Timer.BreakDuration)

	return &UIManager{
		ui:                   tview.NewApplication(),
		pages:                tview.NewPages(),
		logger:               logger,
		stateManager:         NewStateManager(logger, focusTimer, breakTimer, stateChangeChan, cfg.Timer),
		stateUpdates:         stateChangeChan,
		pomodoroTracker:      tm,
		statePomodoroUpdates: events,
		allowedTransitions:   constructAllowedTransitions(),
		keyPageMapping:       constructKeyPageMap(),
	}
}

func constructAllowedTransitions() map[string][]string {
	return map[string][]string{
		pauseFocusPage:   {detailStatsPage, pauseBreakPage, summaryStatsPage},
		pauseBreakPage:   {detailStatsPage, pauseFocusPage, summaryStatsPage},
		detailStatsPage:  {pauseFocusPage, pauseBreakPage, insertStatsPage, summaryStatsPage},
		insertStatsPage:  {detailStatsPage},
		summaryStatsPage: {pauseFocusPage, pauseBreakPage, insertStatsPage, detailStatsPage},
	}
}

func constructKeyPageMap() map[tcell.Key]string {
	return map[tcell.Key]string{
		tcell.KeyF1:    pauseFocusPage,
		tcell.KeyF2:    pauseBreakPage,
		tcell.KeyF3:    detailStatsPage,
		tcell.KeyCtrlI: insertStatsPage,
		tcell.KeyF4:    summaryStatsPage,
	}
}

func (m *UIManager) DefaultTimerPages() {
	m.renderActivePage(activeFocusPage, "red", "Pomodoro", FocusTimer)
	m.renderActivePage(activeBreakPage, "green", "Break", BreakTimer)
	m.renderPausePage(pauseBreakPage, "Break", BreakTimer)
	m.renderPausePage(pauseFocusPage, "Pomodoro", FocusTimer)
}

func (m *UIManager) setKeyboardEvents() {
	m.ui.SetInputCapture(m.keyboardEvents)
}

func (m *UIManager) canSwitchTo(targetPage string) bool {
	curPage, _ := m.pages.GetFrontPage()

	allowedPages, ok := m.allowedTransitions[targetPage]
	if !ok {
		return false
	}
	for _, allowedPage := range allowedPages {
		if curPage == allowedPage {
			return true
		}
	}
	return false
}

func (m *UIManager) keyboardEvents(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyCtrlC {
		m.pomodoroTracker.FinishRunningPomodoro()
		m.ui.Stop()
	}

	targetPage, exists := m.keyPageMapping[event.Key()]
	if !exists || !m.canSwitchTo(targetPage) {
		return event
	}

	switch targetPage {
	case pauseFocusPage, pauseBreakPage:
		m.pages.SwitchToPage(targetPage)
	case detailStatsPage, insertStatsPage:
		m.switchToDetailStats(targetPage)
	case summaryStatsPage:
		m.switchToSummaryStats()
	default:
		return event
	}
	return nil
}

func (m *UIManager) switchToDetailStats(pageName string) {
	pomodoros, err := m.pomodoroTracker.TodayPomodoros()
	if err != nil {
		m.logger.Error("can't get today pomodoros", slog.Any("error", err))
		return
	}
	if len(pomodoros) > statisticsPageSize {
		m.renderInsertStatsPage(0, statisticsPageSize, pomodoros)
		m.renderDetailStatsPage(0, statisticsPageSize, pomodoros)
	} else {
		m.renderInsertStatsPage(0, len(pomodoros), pomodoros)
		m.renderDetailStatsPage(0, len(pomodoros), pomodoros)
	}
	m.pages.SwitchToPage(pageName)
}

func (m *UIManager) switchToSummaryStats() {
	pomodoros, err := m.pomodoroTracker.Pomodoros()
	if err != nil {
		m.logger.Error("can't get all pomodoros", slog.Any("error", err))
		return
	}

	m.renderSummaryStatsPage(
		m.pomodoroTracker.Hours(pomodoros),
		m.pomodoroTracker.CountDays(pomodoros),
		m.pomodoroTracker.HoursInWeek(pomodoros),
	)

	m.pages.SwitchToPage(summaryStatsPage)
}

func (m *UIManager) InitStateAndKeyboardHandling() {
	stopRefreshing := make(chan struct{})
	m.setKeyboardEvents()
	go m.listenToStateChanges(stopRefreshing)
}

func (m *UIManager) listenToStateChanges(stopRefreshing chan struct{}) {
	for event := range m.stateUpdates {
		m.statePomodoroUpdates <- event

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

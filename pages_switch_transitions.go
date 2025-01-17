package main

import (
	"log/slog"
	"time"

	"github.com/gdamore/tcell/v2"
)

func constructAllowedTransitions() map[string][]string {
	return map[string][]string{
		pauseFocusPage:   {detailStatsPage, pauseBreakPage, summaryStatsPage, allTasksPage},
		pauseBreakPage:   {detailStatsPage, pauseFocusPage, summaryStatsPage, allTasksPage},
		detailStatsPage:  {pauseFocusPage, pauseBreakPage, insertStatsPage, summaryStatsPage, allTasksPage},
		insertStatsPage:  {detailStatsPage},
		summaryStatsPage: {pauseFocusPage, pauseBreakPage, insertStatsPage, detailStatsPage, allTasksPage},
		allTasksPage:     {pauseFocusPage, pauseBreakPage, detailStatsPage, summaryStatsPage},
	}
}

func constructKeyPageMap() map[tcell.Key]string {
	return map[tcell.Key]string{
		tcell.KeyF1:    pauseFocusPage,
		tcell.KeyF2:    pauseBreakPage,
		tcell.KeyF3:    detailStatsPage,
		tcell.KeyCtrlI: insertStatsPage,
		tcell.KeyF4:    summaryStatsPage,
		tcell.KeyF5:    allTasksPage,
	}
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
	case pauseFocusPage:
		m.AddPageAndSwitch(m.renderPausePage(pauseFocusPage, "Pomodoro", FocusTimer).WithBottomPanel())
	case pauseBreakPage:
		m.AddPageAndSwitch(m.renderPausePage(pauseBreakPage, "Break", BreakTimer).WithBottomPanel())
	case detailStatsPage:
		m.switchToDetailStats()
	case insertStatsPage:
		m.switchToIntesrStats()
	case summaryStatsPage:
		m.switchToSummaryStats()
	case allTasksPage:
		m.switchToTasks()
	default:
		return event
	}
	return nil
}

func (m *UIManager) switchToDetailStats() {
	pomodoros, err := m.pomodoroTracker.TodayPomodoros()
	if err != nil {
		m.logger.Error("can't get today pomodoros", slog.Any("error", err))
		return
	}

	var page *PageComponent
	if len(pomodoros) > statisticsPageSize {
		page = m.renderDetailStatsPage(0, statisticsPageSize, pomodoros)
	} else {
		page = m.renderDetailStatsPage(0, len(pomodoros), pomodoros)
	}

	m.AddPageAndSwitch(page.WithBottomPanel())
}

func (m *UIManager) switchToIntesrStats() {
	pomodoros, err := m.pomodoroTracker.TodayPomodoros()
	if err != nil {
		m.logger.Error("can't get today pomodoros", slog.Any("error", err))
		return
	}

	var page *PageComponent
	if len(pomodoros) > statisticsPageSize {
		page = m.renderInsertStatsPage(0, statisticsPageSize, pomodoros)
	} else {
		page = m.renderInsertStatsPage(0, len(pomodoros), pomodoros)
	}

	m.AddPageAndSwitch(page)
}

func (m *UIManager) switchToTasks() {
	tasks, err := m.taskTracker.Tasks()
	if err != nil {
		m.logger.Error("can't get tasks", slog.String("func", "renderAllTasls"), slog.Any("error", err))
		return
	}
	tasksPage := m.renderAllTasksPage(tasks)
	m.AddPageAndSwitch(tasksPage.WithBottomPanel())
}

func (m *UIManager) switchToSummaryStats() {
	pomodoros, err := m.pomodoroTracker.Pomodoros()
	if err != nil {
		m.logger.Error("can't get all pomodoros", slog.Any("error", err))
		return
	}

	summaryStatsPage := m.renderSummaryStatsPage(
		m.pomodoroTracker.Hours(pomodoros),
		m.pomodoroTracker.CountDays(pomodoros),
		m.pomodoroTracker.HoursInWeek(pomodoros),
	)

	m.AddPageAndSwitch(summaryStatsPage.WithBottomPanel())
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
			m.switchToActive(event.TimerType, stopRefreshing)

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

func (m *UIManager) switchToActive(timerType TimerType, stopSignal chan struct{}) {
	go playClickSound()

	switch timerType {
	case FocusTimer:
		page := m.renderActivePage(activeFocusPage, "red", "Pomodoro", FocusTimer)
		m.ui.QueueUpdateDraw(func() {
			m.AddPageAndSwitch(page)
		})
	case BreakTimer:
		page := m.renderActivePage(activeBreakPage, "green", "Break", BreakTimer)
		m.ui.QueueUpdateDraw(func() {
			m.AddPageAndSwitch(page)
		})
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
		m.switchToBreak(FocusTimer)
	case BreakTimer:
		m.switchToBreak(BreakTimer)
	}
}

func (m *UIManager) handleStateFinished(timerType TimerType) {
	go playEndSound()

	switch timerType {
	case FocusTimer:
		m.switchToBreak(BreakTimer)
	case BreakTimer:
		m.switchToBreak(FocusTimer)
	}
}

func (m *UIManager) switchToBreak(timerType TimerType) {
	switch timerType {
	case BreakTimer:
		page := m.renderPausePage(pauseBreakPage, "Break", BreakTimer).WithBottomPanel()
		m.ui.QueueUpdateDraw(func() {
			m.AddPageAndSwitch(page)
		})
	case FocusTimer:
		page := m.renderPausePage(pauseFocusPage, "Pomodoro", FocusTimer).WithBottomPanel()
		m.ui.QueueUpdateDraw(func() {
			m.AddPageAndSwitch(page)
		})
	}
}

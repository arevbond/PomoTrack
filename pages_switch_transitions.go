package main

import (
	"github.com/gdamore/tcell/v2"
)

type PageName string

const (
	activeFocusPage PageName = "Active-Focus"
	pauseFocusPage  PageName = "Stop-Focus"

	activeBreakPage PageName = "Active-Break"
	pauseBreakPage  PageName = "Stop-Break"

	allTasksPage PageName = "All-Tasks-Pages"

	detailStatsPage PageName = "Detail-Statistics"
	insertStatsPage PageName = "Insert-Statistics"

	summaryStatsPage PageName = "Summary-Statistics"
)

func constructAllowedTransitions() map[PageName][]PageName {
	return map[PageName][]PageName{
		pauseFocusPage:   {detailStatsPage, pauseBreakPage, summaryStatsPage, allTasksPage},
		pauseBreakPage:   {detailStatsPage, pauseFocusPage, summaryStatsPage, allTasksPage},
		detailStatsPage:  {pauseFocusPage, pauseBreakPage, insertStatsPage, summaryStatsPage, allTasksPage},
		summaryStatsPage: {pauseFocusPage, pauseBreakPage, insertStatsPage, detailStatsPage, allTasksPage},
		allTasksPage:     {pauseFocusPage, pauseBreakPage, detailStatsPage, summaryStatsPage},
	}
}

func (m *UIManager) constructKeyPageMap() map[tcell.Key]*Page {
	pauseFocus := m.NewPausePage(FocusTimer)
	pauseBreak := m.NewPausePage(BreakTimer)
	tasksPage := m.NewTasksPage()
	summaryPage := m.NewSummaryPage()
	detailPage := m.NewDetailStats(-1, -1)

	return map[tcell.Key]*Page{
		tcell.KeyF1: pauseFocus,
		tcell.KeyF2: pauseBreak,
		tcell.KeyF3: tasksPage,
		tcell.KeyF4: summaryPage,
		tcell.KeyF5: detailPage,
	}
}

func (m *UIManager) setKeyboardEvents() {
	m.ui.SetInputCapture(m.keyboardEvents)
}

func (m *UIManager) canSwitchTo(targetPage PageName) bool {
	curPage, _ := m.pages.GetFrontPage()

	allowedPages, ok := m.allowedTransitions[targetPage]
	if !ok {
		return false
	}
	for _, allowedPage := range allowedPages {
		if PageName(curPage) == allowedPage {
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
	if !exists || !m.canSwitchTo(targetPage.name) {
		return event
	}

	// need for reload new pomodoros
	if targetPage.name == detailStatsPage {
		targetPage = m.NewDetailStats(-1, -1)
	}

	m.AddPageAndSwitch(targetPage)
	return nil
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
			m.ui.QueueUpdateDraw(func() {
				m.AddPageAndSwitch(m.NewActivePage(event.TimerType, stopRefreshing))
			})

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

package main

import (
	"log/slog"

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
		insertStatsPage:  {detailStatsPage},
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
	detailInsertPage := m.NewInsertDetailPage(-1, -1)

	return map[tcell.Key]*Page{
		tcell.KeyF1:    pauseFocus,
		tcell.KeyF2:    pauseBreak,
		tcell.KeyF3:    tasksPage,
		tcell.KeyF4:    summaryPage,
		tcell.KeyF5:    detailPage,
		tcell.KeyCtrlI: detailInsertPage,
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

	m.AddPageAndSwitch(targetPage)
	return nil
}

func (m *UIManager) NewDetailStats(start, end int) *Page {
	pomodoros, err := m.pomodoroTracker.Pomodoros()
	if err != nil {
		m.logger.Error("can't get pomodoros", slog.Any("error", err))
		return nil
	}

	if start <= -1 && end <= -1 {
		if len(pomodoros) > statisticsPageSize {
			start = 0
			end = statisticsPageSize
		} else {
			start = 0
			end = len(pomodoros)
		}
	}

	return NewPageComponent(detailStatsPage, true, false, m.renderDetailStatsPage(start, end, pomodoros))
}

func (m *UIManager) NewInsertDetailPage(start, end int) *Page {
	pomodoros, err := m.pomodoroTracker.Pomodoros()
	if err != nil {
		m.logger.Error("can't get pomodoros", slog.Any("error", err))
		return nil
	}

	if start <= -1 && end <= -1 {
		if len(pomodoros) > statisticsPageSize {
			start = 0
			end = statisticsPageSize
		} else {
			start = 0
			end = len(pomodoros)
		}
	}

	return NewPageComponent(insertStatsPage, true, false, m.renderInsertStatsPage(start, end, pomodoros))
}

func (m *UIManager) NewTasksPage() *Page {
	tasks, err := m.taskTracker.Tasks()
	if err != nil {
		m.logger.Error("can't get tasks", slog.String("func", "renderAllTasls"), slog.Any("error", err))
		return nil
	}
	renderFunc := m.renderAllTasksPage(tasks)
	return NewPageComponent(allTasksPage, true, false, renderFunc)
}

func (m *UIManager) NewSummaryPage() *Page {
	pomodoros, err := m.pomodoroTracker.Pomodoros()
	if err != nil {
		m.logger.Error("can't get all pomodoros", slog.Any("error", err))
		return nil
	}

	render := m.renderSummaryStatsPage(
		m.pomodoroTracker.Hours(pomodoros),
		m.pomodoroTracker.CountDays(pomodoros),
		m.pomodoroTracker.HoursInWeek(pomodoros),
	)

	return NewPageComponent(summaryStatsPage, true, false, render)
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

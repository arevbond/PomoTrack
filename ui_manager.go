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

type taskTracker interface {
	Tasks() ([]*Task, error)
	CreateTask(task *Task) error
	DeleteTask(id int) error
}

type UIManager struct {
	ui     *tview.Application
	pages  *tview.Pages
	logger *slog.Logger

	stateManager *StateManager
	stateUpdates chan StateEvent

	pomodoroTracker      pomodoroTracker
	statePomodoroUpdates chan StateEvent

	taskTracker taskTracker

	allowedTransitions map[PageName][]PageName
	keyPageMapping     map[tcell.Key]*Page
}

type StateEvent struct {
	TimerType TimerType
	NewState  TimerState
}

func NewUIManager(l *slog.Logger, c *config.Config, e chan StateEvent, tm pomodoroTracker, tt taskTracker) *UIManager {
	stateChangeChan := make(chan StateEvent)
	focusTimer := NewFocusTimer(c.Timer.FocusDuration)
	breakTimer := NewBreakTimer(c.Timer.BreakDuration)

	m := &UIManager{
		ui:                   tview.NewApplication(),
		pages:                tview.NewPages(),
		logger:               l,
		stateManager:         NewStateManager(l, focusTimer, breakTimer, stateChangeChan, c.Timer),
		stateUpdates:         stateChangeChan,
		pomodoroTracker:      tm,
		statePomodoroUpdates: e,
		allowedTransitions:   constructAllowedTransitions(),
		taskTracker:          tt,
		keyPageMapping:       nil,
	}
	m.keyPageMapping = m.constructKeyPageMap()

	return m
}

func (m *UIManager) DefaultPage() {
	pauseFocus := m.NewPausePage(FocusTimer)
	m.AddPageAndSwitch(pauseFocus)
}

func (m *UIManager) AddPageAndSwitch(page *Page) {
	m.pages.AddAndSwitchToPage(string(page.name), page.WithBottomPanel(), page.resize)
}

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

	allowedTransitions map[string][]string
	keyPageMapping     map[tcell.Key]string
}

type StateEvent struct {
	TimerType TimerType
	NewState  TimerState
}

func NewUIManager(l *slog.Logger, c *config.Config, e chan StateEvent, tm pomodoroTracker, tt taskTracker) *UIManager {
	stateChangeChan := make(chan StateEvent)
	focusTimer := NewFocusTimer(c.Timer.FocusDuration)
	breakTimer := NewBreakTimer(c.Timer.BreakDuration)

	return &UIManager{
		ui:                   tview.NewApplication(),
		pages:                tview.NewPages(),
		logger:               l,
		stateManager:         NewStateManager(l, focusTimer, breakTimer, stateChangeChan, c.Timer),
		stateUpdates:         stateChangeChan,
		pomodoroTracker:      tm,
		statePomodoroUpdates: e,
		allowedTransitions:   constructAllowedTransitions(),
		keyPageMapping:       constructKeyPageMap(),
		taskTracker:          tt,
	}
}

func (m *UIManager) DefaultPage() {
	m.AddPageAndSwitch(m.renderPausePage(pauseFocusPage, "Pomodoro", FocusTimer).WithBottomPanel())
}

func (m *UIManager) AddPageAndSwitch(page *PageComponent) {
	m.pages.AddAndSwitchToPage(page.name, page.item, page.resize)
}

func (m *UIManager) AddPageComponent(page *PageComponent) {
	m.pages.AddPage(page.name, page.item, page.resize, page.visible)
}

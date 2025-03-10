package main

import (
	"log/slog"
	"time"

	"github.com/arevbond/PomoTrack/config"
)

type TimerState int

const (
	StatePaused TimerState = iota
	StateActive
	StateFinished
)

type StateManager struct {
	currentState TimerState
	focusTimer   *Timer
	breakTimer   *Timer
	stateChan    chan StateEvent
	logger       *slog.Logger
	timerConfig  config.TimerConfig
	taskManager  taskManager
}

func NewStateManager(l *slog.Logger, focusT *Timer, breakT *Timer,
	stateChan chan StateEvent, cfg config.TimerConfig, manager taskManager) *StateManager {
	return &StateManager{
		logger:       l,
		currentState: StatePaused,
		focusTimer:   focusT,
		breakTimer:   breakT,
		stateChan:    stateChan,
		timerConfig:  cfg,
		taskManager:  manager,
	}
}

func (sm *StateManager) CurrentState() TimerState {
	return sm.currentState
}

func (sm *StateManager) SetState(state TimerState, timerType TimerType) {
	if sm.currentState == state {
		return
	}

	sm.currentState = state

	event := StateEvent{
		TimerType: timerType,
		NewState:  state,
	}
	sm.stateChan <- event

	timer := sm.getTimer(timerType)
	if nil == timer {
		return
	}

	switch state {
	case StateActive:
		sm.startTimer(timer)
	case StatePaused:
		sm.pauseTimer(timer)
	case StateFinished:
		if timerType == FocusTimer {
			sm.completePomodoro()
		}
		sm.finishTimer(timer)
	}
}

func (sm *StateManager) completePomodoro() {
	err := sm.taskManager.IncPomodoroActiveTask()
	if err != nil {
		sm.logger.Error("can't increment pomodoro in active task", slog.Any("error", err))
	}
}

func (sm *StateManager) pauseTimer(timer *Timer) {
	timer.Stop()
}

func (sm *StateManager) startTimer(timer *Timer) {
	finishChan := timer.Run()

	go func() {
		_, ok := <-finishChan
		if ok {
			sm.SetState(StateFinished, timer.timerType)
		}
	}()
}

func (sm *StateManager) finishTimer(timer *Timer) {
	timer.Stop()

	var duration time.Duration
	switch timer.timerType {
	case FocusTimer:
		duration = sm.timerConfig.FocusDuration
	case BreakTimer:
		duration = sm.timerConfig.BreakDuration
	}

	timer.Reset(duration)
}

func (sm *StateManager) getTimer(timerType TimerType) *Timer {
	switch timerType {
	case FocusTimer:
		return sm.focusTimer
	case BreakTimer:
		return sm.breakTimer
	}
	return nil
}

func (sm *StateManager) timeToFinish(timerType TimerType) time.Duration {
	return sm.getTimer(timerType).timeToFinish
}

func (sm *StateManager) IsFocusTimeHidden() bool {
	return sm.timerConfig.HiddenFocusTime
}

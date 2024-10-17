package main

import (
	"log/slog"
	"time"
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
	stateChan    chan StateChangeEvent
	logger       *slog.Logger
}

func NewStateManager(l *slog.Logger, focusT *Timer, breakT *Timer, stateChan chan StateChangeEvent) *StateManager {
	return &StateManager{
		logger:       l,
		currentState: StatePaused,
		focusTimer:   focusT,
		breakTimer:   breakT,
		stateChan:    stateChan,
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

	event := StateChangeEvent{
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
		sm.finishTimer(timer)
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
		duration = focusDuration
	case BreakTimer:
		duration = breakDuration
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

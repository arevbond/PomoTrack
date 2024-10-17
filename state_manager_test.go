package main

import (
	"log/slog"
	"testing"
	"time"

	"github.com/arevbond/PomoTrack/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	focusDuration = 10 * time.Second
	breakDuration = 3 * time.Second
)

func TestNewStateManager(t *testing.T) {
	focusTimer, breakTimer := NewFocusTimer(focusDuration), NewBreakTimer(breakDuration)
	stateChan := make(chan StateChangeEvent)
	stateManager := NewStateManager(slog.Default(), focusTimer, breakTimer, stateChan,
		config.TimerConfig{FocusDuration: focusDuration, BreakDuration: breakDuration})
	require.Equal(t, StatePaused, stateManager.CurrentState())
}

func TestStateManager_SetState(t *testing.T) {
	focusTimer, breakTimer := NewFocusTimer(focusDuration), NewBreakTimer(breakDuration)
	stateChan := make(chan StateChangeEvent)
	stateManager := NewStateManager(slog.Default(), focusTimer, breakTimer, stateChan,
		config.TimerConfig{FocusDuration: focusDuration, BreakDuration: breakDuration})

	go func() {
		event := <-stateChan
		assert.Equal(t, StateActive, event.NewState)
		assert.Equal(t, FocusTimer, event.TimerType)
	}()

	stateManager.SetState(StateActive, FocusTimer)

	assert.Equal(t, StateActive, stateManager.CurrentState())

	focusTimer.Stop()
}

func TestStateManager_StartTimer(t *testing.T) {
	focusTimer, breakTimer := NewFocusTimer(2*time.Second), NewBreakTimer(breakDuration)
	stateChan := make(chan StateChangeEvent)
	stateManager := NewStateManager(slog.Default(), focusTimer, breakTimer, stateChan,
		config.TimerConfig{FocusDuration: focusDuration, BreakDuration: breakDuration})

	stateManager.startTimer(focusTimer)
	time.Sleep(3 * time.Second)
	assert.Equal(t, StateFinished, stateManager.CurrentState())
}

func TestStateManager_PauseTimer(t *testing.T) {
	focusTimer, breakTimer := NewFocusTimer(10*time.Second), NewBreakTimer(breakDuration)
	stateChan := make(chan StateChangeEvent)
	stateManager := NewStateManager(slog.Default(), focusTimer, breakTimer, stateChan,
		config.TimerConfig{FocusDuration: focusDuration, BreakDuration: breakDuration})

	stateManager.focusTimer.Run()
	time.Sleep(1 * time.Second)
	stateManager.pauseTimer(stateManager.focusTimer)
	time.Sleep(2 * time.Second)

	assert.Equal(t, 9*time.Second, focusTimer.timeToFinish)
}

func TestStateManager_FinishTimer(t *testing.T) {
	focusTimer, breakTimer := NewFocusTimer(focusDuration), NewBreakTimer(breakDuration)
	stateChan := make(chan StateChangeEvent)
	stateManager := NewStateManager(slog.Default(), focusTimer, breakTimer, stateChan,
		config.TimerConfig{FocusDuration: focusDuration, BreakDuration: breakDuration})

	stateManager.focusTimer.Run()
	time.Sleep(3 * time.Second)
	stateManager.finishTimer(stateManager.focusTimer)
	time.Sleep(2 * time.Second)

	assert.Equal(t, focusDuration, focusTimer.timeToFinish)
}

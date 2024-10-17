package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewFocusTimer(t *testing.T) {
	focusTimer := NewFocusTimer(focusDuration)

	assert.Equal(t, FocusTimer, focusTimer.timerType)
	assert.Equal(t, focusDuration, focusTimer.timeToFinish)
}

func TestNewBreakTimer(t *testing.T) {
	breakTimer := NewBreakTimer(breakDuration)

	assert.Equal(t, BreakTimer, breakTimer.timerType)
	assert.Equal(t, breakDuration, breakTimer.timeToFinish)
}

func TestTimer_Stop(t *testing.T) {
	timer := NewFocusTimer(3 * time.Second)
	doneChan := timer.Run()

	time.Sleep(1 * time.Second)

	timer.Stop()
	remainingTime := timer.TimeToFinish()
	time.Sleep(1 * time.Second)

	assert.Equal(t, remainingTime, timer.TimeToFinish())

	select {
	case <-doneChan:
		// OK
	case <-time.After(1 * time.Second):
		t.Errorf("Expected timer goroutine to stop, but it did not")
	}
}

func TestTimer_Run(t *testing.T) {
	timer := NewFocusTimer(1 * time.Second)
	doneChan := timer.Run()

	select {
	case <-doneChan:
		// OK
	case <-time.After(2 * time.Second):
		assert.Errorf(t, nil, "Expected timer to stop, but it did not")
	}
}

func TestTimer_TimeToFinish(t *testing.T) {
	timer := NewFocusTimer(5 * time.Second)
	assert.Equal(t, 5*time.Second, timer.timeToFinish)
	assert.Equal(t, timer.timeToFinish, timer.TimeToFinish())
}

func TestTimer_Reset(t *testing.T) {
	timer := NewFocusTimer(10 * time.Second)
	timer.Reset(5 * time.Second)
	assert.Equal(t, 5*time.Second, timer.timeToFinish)
}

func TestTimer_Tick(t *testing.T) {
	timer := NewFocusTimer(5 * time.Second)
	timer.tick(1 * time.Second)
	assert.Equal(t, 4*time.Second, timer.timeToFinish)

	timer.tick(5 * time.Second)
	assert.Equal(t, 0*time.Second, timer.timeToFinish)
}

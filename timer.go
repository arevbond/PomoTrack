package main

import (
	"sync"
	"time"
)

type TimerType int

const (
	FocusTimer TimerType = iota
	BreakTimer
)

type Timer struct {
	timerType    TimerType
	mu           sync.RWMutex
	timeToFinish time.Duration
	stopSignal   chan struct{}
}

func NewFocusTimer(duration time.Duration) *Timer {
	return &Timer{
		timerType:    FocusTimer,
		mu:           sync.RWMutex{},
		timeToFinish: duration,
		stopSignal:   make(chan struct{}),
	}
}

func NewBreakTimer(duration time.Duration) *Timer {
	return &Timer{
		timerType:    BreakTimer,
		mu:           sync.RWMutex{},
		timeToFinish: duration,
		stopSignal:   make(chan struct{}),
	}
}

func (t *Timer) Stop() {
	select {
	case t.stopSignal <- struct{}{}:
	default:
	}
}

func (t *Timer) TimeToFinish() time.Duration {
	t.mu.RLock()
	duration := t.timeToFinish
	t.mu.RUnlock()
	return duration
}

func (t *Timer) Reset(duration time.Duration) {
	t.mu.Lock()
	t.timeToFinish = duration
	t.mu.Unlock()
}

func (t *Timer) Run() chan struct{} {
	doneChan := make(chan struct{})

	const timeToStop = 0 * time.Second
	go t.run(doneChan, timeToStop)

	return doneChan
}

func (t *Timer) run(doneChan chan struct{}, timeToStop time.Duration) {
	defer close(doneChan)

	const timerDelta = 1 * time.Second
	tick := time.NewTicker(timerDelta)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			t.tick(timerDelta)

			if t.TimeToFinish() <= timeToStop {
				doneChan <- struct{}{}
				return
			}
		case <-t.stopSignal:
			return
		}
	}
}

func (t *Timer) tick(duration time.Duration) {
	t.mu.Lock()
	if t.timeToFinish > duration {
		t.timeToFinish -= duration
	} else {
		t.timeToFinish = 0 * time.Second
	}
	t.mu.Unlock()
}

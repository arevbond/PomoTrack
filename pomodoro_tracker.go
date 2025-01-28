package main

import (
	"fmt"
	"log/slog"
	"time"
)

type Pomodoro struct {
	ID              int       `db:"id"`
	StartAt         time.Time `db:"start_at"`
	FinishAt        time.Time `db:"finish_at"`
	SecondsDuration int       `db:"duration"`

	// don't save to db; need for app logic
	lastStartAt time.Time
	finished    bool
}

type PomodoroManager struct {
	storage           *Storage
	logger            *slog.Logger
	currentPomodoro   *Pomodoro
	statePomodoroChan chan StateEvent
}

func NewPomodoroManager(logger *slog.Logger, storage *Storage, stateEvents chan StateEvent) *PomodoroManager {
	return &PomodoroManager{
		storage:           storage,
		logger:            logger,
		statePomodoroChan: stateEvents,
		currentPomodoro:   nil,
	}
}

func (tm *PomodoroManager) HandlePomodoroStateChanges() {
	for event := range tm.statePomodoroChan {
		if event.TimerType == FocusTimer {
			switch event.NewState {
			case StateActive:
				tm.handleStartPomodoro()
			case StatePaused:
				tm.handlePausePomodoro()
			case StateFinished:
				tm.handleFinishPomodoro()
			}
		}
	}
}

func (tm *PomodoroManager) Pomodoros() ([]*Pomodoro, error) {
	return tm.storage.GetPomodoros()
}

func (tm *PomodoroManager) TodayPomodoros() ([]*Pomodoro, error) {
	return tm.storage.GetTodayPomodoros()
}

func (tm *PomodoroManager) RemovePomodoro(id int) error {
	return tm.storage.RemovePomodoro(id)
}

func (tm *PomodoroManager) CreateNewPomodoro(startAt time.Time, finishAt time.Time, duration int) (*Pomodoro, error) {
	pomodoro := &Pomodoro{
		ID:              0,
		StartAt:         startAt,
		FinishAt:        finishAt,
		SecondsDuration: duration,
		lastStartAt:     time.Now(),
		finished:        false,
	}

	err := tm.storage.CreatePomodoro(pomodoro)
	if err != nil {
		return nil, fmt.Errorf("can't create pomodoro: %w", err)
	}
	return pomodoro, nil
}

func (tm *PomodoroManager) Hours(pomodoros []*Pomodoro) float64 {
	var result time.Duration
	for _, pomodoro := range pomodoros {
		result += time.Duration(pomodoro.SecondsDuration) * time.Second
	}
	return result.Hours()
}

func (tm *PomodoroManager) CountDays(pomodoros []*Pomodoro) int {
	if len(pomodoros) == 0 {
		return 0
	}

	count := 1
	for i := 1; i < len(pomodoros); i++ {
		prev := pomodoros[i-1]
		cur := pomodoros[i]

		if prev.StartAt.Day() != cur.StartAt.Day() {
			count++
		}
	}

	return count
}

func (tm *PomodoroManager) HoursInWeek(pomodoros []*Pomodoro) [7]int {
	if len(pomodoros) == 0 {
		return [7]int{}
	}

	var weekdayHours [7]int
	firstDayIndx := time.Now().Day() - int(time.Now().Weekday()) + 1
	lastDayIndx := time.Now().Day()
	for _, pomodoro := range pomodoros {
		if pomodoro.StartAt.Day() >= firstDayIndx && pomodoro.StartAt.Day() <= lastDayIndx {
			weekdayHours[(pomodoro.FinishAt.Weekday()+6)%7] += pomodoro.SecondsDuration / 3600
		}
	}

	return weekdayHours
}

func (tm *PomodoroManager) FinishRunningPomodoro() {
	if tm.currentPomodoro != nil && !tm.currentPomodoro.finished {
		tm.handleFinishPomodoro()
	}
}

func (tm *PomodoroManager) handleStartPomodoro() {
	// предыдущая задача не создана или завершена
	// создаём новую пустую задачу
	if tm.currentPomodoro == nil || tm.currentPomodoro.finished {
		newPomodoro, err := tm.CreateNewPomodoro(time.Now(), time.Now(), 0)
		if err != nil {
			tm.logger.Error("handle start pomodoro", slog.Any("error", err))
		}
		tm.currentPomodoro = newPomodoro
		return
	}

	// есть текущая незавершённая задача (запуск после паузы)
	tm.currentPomodoro.lastStartAt = time.Now()
}

func (tm *PomodoroManager) handlePausePomodoro() {
	err := tm.updateCurrentPomodoroDuration()
	if err != nil {
		tm.logger.Error("handle start pomodoro", slog.Any("error", err))
	}
}

func (tm *PomodoroManager) handleFinishPomodoro() {
	tm.currentPomodoro.finished = true
	err := tm.updateCurrentPomodoroDuration()
	if err != nil {
		tm.logger.Error("handle start pomodoro", slog.Any("error", err))
	}
}

func (tm *PomodoroManager) updateCurrentPomodoroDuration() error {
	duration := int(time.Since(tm.currentPomodoro.lastStartAt).Seconds())

	tm.currentPomodoro.SecondsDuration += duration
	tm.currentPomodoro.FinishAt = time.Now()

	err := tm.storage.UpdatePomodoro(tm.currentPomodoro)
	if err != nil {
		return fmt.Errorf("can't update pomdooro: %w", err)
	}
	return nil
}

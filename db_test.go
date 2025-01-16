//nolint:exhaustruct,gochecknoglobals,errcheck // still immature code, work in progress
package main

import (
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

const testDBName = ":memory:"

var s *Storage

func TestMain(m *testing.M) {
	err := setup()
	if err != nil {
		log.Fatal(err)
	}

	exitVal := m.Run()

	os.Exit(exitVal)
}

func setup() error {
	var err error
	s, err = NewStorage(testDBName, slog.Default())
	if err != nil {
		return err
	}

	err = s.Migrate()
	if err != nil {
		return fmt.Errorf("can't make migrations: %w", err)
	}
	return nil
}

func TestStorage_CreatePomodoro(t *testing.T) {
	pomodoro := &Pomodoro{
		ID:       100,
		StartAt:  time.Now(),
		FinishAt: time.Now(),
	}
	err := s.CreatePomodoro(pomodoro)
	require.NoError(t, err)

	var pomdoroFromDB Pomodoro
	err = s.DB.QueryRow(`SELECT * FROM pomodoros;`).Scan(&pomdoroFromDB.ID, &pomdoroFromDB.StartAt,
		&pomdoroFromDB.FinishAt, &pomdoroFromDB.SecondsDuration)
	require.NoError(t, err)

	require.Equal(t, pomodoro.ID, pomdoroFromDB.ID)
	require.WithinDuration(t, pomodoro.StartAt, pomdoroFromDB.StartAt, time.Second)
	require.WithinDuration(t, pomodoro.FinishAt, pomdoroFromDB.FinishAt, time.Second)
	require.Equal(t, pomodoro.SecondsDuration, pomdoroFromDB.SecondsDuration)

	clearTable()
}

func TestStorage_UpdatePomodoro(t *testing.T) {
	pomodoro := &Pomodoro{
		ID:       100,
		StartAt:  time.Now(),
		FinishAt: time.Now(),
	}
	err := s.CreatePomodoro(pomodoro)
	require.NoError(t, err)

	pomodoro.FinishAt = time.Now().Add(60 * time.Second)
	pomodoro.SecondsDuration = 60

	err = s.UpdatePomodoro(pomodoro)
	require.NoError(t, err)

	var pomodoroFromDB Pomodoro
	err = s.DB.QueryRow(`SELECT * FROM pomodoros;`).Scan(&pomodoroFromDB.ID, &pomodoroFromDB.StartAt,
		&pomodoroFromDB.FinishAt, &pomodoroFromDB.SecondsDuration)
	require.NoError(t, err)

	require.Equal(t, pomodoro.ID, pomodoroFromDB.ID)
	require.WithinDuration(t, pomodoro.StartAt, pomodoroFromDB.StartAt, time.Second)
	require.WithinDuration(t, pomodoro.FinishAt, pomodoroFromDB.FinishAt, time.Second)
	require.Equal(t, pomodoro.SecondsDuration, pomodoroFromDB.SecondsDuration)

	clearTable()
}

func TestStorage_FetchPomodoros(t *testing.T) {
	const n = 100

	for i := range n {
		pomodoro := &Pomodoro{
			ID:      i,
			StartAt: time.Now(),
		}
		err := s.CreatePomodoro(pomodoro)
		require.NoError(t, err)
	}

	pomodoros, err := s.fetchPomodoros(`SELECT * FROM pomodoros;`)
	require.NoError(t, err)

	require.Len(t, pomodoros, n)

	clearTable()
}

func TestStorage_GetPomodoros(t *testing.T) {
	const n = 100

	for i := range n {
		pomodoro := &Pomodoro{
			ID:      i,
			StartAt: time.Now(),
		}
		err := s.CreatePomodoro(pomodoro)
		require.NoError(t, err)
	}

	const expectedLen = 100
	pomodoros, err := s.GetPomodoros()
	require.NoError(t, err)
	require.Len(t, pomodoros, expectedLen)

	clearTable()
}

func TestStorage_GetTodayPomodoros(t *testing.T) {
	const n = 100

	for i := range n {
		pomodoro := &Pomodoro{
			ID:      i,
			StartAt: randomTimestamp(),
		}
		err := s.CreatePomodoro(pomodoro)
		require.NoError(t, err)
	}

	pomodoros, err := s.GetTodayPomodoros()
	require.NoError(t, err)

	for _, pomodoro := range pomodoros {
		assert.Equal(t, time.Now().Day(), pomodoro.StartAt.Day())
	}

	clearTable()
}

func TestStorage_CreateTask(t *testing.T) {
	expectedTasks := []*Task{
		{
			ID:                 100,
			Name:               "1",
			PomodorosRequired:  5,
			PomodorosCompleted: 0,
			IsActive:           false,
			IsComplete:         false,
			CreateAt:           time.Now(),
		},
		{
			ID:                 100,
			Name:               "2",
			PomodorosRequired:  5,
			PomodorosCompleted: 0,
			IsActive:           false,
			IsComplete:         false,
			CreateAt:           time.Now(),
		},
		{
			ID:                 100,
			Name:               "3",
			PomodorosRequired:  5,
			PomodorosCompleted: 0,
			IsActive:           false,
			IsComplete:         false,
			CreateAt:           time.Now(),
		},
		{
			ID:                 100,
			Name:               "3",
			PomodorosRequired:  5,
			PomodorosCompleted: 0,
			IsActive:           false,
			IsComplete:         false,
			CreateAt:           time.Now(),
		},
		{
			ID:                 100,
			Name:               "4",
			PomodorosRequired:  5,
			PomodorosCompleted: 0,
			IsActive:           false,
			IsComplete:         false,
			CreateAt:           time.Now(),
		},
		{
			ID:                 100,
			Name:               "5",
			PomodorosRequired:  5,
			PomodorosCompleted: 0,
			IsActive:           false,
			IsComplete:         false,
			CreateAt:           time.Now(),
		},
	}

	for _, task := range expectedTasks {
		err := s.CreateTask(task)
		require.NoError(t, err)

		require.NotEqualf(t, 100, task.ID, "autoincrement not working")
	}

	rows, err := s.DB.Query(`SELECT * FROM tasks;`)
	require.NoError(t, err)
	defer rows.Close()

	var taskInDB []*Task
	for rows.Next() {
		var task Task

		err = rows.Scan(&task.ID, &task.Name, &task.PomodorosRequired, &task.PomodorosCompleted,
			&task.IsComplete, &task.IsActive, &task.CreateAt)
		require.NoError(t, err)
		taskInDB = append(taskInDB, &task)
	}
	require.NoError(t, rows.Err())

	sort.Slice(taskInDB, func(i, j int) bool {
		return taskInDB[i].Name < taskInDB[j].Name
	})

	assert.Equal(t, len(expectedTasks), len(taskInDB))

	for i := range taskInDB {
		assert.Equal(t, expectedTasks[i].ID, taskInDB[i].ID)
		assert.Equal(t, expectedTasks[i].Name, taskInDB[i].Name)
		assert.Equal(t, expectedTasks[i].PomodorosRequired, taskInDB[i].PomodorosRequired)
		assert.Equal(t, expectedTasks[i].PomodorosCompleted, taskInDB[i].PomodorosCompleted)
	}

	clearTable()
}

func TestStorage_Tasks(t *testing.T) {
	expectedTasks := []*Task{
		{
			ID:                 100,
			Name:               "1",
			PomodorosRequired:  5,
			PomodorosCompleted: 0,
			IsActive:           false,
			IsComplete:         false,
			CreateAt:           time.Now(),
		},
		{
			ID:                 100,
			Name:               "2",
			PomodorosRequired:  5,
			PomodorosCompleted: 0,
			IsActive:           false,
			IsComplete:         false,
			CreateAt:           time.Now(),
		},
		{
			ID:                 100,
			Name:               "3",
			PomodorosRequired:  5,
			PomodorosCompleted: 0,
			IsActive:           false,
			IsComplete:         false,
			CreateAt:           time.Now(),
		},
		{
			ID:                 100,
			Name:               "3",
			PomodorosRequired:  5,
			PomodorosCompleted: 0,
			IsActive:           false,
			IsComplete:         false,
			CreateAt:           time.Now(),
		},
		{
			ID:                 100,
			Name:               "4",
			PomodorosRequired:  5,
			PomodorosCompleted: 0,
			IsActive:           false,
			IsComplete:         false,
			CreateAt:           time.Now(),
		},
		{
			ID:                 100,
			Name:               "5",
			PomodorosRequired:  5,
			PomodorosCompleted: 0,
			IsActive:           false,
			IsComplete:         false,
			CreateAt:           time.Now(),
		},
	}

	for _, task := range expectedTasks {
		err := s.CreateTask(task)
		require.NoError(t, err)

		require.NotEqualf(t, 100, task.ID, "autoincrement not working")
	}

	taskInDB, err := s.Tasks()
	require.NoError(t, err)

	sort.Slice(taskInDB, func(i, j int) bool {
		return taskInDB[i].Name < taskInDB[j].Name
	})

	assert.Equal(t, len(expectedTasks), len(taskInDB))

	for i := range taskInDB {
		assert.Equal(t, expectedTasks[i].ID, taskInDB[i].ID)
		assert.Equal(t, expectedTasks[i].Name, taskInDB[i].Name)
		assert.Equal(t, expectedTasks[i].PomodorosRequired, taskInDB[i].PomodorosRequired)
		assert.Equal(t, expectedTasks[i].PomodorosCompleted, taskInDB[i].PomodorosCompleted)
	}

	clearTable()
}

func randomTimestamp() time.Time {
	now := time.Now()

	minSeconds := int64(2 * 24 * 60 * 60) // 2 дня
	maxSeconds := int64(3 * 24 * 60 * 60) // 3 дня

	randomSeconds := rand.Int63n(maxSeconds-minSeconds) + minSeconds

	randomTime := now.Add(time.Duration(randomSeconds) * time.Second)

	return randomTime
}

func clearTable() {
	s.DB.Exec(`DELETE FROM pomodoros;`)
	s.DB.Exec(`DELETE FROM TASKS`)
}

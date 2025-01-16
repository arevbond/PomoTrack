//nolint:govet,exhaustruct,gochecknoglobals,errcheck // still immature code, work in progress
package main

import (
	"embed"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

const testDBName = ":memory:"

var s *Storage

//go:embed migrations/*.sql
var embedMigrations embed.FS

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
	s, err = NewStorage(testDBName)
	if err != nil {
		return err
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite"); err != nil {
		panic(err)
	}

	if err := goose.Up(s.DB, "migrations"); err != nil {
		panic(err)
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
}

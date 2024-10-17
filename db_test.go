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
	s, err = newStorage(testDBName)
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

func TestStorage_CreateTask(t *testing.T) {
	task := &Task{
		ID:       100,
		StartAt:  time.Now(),
		FinishAt: time.Now(),
	}
	err := s.CreateTask(task)
	require.NoError(t, err)

	var taskFromDB Task
	err = s.DB.QueryRow(`SELECT * FROM tasks;`).Scan(&taskFromDB.ID, &taskFromDB.StartAt,
		&taskFromDB.FinishAt, &taskFromDB.Duration)
	require.NoError(t, err)

	require.Equal(t, task.ID, taskFromDB.ID)
	require.WithinDuration(t, task.StartAt, taskFromDB.StartAt, time.Second)
	require.WithinDuration(t, task.FinishAt, taskFromDB.FinishAt, time.Second)
	require.Equal(t, task.Duration, taskFromDB.Duration)

	clearTable()
}

func TestStorage_UpdateTask(t *testing.T) {
	task := &Task{
		ID:       100,
		StartAt:  time.Now(),
		FinishAt: time.Now(),
	}
	err := s.CreateTask(task)
	require.NoError(t, err)

	task.FinishAt = time.Now().Add(60 * time.Second)
	task.Duration = 60

	err = s.UpdateTask(task)
	require.NoError(t, err)

	var taskFromDB Task
	err = s.DB.QueryRow(`SELECT * FROM tasks;`).Scan(&taskFromDB.ID, &taskFromDB.StartAt,
		&taskFromDB.FinishAt, &taskFromDB.Duration)
	require.NoError(t, err)

	require.Equal(t, task.ID, taskFromDB.ID)
	require.WithinDuration(t, task.StartAt, taskFromDB.StartAt, time.Second)
	require.WithinDuration(t, task.FinishAt, taskFromDB.FinishAt, time.Second)
	require.Equal(t, task.Duration, taskFromDB.Duration)

	clearTable()
}

func TestStorage_FetchTasks(t *testing.T) {
	const n = 100

	for i := range n {
		task := &Task{
			ID:      i,
			StartAt: time.Now(),
		}
		err := s.CreateTask(task)
		require.NoError(t, err)
	}

	tasks, err := s.fetchTasks(`SELECT * FROM tasks;`)
	require.NoError(t, err)

	require.Len(t, tasks, n)

	clearTable()
}

func TestStorage_GetTasks(t *testing.T) {
	const n = 100

	for i := range n {
		task := &Task{
			ID:      i,
			StartAt: time.Now(),
		}
		err := s.CreateTask(task)
		require.NoError(t, err)
	}

	const expectedLen = 50
	tasks, err := s.GetTasks(expectedLen)
	require.NoError(t, err)
	require.Len(t, tasks, expectedLen)

	clearTable()
}

func TestStorage_GetTodayTasks(t *testing.T) {
	const n = 100

	for i := range n {
		task := &Task{
			ID:      i,
			StartAt: randomTimestamp(),
		}
		err := s.CreateTask(task)
		require.NoError(t, err)
	}

	tasks, err := s.GetTodayTasks()
	require.NoError(t, err)

	for _, task := range tasks {
		assert.Equal(t, time.Now().Day(), task.StartAt.Day())
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
	s.DB.Exec(`DELETE FROM tasks;`)
}

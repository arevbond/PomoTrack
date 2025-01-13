package main

import (
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"

	"github.com/arevbond/PomoTrack/config"

	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

type Storage struct {
	DB *sql.DB
}

func NewStorage(filename string) (*Storage, error) {
	db, err := sql.Open("sqlite3", filepath.Join(config.GetConfigDir(), filename))
	if err != nil {
		return nil, fmt.Errorf("can't connect to db: %w", err)
	}
	return &Storage{DB: db}, nil
}

func (s *Storage) Migrate() error {
	goose.SetBaseFS(embeddedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("can't set database dialect: %w", err)
	}

	if err := goose.Up(s.DB, "migrations"); err != nil {
		return fmt.Errorf("can't make database migrations: %w", err)
	}
	return nil
}

func (s *Storage) CreateTask(task *Task) error {
	query := `INSERT INTO tasks (start_at, finish_at, duration) VALUES (?, ?, ?) RETURNING id;`

	args := []any{task.StartAt, task.FinishAt, task.SecondsDuration}

	err := s.DB.QueryRow(query, args...).Scan(&task.ID)
	if err != nil {
		return fmt.Errorf("can't create task: %w", err)
	}
	return nil
}

func (s *Storage) UpdateTask(task *Task) error {
	query := `UPDATE tasks SET finish_at = ?, duration = ? WHERE id = ?;`

	args := []any{task.FinishAt, task.SecondsDuration, task.ID}

	_, err := s.DB.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("can't update task: %w", err)
	}
	return nil
}

func (s *Storage) RemoveTask(id int) error {
	query := `DELETE FROM tasks WHERE id = ?`

	_, err := s.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("can't remove task: %w", err)
	}
	return nil
}

func (s *Storage) GetTasks() ([]*Task, error) {
	query := `SELECT id, start_at, finish_at, duration
			FROM tasks
			ORDER BY start_at DESC`
	return s.fetchTasks(query)
}

func (s *Storage) GetTodayTasks() ([]*Task, error) {
	query := `SELECT id, start_at, finish_at, duration
			FROM tasks
			WHERE date(start_at) = current_date
			ORDER BY start_at DESC`

	return s.fetchTasks(query)
}

func (s *Storage) fetchTasks(query string, args ...any) ([]*Task, error) {
	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("can't get tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*Task

	for rows.Next() {
		var task Task

		err = rows.Scan(&task.ID, &task.StartAt, &task.FinishAt, &task.SecondsDuration)
		if err != nil {
			return nil, fmt.Errorf("can't scan task: %w", err)
		}
		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return tasks, nil
}

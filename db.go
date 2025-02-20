package main

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/arevbond/PomoTrack/config"
	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID                 int       `db:"id"`
	Name               string    `db:"name"`
	PomodorosRequired  int       `db:"pomodoros_requires"`
	PomodorosCompleted int       `db:"pomodoros_completed"`
	IsComplete         bool      `db:"is_complete"`
	IsActive           bool      `db:"is_active"`
	CreateAt           time.Time `db:"created_at"`
}

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

type Storage struct {
	logger *slog.Logger
	DB     *sql.DB
}

func NewStorage(filename string, logger *slog.Logger) (*Storage, error) {
	db, err := sql.Open("sqlite3", filepath.Join(config.GetConfigDir(), filename))
	if err != nil {
		return nil, fmt.Errorf("can't connect to db: %w", err)
	}
	return &Storage{DB: db, logger: logger}, nil
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

func (s *Storage) CreatePomodoro(pomodoro *Pomodoro) error {
	query := `INSERT INTO pomodoros (start_at, finish_at, duration) VALUES (?, ?, ?) RETURNING id;`

	args := []any{pomodoro.StartAt, pomodoro.FinishAt, pomodoro.SecondsDuration}

	err := s.DB.QueryRow(query, args...).Scan(&pomodoro.ID)
	if err != nil {
		return fmt.Errorf("can't create pomodoro: %w", err)
	}
	return nil
}

func (s *Storage) UpdatePomodoro(pomodoro *Pomodoro) error {
	query := `UPDATE pomodoros SET finish_at = ?, duration = ? WHERE id = ?;`

	args := []any{pomodoro.FinishAt, pomodoro.SecondsDuration, pomodoro.ID}

	_, err := s.DB.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("can't update pomodoro: %w", err)
	}
	return nil
}

func (s *Storage) RemovePomodoro(id int) error {
	query := `DELETE FROM pomodoros WHERE id = ?`

	_, err := s.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("can't remove pomodoro: %w", err)
	}
	return nil
}

func (s *Storage) GetPomodoros() ([]*Pomodoro, error) {
	query := `SELECT id, start_at, finish_at, duration
			FROM pomodoros
			ORDER BY start_at DESC`
	return s.fetchPomodoros(query)
}

func (s *Storage) GetTodayPomodoros() ([]*Pomodoro, error) {
	query := `SELECT id, start_at, finish_at, duration
			FROM pomodoros
			WHERE date(start_at) = current_date
			ORDER BY start_at DESC`

	return s.fetchPomodoros(query)
}

func (s *Storage) fetchPomodoros(query string, args ...any) ([]*Pomodoro, error) {
	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("can't get pomodoros: %w", err)
	}
	defer rows.Close()

	var pomodoros []*Pomodoro

	for rows.Next() {
		var pomodoro Pomodoro

		err = rows.Scan(&pomodoro.ID, &pomodoro.StartAt, &pomodoro.FinishAt, &pomodoro.SecondsDuration)
		if err != nil {
			return nil, fmt.Errorf("can't scan pomodoro: %w", err)
		}
		pomodoros = append(pomodoros, &pomodoro)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return pomodoros, nil
}

func (s *Storage) Tasks() ([]*Task, error) {
	query := `SELECT id, name, pomodoros_required, pomodoros_completed,
						is_complete, is_active, created_at
			  FROM tasks
			  ORDER BY created_at;`
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("can't get tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*Task

	for rows.Next() {
		var task Task

		err = rows.Scan(&task.ID, &task.Name, &task.PomodorosRequired, &task.PomodorosCompleted,
			&task.IsComplete, &task.IsActive, &task.CreateAt)
		if err != nil {
			return nil, fmt.Errorf("can't scan task: %w", err)
		}
		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *Storage) CreateTask(task *Task) error {
	query := `INSERT INTO tasks (name, pomodoros_required, pomodoros_completed, is_complete, is_active)
				VALUES (?, ?, ?, ?, ?)
				RETURNING id`
	args := []any{task.Name, task.PomodorosRequired, task.PomodorosCompleted, task.IsComplete, task.IsActive}

	err := s.DB.QueryRow(query, args...).Scan(&task.ID)
	if err != nil {
		s.logger.Error("can't create task",
			slog.String("name", task.Name),
			slog.Int("pomodoro_required", task.PomodorosRequired),
			slog.Int("pomodoro_completed", task.PomodorosCompleted),
			slog.Bool("is_complete", task.IsComplete),
			slog.Bool("is_active", task.IsActive))

		return fmt.Errorf("can't create task: %w", err)
	}
	return nil
}

func (s *Storage) DeleteTask(id int) error {
	query := `DELETE FROM tasks WHERE id = ?`
	_, err := s.DB.Exec(query, id)
	if err != nil {
		s.logger.Error("can't delete task", slog.Int("id", id))
		return fmt.Errorf("can't delete task: %w", err)
	}
	return nil
}

func (s *Storage) UpdateTask(task *Task) error {
	query := `UPDATE tasks 
				SET name = ?, pomodoros_required = ?, pomodoros_completed = ?, is_complete = ?, is_active = ?
				WHERE id = ?;`
	args := []any{task.Name, task.PomodorosRequired, task.PomodorosCompleted, task.IsComplete, task.IsActive, task.ID}
	_, err := s.DB.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("can't upate task: %w", err)
	}
	return nil
}

func (s *Storage) ActiveTask() (*Task, error) {
	query := `SELECT id, name, pomodoros_required, pomodoros_completed, is_complete, is_active, created_at
				FROM tasks
			  WHERE is_active = true;`
	var task Task
	err := s.DB.QueryRow(query).Scan(&task.ID, &task.Name, &task.PomodorosRequired, &task.PomodorosCompleted,
		&task.IsComplete, &task.IsActive, &task.CreateAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn("query active task, when active task not exist")
			return nil, fmt.Errorf("can't find active task: %w", err)
		}
		return nil, fmt.Errorf("can't find acitve task: %w", err)
	}
	return &task, nil
}

func (s *Storage) IncPomodoroActiveTask() error {
	query := `UPDATE tasks
    			SET
					pomodoros_completed = pomodoros_completed + 1,
					is_complete = (pomodoros_completed + 1) >= pomodoros_required
				WHERE is_active = true;`
	_, err := s.DB.Exec(query)
	if err != nil {
		return fmt.Errorf("can't inc pomodoros in active task: %w", err)
	}
	return nil
}

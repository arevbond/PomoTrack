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

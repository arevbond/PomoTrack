-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
CREATE TABLE IF NOT EXISTS tasks (
    id integer primary key,
    name text NOT NULL DEFAULT '',
    pomodoros_required integer NOT NULL DEFAULT 0,
    pomodoros_completed integer NOT NULL DEFAULT 0,
    is_complete boolean NOT NULL DEFAULT false,
    is_active boolean NOT NULL default false,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
DROP TABLE IF EXISTS tasks;

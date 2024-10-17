-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY,
    start_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    finish_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    duration INTEGER NOT NULL DEFAULT 0
);


-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
DROP TABLE IF EXISTS tasks;
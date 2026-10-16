-- +goose Up
-- +goose StatementBegin

ALTER TABLE users ADD COLUMN password_reset_token TEXT;
ALTER TABLE users ADD COLUMN password_reset_expires TEXT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- SQLite does not support DROP COLUMN easily; leave columns as-is for down migration.

-- +goose StatementEnd

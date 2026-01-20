-- +goose Up
-- +goose StatementBegin
ALTER TABLE sessions ADD COLUMN user_agent TEXT;
ALTER TABLE sessions ADD COLUMN ip_address TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- SQLite doesn't support dropping columns easily, but for a session table we can just recreate it or ignore.
-- +goose StatementEnd

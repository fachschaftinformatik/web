-- +goose Up
-- +goose StatementBegin
ALTER TABLE events ADD COLUMN cover_path TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events DROP COLUMN cover_path;
-- +goose StatementEnd

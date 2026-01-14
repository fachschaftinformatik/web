-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN private INTEGER NOT NULL DEFAULT 0 CHECK (private IN (0,1));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN private;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
ALTER TABLE activities ADD COLUMN target_name TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE activities DROP COLUMN target_name;
-- +goose StatementEnd

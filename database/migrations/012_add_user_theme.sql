-- +goose Up
ALTER TABLE users ADD COLUMN theme TEXT NOT NULL DEFAULT 'system';

-- +goose Down
-- SQLite does not support dropping columns easily, but we can't really undo this without recreating the table.
-- For simplicity in this dev environment, we leave it.

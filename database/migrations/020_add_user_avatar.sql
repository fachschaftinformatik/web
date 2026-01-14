-- +goose Up
-- Add avatar_url column to users table
ALTER TABLE users ADD COLUMN avatar_url TEXT;

-- +goose Down
-- In SQLite, we can't easily drop columns without recreating the table.
-- Since this is a new column, we can leave it or recreate the table if rollback is strictly needed.
-- For now, we leave it empty to avoid accidental data loss.

-- +goose Up
-- +goose StatementBegin
CREATE TABLE settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
) STRICT;

INSERT INTO settings (key, value) VALUES ('office_occupied', 'false');

CREATE TRIGGER trg_settings_updated_at AFTER UPDATE ON settings BEGIN
    UPDATE settings SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE key = OLD.key;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS settings;
-- +goose StatementEnd

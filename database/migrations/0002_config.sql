-- +goose Up
-- +goose StatementBegin
CREATE TABLE global_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    office_occupied INTEGER NOT NULL DEFAULT 0,
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
) STRICT;

INSERT INTO global_config (id, office_occupied) VALUES (1, 0);

CREATE TRIGGER trg_global_config_updated_at AFTER UPDATE ON global_config BEGIN
    UPDATE global_config SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = 1;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS global_config;
-- +goose StatementEnd

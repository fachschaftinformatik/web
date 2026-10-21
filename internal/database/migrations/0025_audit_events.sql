-- +goose Up
CREATE TABLE audit_events (
    id INTEGER PRIMARY KEY,
    occurred_at TEXT NOT NULL,
    event TEXT NOT NULL,
    outcome TEXT NOT NULL,
    actor_user_id INTEGER,
    user_id INTEGER,
    request_id TEXT NOT NULL DEFAULT '',
    method TEXT NOT NULL DEFAULT '',
    route TEXT NOT NULL DEFAULT '',
    details TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(details))
) STRICT;

-- Account identifiers intentionally have no foreign keys: deleting an account
-- must not delete its history or prevent retaining that history.
CREATE INDEX idx_audit_events_user ON audit_events(user_id, occurred_at);
CREATE INDEX idx_audit_events_actor ON audit_events(actor_user_id, occurred_at);
CREATE INDEX idx_audit_events_event ON audit_events(event, occurred_at);
CREATE INDEX idx_audit_events_request ON audit_events(request_id) WHERE request_id != '';

-- +goose StatementBegin
CREATE TRIGGER trg_utc_audit_events_insert BEFORE INSERT ON audit_events
WHEN NOT (typeof(NEW.occurred_at)='text' AND length(NEW.occurred_at)=20
    AND strftime('%Y-%m-%dT%H:%M:%SZ',NEW.occurred_at,'+0 seconds') IS NEW.occurred_at)
BEGIN
    SELECT RAISE(ABORT,'timestamps must be UTC RFC 3339 ending in Z; calendar dates must be YYYY-MM-DD');
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trg_utc_audit_events_update BEFORE UPDATE ON audit_events
WHEN NOT (typeof(NEW.occurred_at)='text' AND length(NEW.occurred_at)=20
    AND strftime('%Y-%m-%dT%H:%M:%SZ',NEW.occurred_at,'+0 seconds') IS NEW.occurred_at)
BEGIN
    SELECT RAISE(ABORT,'timestamps must be UTC RFC 3339 ending in Z; calendar dates must be YYYY-MM-DD');
END;
-- +goose StatementEnd

-- +goose Down
DROP TABLE audit_events;

-- +goose Up
CREATE TABLE archive_requests (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'pending', 'approved', 'rejected')),
    revision_of INTEGER REFERENCES archive(id),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    submitted_at TEXT,
    reviewed_at TEXT,
    reviewed_by INTEGER REFERENCES users(id)
) STRICT;
CREATE INDEX archive_requests_owner ON archive_requests(user_id, created_at);
CREATE INDEX archive_requests_status ON archive_requests(status, submitted_at);
ALTER TABLE archive ADD COLUMN request_id TEXT REFERENCES archive_requests(id);
CREATE INDEX archive_request ON archive(request_id);

-- +goose Down
DROP INDEX archive_request;
ALTER TABLE archive DROP COLUMN request_id;
DROP TABLE archive_requests;

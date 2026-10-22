-- +goose Up
CREATE TABLE submission_decision_mail_outbox (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    submission_id TEXT NOT NULL UNIQUE REFERENCES archive_requests(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('approved', 'rejected'))
) STRICT;

-- +goose Down
DROP TABLE submission_decision_mail_outbox;

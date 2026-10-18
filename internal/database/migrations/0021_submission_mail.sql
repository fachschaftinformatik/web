-- +goose Up
CREATE TABLE submission_mail_subscriptions (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE
) STRICT;

CREATE TABLE submission_mail_outbox (
    submission_id TEXT NOT NULL REFERENCES archive_requests(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (submission_id, user_id)
) STRICT;

-- +goose Down
DROP TABLE submission_mail_outbox;
DROP TABLE submission_mail_subscriptions;

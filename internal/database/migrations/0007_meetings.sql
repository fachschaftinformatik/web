-- +goose Up
CREATE TABLE meeting_publications (
    id TEXT PRIMARY KEY,
    owner TEXT NOT NULL CHECK(length(owner) > 0),
    date TEXT NOT NULL,
    number INTEGER NOT NULL,
    kind TEXT NOT NULL,
    filename TEXT NOT NULL,
    blob_key TEXT NOT NULL UNIQUE,
    bytes INTEGER NOT NULL CHECK(bytes >= 0 AND bytes <= 8388608),
    sha256 TEXT NOT NULL,
    published INTEGER NOT NULL DEFAULT 0 CHECK(published IN (0, 1))
) STRICT;
CREATE TABLE published_meetings (
    year INTEGER NOT NULL,
    number INTEGER NOT NULL,
    date TEXT NOT NULL UNIQUE,
    PRIMARY KEY(year, number)
) STRICT;

CREATE TABLE meeting_blob_cleanup (
    blob_key TEXT PRIMARY KEY,
    attempts INTEGER NOT NULL DEFAULT 0
) STRICT;

-- +goose Down
DROP TABLE meeting_blob_cleanup;
DROP TABLE published_meetings;
DROP TABLE meeting_publications;

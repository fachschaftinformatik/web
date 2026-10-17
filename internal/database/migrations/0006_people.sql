-- +goose Up
CREATE TABLE people_state (id INTEGER PRIMARY KEY CHECK(id=1)) STRICT;
CREATE TABLE people (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL DEFAULT '',
    user_id INTEGER UNIQUE REFERENCES users(id),
    position INTEGER NOT NULL
) STRICT;
CREATE TABLE responsibilities (
    id TEXT PRIMARY KEY,
    group_name TEXT NOT NULL CHECK(group_name IN ('board','departments','project')),
    title TEXT NOT NULL,
    email TEXT NOT NULL DEFAULT '',
    position INTEGER NOT NULL
) STRICT;
CREATE TABLE responsibility_members (
    responsibility_id TEXT NOT NULL REFERENCES responsibilities(id),
    person_id TEXT NOT NULL REFERENCES people(id),
    position INTEGER NOT NULL,
    PRIMARY KEY(responsibility_id,person_id)
) STRICT;

-- +goose Down
DROP TABLE responsibility_members;
DROP TABLE responsibilities;
DROP TABLE people;
DROP TABLE people_state;

# db

SQLite schema and queries for the server.

We use [STRICT][1] tables introduced in SQLite version 3.37.0 to enable strict typing.

Enable these SQLite pragmas in the DSN:
```sql
-- Use the "Write-Ahead Log" to support concurrent reading/writing
-- See: sqlite.org/wal.html
PRAGMA journal_mode = WAL;
-- By default, foreign keys are not enabled in SQLite
-- See: sqlite.org/foreignkeys.html
PRAGMA foreign_keys = ON;
-- Disable recursive triggers
PRAGMA recursive_triggers = OFF;
```
For example:
```
file:dev.db?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=recursive_triggers(OFF)
```

To compile the SQL:
```sh
go generate
```
The generated files are then written to `internal/db`.

To review changes:
```
git diff internal/db
```

Relevant documentation:
- [SQLite STRICT Tables][1]
- [SQLite Write-Ahead Logging][2]
- [SQLite Foreign Key Support][3]

[1]: https://sqlite.org/stricttables.html
[2]: https://sqlite.org/wal.html
[3]: https://sqlite.org/foreignkeys.html

# One-time import of existing website content

This directory preserves the website content formerly compiled into Go and HTML.
It is an explicit data migration for the running installation, not application
startup logic or a development fixture generator. It does not insert members,
accounts or responsibility assignments.

Run against a writable copy first:

```sh
go run . -database .dev-data/local/sqlite.db -migrate-only
go run ./tools/import-site-data .dev-data/local/sqlite.db .dev-data/local/blobs
```

The import retains the existing organization details and 48 original PDFs,
including all source mappings, dates, notes, byte counts and SHA-256 digests.
Historical document URLs remain `/assets/content/FILE.pdf`. Bytes are now read
from object storage using `meetings/historical/SHA256` keys.

`content.sql` is the manually applied data import. Review its organization
settings before applying it to the running installation. Chair and publicity
responsibility references use the existing IDs `1` and `6`; verify these against
the target database. Contact pages use the addresses stored on those existing
responsibilities rather than a second set of template literals. No domain aliases
are inferred or rewritten.

The Go command verifies bytes, copies objects without replacing differing
content, and applies the SQL transaction. Repeating it preserves existing settings
and document records. If verification fails, database changes are rolled back.

For S3, use the local blob directory as staging, upload `meetings/historical/`
under that exact prefix in the configured bucket, and verify the objects before
starting the application with the migrated database. Apply the SQL explicitly
only after the corresponding objects are available.

Set `ACCOUNT_DOMAINS`, `REGISTRATION_DOMAINS` and the existing deployment settings
before starting the new executable. Organization settings are loaded at startup;
administrators then maintain them under **Verwaltung → Stammdaten**, with immediate
effect on the website and email. Direct SQL edits take effect after restart. Development thereafter uses production
database/object copies; no setup or content import runs automatically.

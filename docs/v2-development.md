# V2 development

`v2/go-templates` is the development branch for the Go-template replacement.
The reviewed public-site and administration prototype is now the sole application
in the repository. The V1 source remains in Git history at `86a6968`.

## Run and verify

```sh
go run .
go test -race ./...
go vet ./...
go generate ./database
go build -o /path/to/output/web .
```

Use `go run . -data .dev-data/production-2026-10-14` for read-only account display
data from the supplied snapshot. The site is at <http://127.0.0.1:8090/>; local
administration uses `admin` / `admin`. See [implemented journeys](web-demo.md)
and [pattern research](govuk-patterns.md) for the reviewed interface and its
current fixture/in-memory behaviour.

## Repository cutover completed

- `main.go` starts the HTML application directly. There is one executable and no
  `cmd/` directory. It supports graceful shutdown, version output and a health
  probe used by the container.
- Domain packages live directly under `internal/`; `internal/handler/http`
  contains templates, embedded assets and HTML transport. Albums and photographs
  share `internal/images` with the existing image-processing functions.
- Removed `website/`, generated TypeScript client, V1 REST transport, Swagger
  output/tooling, Node runtime, s6 scripts and obsolete proxy configuration.
- Docker builds one static Go executable and runs it as UID 1001 on port 8080.
  Compose exposes localhost:8090 and keeps the existing `/data` volume. Garage is
  an optional development profile, not a startup requirement.
- `ci:app` covers branch/PR tests, formatting, vet, race detection, SQLC generation,
  security analysis, release-tool tests and the application image build. The
  release gate remains `release:check`. Smoke tests and their Compose fixture
  were removed at the user's request.
- Renovate no longer regenerates a website API client. Go generation remains;
  Node/Bun are used only for release tooling. Releases keep title-only commits
  and empty descriptions.

This is a repository/build cutover on a development branch. Persistent account,
archive and forum behaviour is still being ported. Removing the old HTTP boundary
does not mean that all its business operations are connected to the new handlers.

## Reuse and remaining domain migration

The existing SQLite migration history, SQLC queries, ID generation, object-store
interface/S3 adapter, image/avatar processing and SMTP transport are retained.
The reference implementation of the old operations is at
`86a6968:internal/api/handler/`. Port operations into the owning domain, rather
than reintroducing a REST layer or making handlers invoke other handlers.

| Domain | Remaining operations |
| --- | --- |
| Accounts | Registration, existing-password verification, sessions and role/active/verified policy, account edits, expiring single-use reset/verification links and SMTP delivery. |
| Archive | Programmes/modules, latest revisions and version history, authenticated downloads, submission/update/delete operations and blob metadata. |
| Forum | Persistent posts/comments, ownership and moderation rules, votes, deletion and search. Keep the reviewed simplified editor and newest-first comments. |
| Images | Persistent event/album/media lookup, upload validation, previews, originals and deletion. Retain IDs and object keys. |
| Team | Persistent member profiles, responsibilities and assignments, distinct from user accounts. |
| Meetings | Persistent meeting/publication metadata, draft ownership, retained PDF uploads and explicit publication. |

The current team and meeting stores own validation and mutations, but their
state is process-local. Forum editing remains a preview. Normal registration,
login and password-reset submissions explicitly report their demo outcome.

## Data migration findings

The supplied read-only snapshot contains 362 accounts, 244 modules, 796 archive
rows, 3 events, 228 media records, 5 posts and 8 comments, before feature-specific
soft-deletion/version filters. Archive rows include revision/group information;
796 rows must not be treated as 796 independently listed latest uploads.

There are no tables for public members, departments, project roles or meetings.
These need migrations populated from the curated public-content sources.
Retain account IDs and role values (`admin`, `editor`, `user`). FSV membership
and administration permissions need an explicit mapping: an editor is not
automatically an FSV member.

The snapshot separates `active` from `verified`; the UI displays both. Lowercase
email display does not rewrite identities in the source snapshot.

V1 archive downloads require authentication. Apply that policy in domain services
before connecting real files. Public meeting documents have a different policy;
link visibility is not access control.

## Next implementation sequence

1. Connect read-only models for accounts, programmes/modules, latest archive
   revisions, events/media and discussions; retain IDs and blob references.
2. Add persistent member profiles, responsibility assignments, meetings and
   publication metadata. Move PDFs from process memory to blob storage.
3. Connect existing password hashes, account/session policy and SMTP workflows.
   Keep exact email-domain rules and the reviewed reset confirmation.
4. Complete forum/archive writes and compatibility redirects. Verify copied data
   through local adapters before a production deployment.

## Local files

Production copies, object metadata and credentials stay under ignored paths.
Only curated public content and fictional account fixtures are committed.
The pre-existing `website/next-env.d.ts` and `website/tsconfig.tsbuildinfo` contents
were copied byte-for-byte to `.dev-data/v1-generated/` before removing V1.
Other untracked website build files were moved to `.dev-data/v1-website-local/`.

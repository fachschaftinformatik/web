# V2 development

`v2/go-templates` is the development branch for the Go-template replacement.
The reviewed public-site and administration prototype is the starting point for
the new application, rather than another frontend to maintain alongside Next.js.
The branch starts from `86a6968`; Git history retains the V1 implementation.

## Run and verify

```sh
go run ./cmd/web-demo
go test -race ./internal/webdemo/... ./cmd/web-demo
go vet ./internal/webdemo/... ./cmd/web-demo
```

To load the supplied local account snapshot:

```sh
go run ./cmd/web-demo -data .dev-data/production-2026-10-14
```

The site is at <http://127.0.0.1:8090/>. Sign in with `admin` / `admin` for the
administration prototype. See [the current implementation](web-demo.md) and
[pattern research](govuk-patterns.md) for details, limitations and browser findings.
The `web-demo` names describe its current fixture/in-memory runtime; promoting
those names belongs with the application/infrastructure cutover below.

The `ci:v2` workflow verifies V2-branch pushes and relevant pull requests with
race-enabled tests, vet and a Go executable build. It needs neither a running
Next.js application nor a production-copy database.

## Repository investigation, 17 October 2026

| Area | Current implementation | V2 direction |
| --- | --- | --- |
| Browser application | `website/`, Next.js, generated TypeScript API client, NextAuth | Go templates and embedded static assets; HTMX only for interactions that benefit from it. |
| Entrypoint | Root `main.go` starts REST API; `cmd/web-demo` starts HTML prototype | One `cmd/web` executable wiring domain services, persistence and HTML handlers. |
| Domain operations | Existing API handlers combine transport and business logic; prototype has account/archive/forum/team/meeting packages | Move operations into domain-first packages and call them directly from thin HTTP handlers. |
| Database | SQLite, Goose migrations, SQLC queries | Retain schema/history and stable IDs; add adapters to domain services and migrations for new content. |
| Blob storage | `internal/storage`, S3/Garage, metadata in SQLite | Retain an object-store adapter and use a local bucket adapter for development. |
| Mail | `internal/email`, SMTP templates and transport | Reuse transport through a narrow interface; update reset/verification URLs for HTML routes. |
| Image processing | `internal/images`, `internal/avatars` | Reuse thumbnail/avatar operations where the new routes need them. |
| Runtime image | `Dockerfile` builds Go plus Next.js; Node and s6 run two services | One Go process and embedded assets, no Node runtime or process supervisor. |
| Local runtime | `compose.yaml` starts the combined app and Garage | Go app plus optional required infrastructure; keep SQLite/blob volumes compatible. |
| CI/release | `.github/workflows/ci.yaml`, `.github/ci/compose.yaml`, `.github/release` | Replace V1 checks/smoke tests with the Go application and V2 HTTP flows; preserve release/check naming deliberately. |
| Dependency maintenance | Renovate regenerates the website client after Go/Bun changes | Remove website generation hooks and obsolete managers; retain tools actually used for release tooling and vendored assets. |

`internal/mvcdemo` contains empty controller/model/view directories and provides
no implementation to migrate. `internal/webdemo` contains the actual new work.

### Existing infrastructure is coupled to Next.js

Removing `website/` also requires these coordinated changes:

- Replace both Node/Bun builder stages, Next standalone copy steps and runtime
  supervision in `Dockerfile` and `docker/`.
- Change app ports, environment and health checks in `compose.yaml`; the V1
  application currently exposes the frontend on port 80 and API on port 8080.
- Update `bake.hcl`, its version injection and image smoke tests together.
- Replace website lint, frontend tests, client-generation checks, JavaScript
  security analysis and the Next routing/image-processing assertions in CI.
- Remove root Swagger generation, `docs/docs.go` and generated OpenAPI files once
  the old REST boundary is removed. The root entrypoint imports these today.
- Remove `website`-specific Renovate post-upgrade commands. `.node-version` may
  still be needed by release tooling even after the application is Go-only.
- Update README, contribution instructions and release documentation for one
  application and the new startup command.

## Data migration findings

The supplied read-only snapshot contains 362 accounts, 244 modules, 796 archive
rows, 3 events, 228 media records, 5 posts and 8 comments, before any feature's
soft-deletion or version filters. Archive rows include revision/group information;
796 rows must not be treated as 796 independently listed latest uploads.

There are no tables for public members, departments, project roles or meetings.
These need new migrations, populated from the curated public-content sources in
the prototype. Existing account IDs and role values (`admin`, `editor`, `user`)
should be retained. FSV membership/administration permissions need an explicit
mapping: an `editor` is not automatically an FSV member.

The snapshot already separates `active` from `verified`; V2 displays both without
conflating activation and email confirmation. Lowercase email display does not
rewrite account identities in the source snapshot.

Archive downloads in the existing API require authentication. Before connecting
real archive files, the HTML routes need the corresponding account permission
checks. Public meeting documents have a different access policy. Keep these
policies in their domain services rather than relying on link visibility.

## Implementation sequence

1. **Promote the application structure and build together.** Move `cmd/web-demo`
   to `cmd/web`; promote domain packages out of the prototype namespace. Replace
   root startup/container/Compose/CI references in one coherent change. Delete
   `website/`, its API client and V1 runtime/supervision code from this branch;
   Git history supplies the reference implementation.
2. **Connect read models.** Use explicit read-only adapters for accounts,
   programmes/modules, latest archive revisions, events/media and discussions.
   Retain original IDs and blob references; fixtures remain available to tests.
3. **Persist new content.** Add migrations and repository operations for member
   profiles, responsibility assignments, meetings and publication metadata. Keep
   account/profile/assignment identity separate. Put uploaded PDFs in blob
   storage rather than the process-local draft map.
4. **Connect authentication and mail.** Verify existing password hashes, apply
   active/verified/role policy and replace demo-only sessions with the account
   session service. Connect the reviewed reset dialog to expiring, single-use
   reset links and the SMTP sender. Registration and verification use the agreed
   exact-domain policy.
5. **Complete write journeys and cutover.** Replace forum preview writes and
   archive submission links with domain operations, cover redirects from existing
   public URLs, test the copied database/bucket through the local adapters and
   update release smoke tests to the complete Go application.

This sequence makes removal of V1 a repository/build change rather than deleting
files while leaving generators, containers or deployment checks pointing at them.
V2 branch development can proceed independently of production deployment.

## Local files

Production copies, object metadata and credentials belong under ignored local
paths and are never application fixtures. The V2 baseline commits contain only
curated public content and fictional account fixtures. Existing local generated
Next.js files are outside those commits; their contents are preserved while
preparing the repository cutover.

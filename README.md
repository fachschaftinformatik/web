# web

[![Release](https://img.shields.io/github/v/release/fachschaftinformatik/web?include_prereleases&label=Release)](https://github.com/fachschaftinformatik/web/releases/latest)
[![Go](https://img.shields.io/github/go-mod/go-version/fachschaftinformatik/web/main?label=Go)](https://github.com/fachschaftinformatik/web/blob/main/go.mod)
[![CI](https://img.shields.io/github/actions/workflow/status/fachschaftinformatik/web/ci.yaml?label=CI)](https://github.com/fachschaftinformatik/web/actions/workflows/ci.yaml)

Website of the student representative body for computer science at Westphalian University.

## Getting started

The application uses Go templates, embedded browser assets and selective HTMX.
Go 1.27.2 is required. Development uses a copy of the production database and
object storage. The executable never inserts member rosters or site content.
Keep the original backup immutable and run against a writable copy.

```sh
export ACCOUNT_DOMAINS=studmail.w-hs.de,fsv-wh.de
export REGISTRATION_DOMAINS=studmail.w-hs.de
go run . -database .dev-data/local/sqlite.db -blobs .dev-data/local/blobs
```

Open <http://127.0.0.1:8090/>. Without flags, SQLite data is stored in
`data/sqlite.db`, with objects in `data/blobs` unless S3 is configured. All environments use
the same database-backed application. Keep local data and credentials under
ignored paths.

### Site data and configuration

`DOMAIN` is the public HTTP(S) origin. `ACCOUNT_DOMAINS` is a required,
comma-separated list of exact email domains allowed for login, recovery and
invitations. `REGISTRATION_DOMAINS` is the subset allowed for self-registration;
an empty list disables self-registration. Form hints use the same policy.

Organization identity, contacts, address, office, social links, legal metadata
and resource links live in the `organization` singleton's JSON `data` column.
Its shape is defined by `internal/model/organization.Details`. Administrators edit
these values under **Verwaltung → Stammdaten**, grouped into identity, contact,
legal information and links. Changes apply immediately to both website and email.
Direct SQL changes require an application restart to refresh the shared snapshot.
Chair and publicity contacts reference responsibility IDs; their addresses and
member assignments remain editable through the existing administration pages.

Timestamps are stored in UTC as RFC 3339 values ending in `Z`. Calendar-only
values (exam dates, meeting dates, all-day events and legal revision dates) use
`YYYY-MM-DD` without a timezone. The view formats dates centrally for `de-DE`
and converts timestamps to `Europe/Berlin` for display and event editing; locale
and timezone are separate concerns. Adding another display format belongs in
`internal/view/dates.go`, not in database records.
Migration 16 checks existing timestamps and installs database triggers on insert
and update. These reject non-canonical non-empty values, including timestamps
without `Z`, numeric offsets and invalid calendar/time values, even through direct
SQL connections. Timestamps have second precision; existing optional NULL/empty
values remain supported. Event dates additionally allow ISO calendar-only dates.

Legacy content imports and timestamp conversion are managed outside this
repository. Prepare the production database and objects before starting the new
application; the application does not import or repair legacy content. Development
uses copies of that prepared database and object storage. Schema migrations can be run with
`go run . -database PATH -migrate-only`; they insert no organization content.

### Container

```sh
docker compose up --build
```

The same site is available at <http://127.0.0.1:8090/>. The image runs one static
Go executable as UID 1001 on port 8080, with embedded templates and assets.
`/healthz` is the container health endpoint. SIGTERM drains HTTP requests before
exit. Run `go run . -help` for executable flags.

Compose persists the database and local objects in its `/data` volume.
Optional Garage object storage is available
with `docker compose --profile storage up -d garage`; its configuration and
default credentials are for local development. SMTP and S3 settings are configured
through the environment variables listed in `compose.yaml`.

## Architecture

```text
main.go                 Executable flags and HTTP lifecycle
application.go          Dependency wiring and background maintenance
internal/
  controller/           Routing, HTTP handlers, sessions and request validation
  model/
    archive/            Exam files, submissions, assignments and review
    forum/              Discussions, announcements, comments and voting
    gallery/            Albums, photographs and covers
    meetings/           Meeting documents and publication
    events/             Calendar events and imports
    users/              Accounts, authentication and access policies
    people/             Members, groups and responsibilities
    organization/       Database-backed site identity, contacts and resource links
  view/                 Presentation types, templates and embedded assets
  content/              Shared text sanitization and content validation
  database/             SQLite connections, migrations and generated queries
  storage/              Binary objects and image processing; disk/S3 adapters
  email/                SMTP transport and email templates
```

Controllers call model operations and pass presentation data to views. Models
own assignment resolution, publication rules and event classification. Typed view
adapters own website URLs, display labels, summaries and image-preview paths.
Models do not depend on controllers or views, and views do not access storage or
the database. The V1 REST API, Next.js website,
OpenAPI client generation and multi-process runtime are retired;
their implementation remains in Git history at `86a6968`.

## Checks and documentation

```sh
go test -race ./...
go vet ./...
go generate ./internal/database
go build -o /path/to/output/web .
```

See [contribution instructions](.github/CONTRIBUTING.md)
and [release operations](.github/release/README.md).

If you believe you have found a security issue, contact security@fsv-wh.de.

## Legal

SPDX-License-Identifier: MPL-2.0

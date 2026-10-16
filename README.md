# web

[![Release](https://img.shields.io/github/v/release/fachschaftinformatik/web?include_prereleases&label=Release)](https://github.com/fachschaftinformatik/web/releases/latest)
[![Go](https://img.shields.io/github/go-mod/go-version/fachschaftinformatik/web/main?label=Go)](https://github.com/fachschaftinformatik/web/blob/main/go.mod)
[![CI](https://img.shields.io/github/actions/workflow/status/fachschaftinformatik/web/ci.yaml?label=CI)](https://github.com/fachschaftinformatik/web/actions/workflows/ci.yaml)

Website of the student representative body for computer science at Westphalian University.

## Getting started

The `v2/go-templates` branch develops the server-rendered replacement with Go
templates, embedded browser assets and selective HTMX. Go 1.27.2 is required.

```sh
go run .
```

Open <http://127.0.0.1:8090/>. Sign in with **admin / admin** to try the local
administration workflows. Edits and uploaded documents currently live in memory
and reset on restart. Normal account authentication and mail are not connected.

To display accounts from a local read-only SQLite copy:

```sh
go run . -data .dev-data/production-2026-10-14
```

Only user display fields are imported. Passwords, tokens and production sessions
are not loaded. Keep local data and credentials under ignored paths.

### Container

```sh
docker compose up --build
```

The same site is available at <http://127.0.0.1:8090/>. The image runs one static
Go executable as UID 1001 on port 8080, with embedded templates and assets.
`/healthz` is the container health endpoint. SIGTERM drains HTTP requests before
exit. `-addr`, `-data`, `-version` and `-check-health` are executable flags.

Compose retains the `/data` volume for the persistence work. The current
application does not write to it. Optional Garage object storage is available
with `docker compose --profile storage up -d garage`; its configuration and
default credentials are for local development. Existing SQLite migrations, SQLC
queries, SMTP transport and S3/image adapters remain available for domain wiring.

## Architecture

```text
main.go                 Executable and HTTP lifecycle
internal/
  account/              Account policies and directory
  archive/              Modules, uploads and filtering
  forum/                Discussions, rich text and drafts
  images/               Albums, photographs and image processing
  home/                 Landing-page content
  meetings/             Documents and publication workflow
  team/                 Member profiles and responsibilities
  handler/http/         HTML handlers, templates and embedded assets
  database/             Generated SQLC queries
  storage/              Blob-storage adapters
  email/                SMTP transport and email templates
database/               SQLite migrations, SQL queries and generation
```

HTML handlers call domain operations directly. The V1 REST API, Next.js website,
OpenAPI client generation and multi-process runtime are retired on this branch;
their implementation remains in Git history at `86a6968`.

## Checks and documentation

```sh
go test -race ./...
go vet ./...
go generate ./database
go build -o /path/to/output/web .
```

See [V2 migration status](docs/v2-development.md),
[implemented journeys and checks](docs/web-demo.md),
[design research](docs/govuk-patterns.md), [contribution instructions](.github/CONTRIBUTING.md)
and [release operations](.github/release/README.md).

If you believe you have found a security issue, contact security@fsv-wh.de.

## Legal

SPDX-License-Identifier: MPL-2.0

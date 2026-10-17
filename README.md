# web

[![Release](https://img.shields.io/github/v/release/fachschaftinformatik/web?include_prereleases&label=Release)](https://github.com/fachschaftinformatik/web/releases/latest)
[![Go](https://img.shields.io/github/go-mod/go-version/fachschaftinformatik/web/main?label=Go)](https://github.com/fachschaftinformatik/web/blob/main/go.mod)
[![CI](https://img.shields.io/github/actions/workflow/status/fachschaftinformatik/web/ci.yaml?label=CI)](https://github.com/fachschaftinformatik/web/actions/workflows/ci.yaml)

Website of the student representative body for computer science at Westphalian University.

## Getting started

The application uses Go templates, embedded browser assets and selective HTMX.
Go 1.27.2 is required.

```sh
go run .
```

Open <http://127.0.0.1:8090/>. SQLite data is stored in `data/sqlite.db`, with
uploaded objects in `data/blobs` unless S3 is configured. For an in-memory demo
with **admin / admin** credentials, run `go run . -demo`.

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
exit. Run `go run . -help` for executable flags.

Compose persists the database and local objects in its `/data` volume.
Optional Garage object storage is available
with `docker compose --profile storage up -d garage`; its configuration and
default credentials are for local development. SMTP and S3 settings are configured
through the environment variables listed in `compose.yaml`.

## Architecture

```text
main.go                 Executable and HTTP lifecycle
internal/
  users/                Login identities, access policies and user management
  people/               Public profiles, groups and responsibilities
  content/              Documents and publication
    files/              File records, grouping and filtering
    posts/              Posts, comments, drafts and rich text
    media/              Albums, images and processing
  data/                 Database access, migrations and queries
    blob/               Binary-object storage
      s3/               S3-compatible adapter
  web/                  HTML handlers, templates and embedded assets
  email/                SMTP transport and email templates
```

HTML handlers call domain operations directly. The V1 REST API, Next.js website,
OpenAPI client generation and multi-process runtime are retired;
their implementation remains in Git history at `86a6968`.

## Checks and documentation

```sh
go test -race ./...
go vet ./...
go generate ./internal/data
go build -o /path/to/output/web .
```

See [contribution instructions](.github/CONTRIBUTING.md)
and [release operations](.github/release/README.md).

If you believe you have found a security issue, contact security@fsv-wh.de.

## Legal

SPDX-License-Identifier: MPL-2.0

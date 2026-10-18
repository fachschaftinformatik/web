# web

[![Release](https://img.shields.io/github/v/release/fachschaftinformatik/web?include_prereleases&label=Release)](https://github.com/fachschaftinformatik/web/releases/latest)
[![Go](https://img.shields.io/github/go-mod/go-version/fachschaftinformatik/web/main?label=Go)](https://github.com/fachschaftinformatik/web/blob/main/go.mod)
[![CI](https://img.shields.io/github/actions/workflow/status/fachschaftinformatik/web/ci.yaml?label=CI)](https://github.com/fachschaftinformatik/web/actions/workflows/ci.yaml)

Server-rendered website of the student representative body for computer science at Westphalian University, with separate model, controller (`site`), and view layers.

```text
       site
      /    \
     ▼      ▼
   model   view
```

## Getting started

With a writable development copy of the database and blobs in `.dev-data/local/`, start the site at <http://localhost:8090>:

```sh
ACCOUNT_DOMAINS=studmail.w-hs.de,fsv-wh.de REGISTRATION_DOMAINS=studmail.w-hs.de nix shell nixpkgs#go -c go run . -database .dev-data/local/sqlite.db -blobs .dev-data/local/blobs
```

[Contributing](.github/CONTRIBUTING.md) · [Releases](https://github.com/fachschaftinformatik/web/releases)

Security issues: security@fsv-wh.de.

SPDX-License-Identifier: MPL-2.0

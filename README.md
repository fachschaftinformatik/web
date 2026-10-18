# web

[![Release](https://img.shields.io/github/v/release/fachschaftinformatik/web?include_prereleases&label=Release)](https://github.com/fachschaftinformatik/web/releases/latest)
[![Go](https://img.shields.io/github/go-mod/go-version/fachschaftinformatik/web/main?label=Go)](https://github.com/fachschaftinformatik/web/blob/main/go.mod)
[![CI](https://img.shields.io/github/actions/workflow/status/fachschaftinformatik/web/ci.yaml?label=CI)](https://github.com/fachschaftinformatik/web/actions/workflows/ci.yaml)
[![Coverage](https://img.shields.io/badge/Coverage-report-blue)](https://github.com/fachschaftinformatik/web/actions/workflows/ci.yaml)

Server-rendered website for the computer science student representatives at Westphalian University.

```text
       site
      /    \
     ▼      ▼
   model   view
```

## Run locally

```sh
nix develop -c go run . -database sqlite.db
```

[Contributing](.github/CONTRIBUTING.md) · [Releases](https://github.com/fachschaftinformatik/web/releases)

Security issues: security@fsv-wh.de.

SPDX-License-Identifier: MPL-2.0

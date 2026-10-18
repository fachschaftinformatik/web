# web

[![Release](https://img.shields.io/github/v/release/fsiwh/web?label=Release)](https://github.com/fsiwh/web/releases/latest)
[![Go](https://img.shields.io/github/go-mod/go-version/fsiwh/web/main?label=Go)](https://github.com/fsiwh/web/blob/main/go.mod)
[![CI](https://img.shields.io/github/actions/workflow/status/fsiwh/web/ci.yaml?branch=main&label=CI)](https://github.com/fsiwh/web/actions/workflows/ci.yaml)
[![Coverage](https://img.shields.io/endpoint?url=https%3A%2F%2Fraw.githubusercontent.com%2Ffsiwh%2Fweb%2Fcoverage%2Fcoverage.json)](https://github.com/fsiwh/web/actions/workflows/ci.yaml)

Server-rendered website for the computer science student representatives at Westphalian University.

```text
           site
         /  |  \
        ▼   ▼   ▼
     model view storage
```

## Run locally

```sh
nix develop -c go run . -database sqlite.db
```

[Contributing](.github/CONTRIBUTING.md)<br>
[Releases](https://github.com/fsiwh/web/releases)

Security issues: security@fsv-wh.de.

SPDX-License-Identifier: MPL-2.0

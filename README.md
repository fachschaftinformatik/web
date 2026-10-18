# web

[![Release](https://img.shields.io/github/v/release/fsiwh/web?label=Release)](https://github.com/fsiwh/web/releases/latest)
[![Go](https://img.shields.io/github/go-mod/go-version/fsiwh/web/main?label=Go)](https://github.com/fsiwh/web/blob/main/go.mod)
[![CI](https://img.shields.io/github/actions/workflow/status/fsiwh/web/ci.yaml?branch=main&label=CI)](https://github.com/fsiwh/web/actions/workflows/ci.yaml)
[![Coverage](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fapi.github.com%2Frepos%2Ffsiwh%2Fweb%2Freleases%2Flatest&query=%24.assets%5B0%5D.label&label=Coverage&color=blue)](https://github.com/fsiwh/web/actions/workflows/ci.yaml)

Server-rendered website for the computer science student representatives at Westphalian University.

```text
           site
         /  |  \
        ▼   ▼   ▼
     model view storage
```

## Getting started

```sh
nix develop -c go run . -database sqlite.db
```

To contribute, see the [contributing guide](.github/CONTRIBUTING.md) for setup,
development and checks. Published versions are listed in [Releases](https://github.com/fsiwh/web/releases).

## Security

[![Vulnerabilities](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fapi.github.com%2Frepos%2Ffsiwh%2Fweb%2Freleases%2Flatest&query=%24.assets%5B1%5D.label&label=Vulnerabilities&color=brightgreen)](https://github.com/fsiwh/web/releases/latest/download/vulnerabilities.json)

Release images are scanned with Trivy. The badge shows fixable High/Critical
vulnerabilities found at build time; releases require zero findings at those
severities. Download the latest [scan report](https://github.com/fsiwh/web/releases/latest/download/vulnerabilities.json)
or [SPDX SBOM](https://github.com/fsiwh/web/releases/latest/download/web.sbom.spdx.json).
Report security issues to [team@fsi-wh.de](mailto:team@fsi-wh.de).

SPDX-License-Identifier: MPL-2.0

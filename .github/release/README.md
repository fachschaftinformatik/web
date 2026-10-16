# Releases

Releases run automatically after a pull request is squash-merged into `main`.
The applicable lint, test, security, and build jobs in the `ci:app` workflow must pass
`release:check` before `release:publish` runs semantic-release.
The highest applicable Conventional Commit since the last `v*` tag determines
the next version:

| Commit | Release |
| --- | --- |
| `feat` | Minor |
| `fix`, `perf` | Patch |
| `chore(deps)` | Patch |
| Breaking change (`!` or `BREAKING CHANGE:`) | Major |
| Other types, including `ci` and `chore(ci)` | None |

Standard SemVer applies below 1.0 too: a feature after 0.1.18 produces 0.2.0,
and a breaking change produces 1.0.0. No source version bump or release commit is
needed. The image build injects the release tag into `main.Version`.

During semantic-release's `prepare` step, Docker Bake builds and pushes the
amd64 application image, `ghcr.io/fachschaftinformatik/web:<version>`. Only after the
build succeeds does semantic-release push the tag and publish an empty GitHub
Release. The workflow then promotes the `latest` alias. Consumers should deploy
the versioned image or its digest. The release job exposes `released` and `tag`
outputs for subsequent deployment jobs.

## Application image

Published images include a BuildKit-generated SPDX SBOM as an OCI attestation
in GHCR. It describes the final runtime image and is published by the existing
Docker Bake release build. The same SPDX document is also attached to the GitHub
Release as `web-<version>.sbom.spdx.json` (for example,
`web-v0.1.22.sbom.spdx.json`) for direct download. Inspect a version's SBOM with:

```sh
docker buildx imagetools inspect ghcr.io/fachschaftinformatik/web:<version> \
  --format '{{ json .SBOM.SPDX }}'
```

One Alpine runtime contains a stripped, path-trimmed static Go executable with
embedded templates and browser assets. The builder keeps sources and generation
tools out of the runtime. `test:go` verifies committed SQLC query generation and
tests all Go packages with race detection. Node and Bun are release tools only.

The application is PID 1, runs as UID 1001 on port 8080 and owns `/data`.
SIGTERM initiates graceful HTTP shutdown. The container health command checks
`/healthz`. No process supervisor or internal API proxy is required.

For version diagnostics:

```sh
docker run --rm ghcr.io/fachschaftinformatik/web:<version> -version
```

### V2 development cutover

The V2 branch replaces the combined V1 runtime. The repository cutover does not
complete the production data migration: account authentication, mail and domain
persistence still need wiring. See [migration status](../../docs/v2-development.md).
The current image serves fixture/in-memory workflows and is not deployed by CI.

When preparing deployment, preserve the existing SQLite data, Garage bucket,
credentials and object keys. The new container listens on port 8080, replacing
V1's public port 80. V1 NextAuth/internal JWT variables and `/api/v1`/Swagger
routes no longer configure the HTML application. Persistent service configuration
and route compatibility belong to the remaining migration work.

Releases publish images only; production deployment remains a separate operation.

The `ci:app` workflow can be dispatched manually; `dry-run` defaults to true.
A dry run still runs full verification, including image build validation. The
publication step checks authentication and version calculation but skips pushing
images, tags, releases, and alias promotion. It must target `main`.

## Recovery

- An image build fails before tagging: rerun the workflow; no Git tag or GitHub
  Release will have been published.
- The tag exists but GitHub Release publication failed: verify that the versioned
  image exists, then create the missing release for that existing tag with an empty
  body. Do not delete the tag or allocate another version to repair publication.
- The release exists but alias promotion failed: rerun only the failed release job
  to promote the latest existing release's alias. The workflow verifies that its
  commit is still the latest released commit before performing this repair.
- A deployment fails: redeploy the already published image version; publishing and
  deployment are separate operations.

## Release-tool security patch

semantic-release depends on `micromatch`, which depends on `braces@3.0.3`.
GHSA-vfj7-8cjw-p6xm has no published fixed version. The committed Bun patch bounds
brace/parenthesis parsing and recursive AST walkers at depth 100, based on the
mitigation discussed in <https://github.com/micromatch/braces/pull/72>.
The regression suite exercises deeply nested and malformed input, direct ASTs,
ordinary expansion, and semantic-release's real commit analyzer.

`bun run audit` excludes only that advisory because registry audits identify the
original version, not Bun patches. Every other high/critical advisory still fails
the check. CI must install with the frozen lockfile and run the patch regression
tests before auditing or releasing. Revisit the patch and exception when a fixed
upstream release is available. The patch is isolated to release tooling.

Run locally from this directory:

```sh
bun install --frozen-lockfile --ignore-scripts
bun run test
bun run audit
```

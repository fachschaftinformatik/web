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
needed. The website's package version is not the deployment version.

During semantic-release's `prepare` step, Docker Bake builds and pushes the
combined amd64 image, `ghcr.io/fachschaftinformatik/web:<version>`. Only after the
build succeeds does semantic-release push the tag and publish an empty GitHub
Release. The workflow then promotes the `latest` alias. Consumers should deploy
the versioned image or its digest. The release job exposes `released` and `tag`
outputs for subsequent deployment jobs.

## Application image

Published images include a BuildKit-generated SPDX SBOM as an OCI attestation
in GHCR. It describes the final runtime image and is published by the existing
Docker Bake release build. Inspect a version's SBOM with:

```sh
docker buildx imagetools inspect ghcr.io/fachschaftinformatik/web:<version> \
  --format '{{ json .SBOM }}'
```

One Alpine runtime contains the stripped, path-trimmed static Go API binary,
database migrations, Node, and Next.js standalone output. Separate builder stages
keep Go sources, generated-code tools, npm, Bun, and other build tools out of the
runtime image. `test:api` verifies the committed Go and OpenAPI generated sources.

Only the Node executable is copied from the official Node Alpine
image. Only musl Sharp binaries are retained, and website dependencies are built
for the target image architecture. CA certificates, timezone data, and the shared
libraries Node requires remain available.

`docker/entrypoint.sh` starts Alpine's `s6` supervisor, which runs the API on port
8080 and the website on port 80. Failed services are restarted by s6. The website
runs as UID 1001; the API retains write access to `/data`. Container shutdown stops
both services. The health check verifies both servers. Next.js proxies `/api/v1`
and `/swagger` to the local API. Only port 80 needs to be routed in production.

For diagnostics, an explicit command bypasses service startup, for example
`docker run --rm ghcr.io/fachschaftinformatik/web:<version> node --version`.

### Migration from separate services

This replaces the `/api` and `/website` release images. On the first deployment:

1. Use the combined versioned image for a single application service.
2. Preserve the existing SQLite bind mount at `/data` and the existing Garage
   service, bucket, credentials, and network connectivity.
3. Combine the existing API and website environment settings, including `DOMAIN`,
   `AUTH_URL`, `AUTH_SECRET`, `INTERNAL_JWT_SECRET`, SMTP, and S3 settings. Internal
   backend routing and the API port are set by the entrypoint.
4. Route the public hostname, including API and Swagger requests, to the
   application service on port 80. Retire the old API and website services as part
   of this cutover; keep the database mount and Garage data.

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

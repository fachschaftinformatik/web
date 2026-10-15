# Releases

Releases run automatically after a pull request is squash-merged into `main`.
API, website, and security checks must succeed before semantic-release runs.
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

During semantic-release's `prepare` step, Docker Bake builds and pushes both
amd64 images concurrently with the calculated version. Only after both builds
succeed does semantic-release push the tag and publish an empty GitHub Release.
The workflow then promotes the `latest` aliases. Consumers should deploy the
versioned images or their digests. The release job exposes `released` and `tag`
outputs for subsequent deployment jobs.

The Release workflow can be dispatched manually; `dry-run` defaults to true.
A dry run checks authentication and version calculation but skips image builds,
tags, releases, and alias promotion. It must target `main`.

## Recovery

- An image build fails before tagging: rerun the workflow. One image may already
  exist, but no Git tag or GitHub Release will have been published.
- The tag exists but GitHub Release publication failed: verify that both versioned
  images exist, then create the missing release for that existing tag with an empty
  body. Do not delete the tag or allocate another version to repair publication.
- The release exists but alias promotion failed: rerun only the failed release job
  to promote the latest existing release's aliases. The workflow verifies that its
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

# Contributing

If you believe you have found a security issue, please contact us at security@fsv-whs.de first before opening any issues or pull requests (re: Responsible Disclosure).

When you want to work on an issue, comment on it first and tell us the approach you want to take.
Please always adhere to these guidelines when contributing.

That said, thank you for your interest in helping improve the project.

## Getting started
There are several ways to contribute to the project:

- Report any potential bugs
- Propose improvements and enhancements
- Implement a (requested) feature or [fix a bug](https://github.com/fachschaftinformatik/web/labels/bug)
- Improve the documentation

If in doubt, reach out to us. You can [open an issue](https://github.com/fachschaftinformatik/web/issues/new) anytime to share ideas or invite feedback.

## Building

The application needs [Go](https://go.dev/) 1.27.2. Node (pinned in
`.node-version`) and Bun are needed only when working on `.github/release`.

Make sure you have the required Go version by running:
```
go version
```

Start the application from the repository root:
```bash
go run .
```

Open <http://127.0.0.1:8090/>. Alternatively, use `docker compose up --build`.
Garage is optional through the `storage` Compose profile.

When changing SQL queries, regenerate the committed SQLC code before building:
```bash
go generate ./internal/data
go build -o /path/to/output/web .
```

Run `go test -race ./...` and `go vet ./...`. Templates and browser assets are
embedded; rebuild the executable after changing them.

## Dependency updates

Renovate runs on Mondays at 08:00 Europe/Berlin, with automatic daylight-saving adjustment.
It opens one pull request per dependency, including lockfile and generated-code changes,
using titles such as `chore(deps): bump golang from 1.27.2 to 1.27.3`.
Updates require review and are not automatically merged.

The workflow uses a dedicated GitHub App installed on this repository. Configure its
App ID as `RENOVATE_APP_ID` and its PEM private key as `RENOVATE_APP_PRIVATE_KEY`
in repository Actions secrets. The App needs repository read/write permissions for
Contents, Pull requests, and Workflows, plus read access to Checks and Commit statuses.

The Renovate workflow can also be started manually. Its `dry-run` input defaults to
true: Renovate checks updates and artifact generation without creating branches or
pull requests. Disable it to publish updates immediately.

Go updates run `go mod tidy` and `go generate ./...`; CI verifies generated SQLC
queries. Bun updates apply to release tooling only.
Indirect Go modules are included. Floating Docker tags remain floating; existing
digest-pinned images receive digest updates.

## CI

The `ci:app` workflow in [ci.yaml](workflows/ci.yaml) runs on pull requests and pushes
to `main`. It uses these job names:

- `test:go` (formatting, vet, workflow/shell lint, query generation, race tests and govulncheck)
- `security:dependencies`, `security:release`
- `security:codeql` (weekly and manual scans for Go, JavaScript/TypeScript and Actions)
- `build:app`
- `release:check`, `release:publish`

Independent jobs run in parallel. Dependency review runs only for pull requests.
Container validation runs on pull requests, scheduled and manual runs. Pushes to
`main` build the container during release publication instead of building it twice.

`release:check` rejects any failed or cancelled check. Branch protection should require `release:check` and
`lint:commit`. The latter runs in the separate `ci:lint` workflow so editing a PR
title does not restart application checks. Keep both required-check names in sync
when renaming jobs.

The dependency-update workflow is named `ci:renovate`, with the job
`maintenance:dependencies`. Deployment is managed separately from CI.

Go setup and caching use the standard `actions/setup-go` action. Production
images build the committed generated files; `test:go` verifies database query
generation and runs the Go tests with race detection. `build:app` builds the
Go application image.

### Security checks

GitHub secret scanning checks repository history and new commits; push protection
blocks supported secrets before they enter the repository.

The full pipeline runs on Mondays at 09:00 Europe/Berlin and can be
started manually. Security checks include:

- Weekly/manual CodeQL with extended security queries for Go, JavaScript/TypeScript,
  and Actions. Findings appear in the repository's Security tab.
- Dependency review on pull requests, failing for newly introduced high or critical
  vulnerabilities in runtime and development dependencies recognized by GitHub.
- `govulncheck` for known vulnerabilities reachable from the Go application.
- `bun audit` against the release-tool lockfile, including development dependencies,
  failing for high or critical vulnerabilities.

Scheduled scans also detect advisories published since the last code change.

The release tooling carries a scoped `braces` patch; see
[release operations](release/README.md#release-tool-security-patch).

## Making changes
If you plan to make any changes, the workflow is typically like this:

1. Make sure you are on the `main` branch and have the latest changes locally
2. Create and switch to a new appropiatley named branch like `api/registration`
3. Make the changes you wish to commit and test them
4. Review your changes
5. Add, commit and push the changes
6. Open a PR to squash-merge the changes into `main`

Example workflow:
```
(1) git branch
 * main
(1) git pull
 Already up to date.
(2) git switch -c api/registration
(3) [...]
(3) go test ./...
(4) git status -s
 M main.go
(4) git diff main.go
 [...]
(5) git add main.go
(5) git commit -m "feat(api): add user registration endpoint"
(5) git push -u origin api/registration
```
Once you have pushed the branch, GitHub will ask if you want to create a PR.
Accept the proposal but make sure to adjust the title and description according to [Commits and Pull Requests](#commits-and-pull-requests) below.

You can block the PR from being merged by converting it to a draft. This is useful if you plan to make further changes.

If you need to make any changes to the PR you can just keep committing and pushing into the branch and the changes will be associated with the PR.

## Opening a PR

1. Title the PR according to [Commits and Pull Requests](#commits-and-pull-requests) below
2. Leave the PR description empty
5. Request review from one of the [maintainers][3]

## Style
Follow Go’s official [Code Review Comments][1]. Prefer clarity over cleverness.

### Commits and Pull Requests
Use the [Conventional Commits][2] spec when writing commit messages and titles:
```
<type>[optional scope][optional !]: <description>
```
Common types are `feat`, `fix`, `docs`, `refactor`, `test`, `perf`, `build`, `ci`, `chore`, and `style`.
Please keep the description short and meaningful.

When a commit or PR directly relates to an existing issue or pull request, append
its number in parentheses where useful: `fix(api): correct password reset logging (#123)`.
Use real, relevant issue or PR numbers; omit the suffix when there is no useful
reference. Commit messages are title-only; PR and release descriptions are empty.

Mark breaking changes with `!` in the title:
```
feat!: replace the application runtime
```

Examples:
```
feat(api): add user registration endpoint
fix(db): enforce unique email constraint
fix(api): correct password reset logging (#123)
```

PR titles should follow the same format. See the [full spec][2] for details.

Pull requests use lowercase Conventional Commit titles, checked by CI. Squash
merging uses the PR title as the final commit title, with the merged PR number
appended as `(#123)` where appropriate. Avoid duplicating the same reference.
`feat` releases a minor version; `fix`, `perf`, and `chore(deps)` release a patch; breaking changes release
a major version. Other changes do not release by themselves.

After a merge, `ci:app` runs the applicable checks. Once `release:check` succeeds,
`release:publish` uses semantic-release to build and publish versioned images and push the version tag,
and creates the GitHub Release automatically. See [release operations](release/README.md)
for dry runs, recovery, and the release-tooling security patch.

Relevant documentation:
- [Conventional Commits 1.0.0][2]
- [Code Review Comments][1]
- [Effective Go](https://go.dev/doc/effective_go)

[1]: https://go.dev/wiki/CodeReviewComments
[2]: https://www.conventionalcommits.org/en/v1.0.0/#specification
[3]: https://github.com/fachschaftinformatik/web/blob/main/MAINTAINERS

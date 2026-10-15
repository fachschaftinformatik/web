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

To build the project you will need to have [Go](https://go.dev/) (version 1.27.2 or newer), [Node.js](https://nodejs.org/) (version 26.12.0, pinned in `.node-version`), and [Bun](https://bun.sh/) installed.

Make sure you have the required Go version by running:
```
go version
```

Start the local API and Garage object storage from the repository root:
```bash
docker compose up --build api garage
```

Configure `website/.env.local` as described in the [README](../README.md#getting-started), then start the frontend:
```bash
cd website
bun i
bun run dev
```

To build the backend, make sure to `go generate` the required code first:
```bash
go generate ./...
go build -o web .
```

If you have made any changes to the API or DTOs, make sure to re-generate the client too:
```
cd website
bun run gen:api
```

## Dependency updates

Renovate runs on Mondays at 08:00 Europe/Berlin, with automatic daylight-saving adjustment.
It opens one pull request per dependency, including lockfile and generated-code changes,
using titles such as `chore(deps): bump next from 16.4.1 to 16.4.2`.
Updates require review and are not automatically merged.

The workflow uses a dedicated GitHub App installed on this repository. Configure its
App ID as `RENOVATE_APP_ID` and its PEM private key as `RENOVATE_APP_PRIVATE_KEY`
in repository Actions secrets. The App needs repository read/write permissions for
Contents, Pull requests, and Workflows, plus read access to Checks and Commit statuses.

The Renovate workflow can also be started manually. Its `dry-run` input defaults to
true: Renovate checks updates and artifact generation without creating branches or
pull requests. Disable it to publish updates immediately.

Go updates run `go mod tidy` and `go generate ./...`; Go and Bun updates regenerate
the website API client. Existing API and website CI checks validate the results.
Indirect Go modules are included. Floating Docker tags remain floating; existing
digest-pinned images receive digest updates.

## Security checks

GitHub secret scanning checks repository history and new commits; push protection
blocks supported secrets before they enter the repository.

The Security workflow runs on pull requests, pushes to `main`, and Mondays at
09:00 Europe/Berlin. It can also be started manually. Its parallel jobs include:

- CodeQL with extended security queries for Go, JavaScript/TypeScript, and Actions.
  Findings appear in the repository's Security tab and pull-request code scanning.
- Dependency review on pull requests, failing for newly introduced high or critical
  vulnerabilities in runtime and development dependencies recognized by GitHub.
- `govulncheck` for known vulnerabilities reachable from the Go application.
- `bun audit` against the frontend lockfile, including development dependencies,
  failing for high or critical vulnerabilities.

Scheduled scans also detect advisories published since the last code change.

The frontend overrides `js-yaml` to a patched release because the API generator's
parser pins an affected version. It also replaces `fast-glob` with `tinyglobby`:
Next's ESLint plugin uses its compatible `globSync` API, avoiding the unpatched
`braces` dependency. `bun run test:eslint-globs` verifies the plugin's root-directory
discovery. Revisit these overrides when the upstream dependencies are updated.

## Making changes
If you plan to make any changes, the workflow is typically like this:

1. Make sure you are on the `main` branch and have the latest changes locally
2. Create and switch to a new appropiatley named branch like `api/registration`
3. Make the changes you wish to commit and test them
4. Review your changes
5. Add, commit and push the changes
6. Open a PR to merge the changes into `main`

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
 M cmd/web/main.go
(4) git diff api/api.go
 [...]
(5) git add api/api.go
(5) git commit -m "feat(api): add user registration endpoint"
(5) git push -u origin api/registration
```
Once you have pushed the branch, GitHub will ask if you want to create a PR.
Accept the proposal but make sure to adjust the title and description according to [Commits and Pull Requests](#commits-and-pull-requests) below.

You can block the PR from being merged by converting it to a draft. This is useful if you plan to make further changes.

If you need to make any changes to the PR you can just keep committing and pushing into the branch and the changes will be associated with the PR.

## Opening a PR

1. Title the PR according to [Commits and Pull Requests](#commits-and-pull-requests) below
2. Include helpful information in the description and, if applicable, include information on how to test manually
3. For a bug fix, explain or show an example of the behavior before and after the change
5. Request review from one of the [maintainers][3]
6. Link issues with “Fixes #123” or “Closes #123” in the description

## Style
Follow Go’s official [Code Review Comments][1]. Prefer clarity over cleverness.

### Commits and Pull Requests
Use the [Conventional Commits][2] spec when writing commit messages and titles:
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```
Common types are `feat`, `fix`, `docs`, `refactor`, `test`, `perf`, `build`, `ci`, `chore`, and `style`.
Please keep the description short and meaningful.

Breaking changes must include a footer:
```
BREAKING CHANGE: explain what changed and how to migrate
```

Examples:
```
feat(api): add user registration endpoint
fix(db): enforce unique email constraint
```

PR titles should follow the same format. See the [full spec][2] for details.

Relevant documentation:
- [Conventional Commits 1.0.0][2]
- [Code Review Comments][1]
- [Effective Go](https://go.dev/doc/effective_go)

[1]: https://go.dev/wiki/CodeReviewComments
[2]: https://www.conventionalcommits.org/en/v1.0.0/#specification
[3]: https://github.com/fachschaftinformatik/web/blob/main/MAINTAINERS

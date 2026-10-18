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

Use [Nix](https://nixos.org/download/) with flakes enabled for Go 1.27.2 and
browser-asset generation. Node (pinned in the CI workflow) and Bun are needed
only when working on `.github/release`.

Make sure you have the required Go version by running:
```
nix develop -c go version
```

Use a writable copy of the production database and object storage for development.
The application does not seed members or organization content. Legacy content
imports and data conversion are managed outside this repository.
Start the application from the repository root with explicit account-domain policy:
```bash
export ACCOUNT_DOMAINS=studmail.w-hs.de,fsv-wh.de
export REGISTRATION_DOMAINS=studmail.w-hs.de
nix develop -c go run . -database .dev-data/local/sqlite.db -blobs .dev-data/local/blobs
```

Open <http://127.0.0.1:8090/>. Alternatively, use `docker compose up --build`.
Garage is optional through the `storage` Compose profile; its configuration is
embedded in `compose.yaml` using inline configs (Docker Compose 2.23.1 or newer).

When changing SQL queries, regenerate the committed SQLC code before building:
```bash
nix develop -c go generate ./internal/database
nix develop -c go build -o /path/to/output/web .
```

Use `nix develop` for the locked Go toolchain, or prefix a command with
`nix develop -c`, for example `nix develop -c go test -race ./...`. Its Go version
must match `go.mod`, which also selects the version in CI; keep Docker aligned.
The shell supplies development account-domain defaults when those variables are
unset, preserving explicit values (including empty registration domains).

Run `go test -race ./...` and `go vet ./...`. Templates and browser assets are
embedded; rebuild the executable after changing them.

`nix develop` prepares the GOV.UK, HTMX, TinyMCE, language-pack and GDS illustration
assets automatically for Go's `embed`. Run `nix run .#vendor` to prepare them
separately when using your own Go installation. Generated files are ignored by
Git; handwritten SCSS in `view/styles/`, JavaScript and the event placeholder remain
source-controlled. Nix compiles the SCSS to CSS for embedding.
`flake.lock` pins the Nix toolchain, including Dart Sass; `flake.nix` pins
upstream versions/revisions and verifies every download with a SHA-256 hash.
Use `nix build .#assets` to build without modifying the working tree. Updating an
asset requires changing its version/revision and hash, then checking the build.
Update the toolchain with `nix flake update nixpkgs`. `nix flake check` builds the
hash-verified assets. Docker generates the same assets in a separate Nix build
stage, so building an image does not require a host Nix installation. Runtime
images contain the Go executable with embedded assets, without Nix.

Format HTML templates with Prettier 3.6.2 and `prettier-plugin-go-template` 0.0.15,
using the `go-template` parser, CSS whitespace sensitivity and a print width of 100.
Use whitespace-trimming template actions inside URL and enumerated attributes.

## Security defaults

Keep authorization in models and route policies, use parameterized SQL, and pass
untrusted text through `html/template` or the rich-text allowlist. Browser headers
disable MIME sniffing, cross-origin framing, foreign form targets and base URLs.
The CSP is deliberately limited to those protections; it is not a script allowlist.
Compose runs the non-root app with a read-only root filesystem, no capabilities,
no privilege escalation, and bounded temporary storage; `/data` remains writable.
Container bases and CI actions are pinned; image scans block fixable high/critical
OS and library vulnerabilities. Renovate maintains the pins.

Sessions use random cookie-only IDs with hashed database verifiers, an eight-hour
server-side expiry, login rotation, and revocation on password reset or promotion.
Email links expire, are consumed only by a CSRF-protected POST, and cannot be
reused. Token pages disable caching and referrers. The mail outbox necessarily
contains usable links until delivery; protect the database and backups accordingly.
SMTP requires STARTTLS except for a loopback development relay.

Production's HTTPS proxy must handle TLS/HSTS, request throttling (especially
login, registration and password reset), and access logs without query strings or
credentials. Set `DOMAIN` to the public HTTPS URL so session cookies are secure.
These defaults follow the [OWASP Top 10](https://owasp.org/Top10/2025/),
[headers](https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Headers_Cheat_Sheet.html),
[containers](https://cheatsheetseries.owasp.org/cheatsheets/Docker_Security_Cheat_Sheet.html)
and [supply chain](https://cheatsheetseries.owasp.org/cheatsheets/Software_Supply_Chain_Security_Cheat_Sheet.html) guidance,
plus the [session](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)
and [password reset](https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html) guidance.

## Source layout

The application keeps its MVC responsibilities in separate packages:

- `model/`: feature packages for business rules, authorization and persistence.
- `site/`: the controller layer, with HTTP routes, request handling and sessions.
- `view/`: presentation types, display formatting, embedded templates and assets.
- `storage/`: object-storage interfaces and disk/S3 implementations.
- `internal/`: shared implementation helpers, configuration and database migrations.
- `flake.nix` and `flake.lock`: development environment and pinned browser-asset build.

`main.go` and `application.go` handle startup and dependency wiring. Models must
not depend on `site` or `view`. Views have independent presentation structs and
must not depend on `model`, `site`, storage or the database. Model-to-view mapping,
validation calls and authorization decisions belong in `site`; formatting stays
in `view`. `internal/architecture` checks these import boundaries, including
transitive local dependencies and test files, during `go test ./...`.

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

- `test:go` (formatting, vet, workflow lint, query generation, race tests and govulncheck)
- `test:assets` (locked Nix asset build with verified source hashes)
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

`test:go` uses the locked Nix development environment, verifies database query
generation and runs the Go tests with race detection. Production images build
the committed generated SQLC files and Nix-generated browser assets;
`build:app` validates the complete application image.

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
(5) git commit -m "feat(auth): add user registration"
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
Use [Conventional Commits 1.0.0][2] for commit messages and PR titles:
```
<type>[optional scope][optional !]: <description>
```
Common types are `feat`, `fix`, `docs`, `refactor`, `test`, `perf`, `build`, `ci`,
`chore`, `style`, and `revert`. Choose the type that describes the actual change.

Follow the [Go commit message guidance](https://go.dev/wiki/CommitMessage) for
clear, imperative wording, using the format and limits below.

- Keep the description to **42 characters maximum**, excluding the type, scope,
  breaking marker, colon and following space.
- Use lowercase only. Descriptions contain only letters `a`–`z`, digits and spaces:
  no punctuation or special characters such as commas, semicolons or periods.
  The Conventional Commit prefix may contain `(`, `)`, `!` and `:`.
- Use an optional scope only for an affected code area, such as `web`, `db`,
  `auth`, `storage`, `deps`, `ci` or `release`. Keep scopes lowercase and
  alphanumeric. Omit the scope for cross-cutting changes; versions such as `v2`
  are not scopes.
- Prefer simple, human-readable wording over jargon, cleverness or compressed
  phrases. Say clearly what changed.
- Mark incompatible changes with `!` immediately before the colon. Describe what
  breaks, rather than marking an internal refactor as breaking by default.
- Commit messages are title-only. Do not append issue or PR numbers; use GitHub
  links for references. PR and release descriptions are empty.

Examples:
```
feat(auth): add user registration
fix(db): enforce unique email constraint
fix(auth): correct password reset logging
feat(auth)!: require token based sessions
```

PR titles follow the same rules. When squash merging, use the PR title verbatim
as the commit title and remove any automatically appended PR number.
`feat` releases a minor version; `fix`, `perf`, and `chore(deps)` release a patch; breaking changes release
a major version. Other changes do not release by themselves.

After a merge, `ci:app` runs the applicable checks. Once `release:check` succeeds,
`release:publish` uses semantic-release to build and publish versioned images and push the version tag,
and creates the GitHub Release automatically. See [release operations](release/README.md)
for dry runs, recovery, and the release-tooling security patch.

Relevant documentation:
- [URL convention and canonical tree](urls.md)
- [Canonical code and UI naming](naming.md)
- [Temporary proxy redirect plan](url-redirects.md)
- [Conventional Commits 1.0.0][2]
- [Code Review Comments][1]
- [Effective Go](https://go.dev/doc/effective_go)

[1]: https://go.dev/wiki/CodeReviewComments
[2]: https://www.conventionalcommits.org/en/v1.0.0/#specification
[3]: https://github.com/fachschaftinformatik/web/blob/main/MAINTAINERS

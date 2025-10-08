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
To build the project you will need to have both `make` and [Go](https://go.dev/) installed.
If you are on either Linux or macOS, chances are `make` is already installed, in which case you will only need to install [Go](https://go.dev/).

The project uses several features that are only in Go version 1.24 and newer (most notably `go tool` and the [new net/http router](https://go.dev/blog/routing-enhancements)),
so make sure you have the correct version by running:
```
go version
```

If everything is setup correctly, you can build the server by running:
```
make build
```

To run the test suite:
```
make test
```
Always run tests before committing any changes.

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
(3) make test
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

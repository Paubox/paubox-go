# Contributing to paubox-go

Thank you for your interest in contributing!

By submitting a contribution, you agree that it is licensed under the project's [Apache 2.0 license](LICENSE).

## Getting started

```bash
git clone https://github.com/paubox/paubox-go
cd paubox-go
go mod download
```

No external tools are needed to run tests — only a Go 1.23+ toolchain.

## Running tests

```bash
# All tests with race detector
go test -race ./...

# With coverage
go test -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Single test
go test -run TestSendMessage_HappyPath ./...
```

All tests use `httptest.Server`. **There are no live API calls in the test suite.**

## Running the linter

```bash
# Install once (golangci-lint v2 — required for the v2 .golangci.yml config)
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2

# Run
golangci-lint run
```

## Running govulncheck

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## Checking for breaking API changes

CI runs [`apidiff`](https://pkg.go.dev/golang.org/x/exp/cmd/apidiff) against the
base branch on every pull request and fails on an incompatible change to the
public API. To reproduce a failure locally:

```bash
go install golang.org/x/exp/cmd/apidiff@latest

git worktree add /tmp/baseline origin/master
(cd /tmp/baseline && apidiff -w /tmp/baseline.api .)
apidiff -incompatible /tmp/baseline.api .
```

Empty output means the change is backward compatible. Anything printed is a
break, and the options are to make the change additive or to land it as a major
release — see **Releasing** below.

## Pull request titles

Titles must follow [Conventional Commits](https://www.conventionalcommits.org)
and are checked by CI. Because pull requests are squash-merged, the title
becomes the commit subject that release-please reads to decide the next version:

| Title prefix | Effect on the next release |
| --- | --- |
| `fix:` | patch bump |
| `feat:` | minor bump |
| `feat!:` or a `BREAKING CHANGE:` footer | major bump |
| `docs:`, `chore:`, `ci:`, `test:`, `refactor:`, `style:`, `build:`, `perf:`, `revert:`, `deps:` | no bump |

Write the title for the changelog, since that is where it ends up verbatim.

## Pull request expectations

- All tests must pass: `go test -race ./...`
- Lint must be clean: `golangci-lint run`
- New endpoints require tests covering: happy path, validation errors, and at least 400/401/404 error responses
- Public API additions require documentation in README.md
- **Do not edit CHANGELOG.md or the `Version` constant** — release-please owns both
- No external runtime dependencies — see CLAUDE.md for the full rules

## Releasing

Releases are automated. Merging to `master` makes release-please open or update
a release pull request that bumps `Version` in `paubox.go` and writes
CHANGELOG.md. Merging *that* pull request tags `vX.Y.Z` and publishes the GitHub
release; the Go module proxy then serves the version straight from the tag,
since Go has no separate publish step.

Edit the release pull request directly if the generated notes need prose, such
as migration instructions — the edits are preserved on merge.

### Major versions

Go encodes the major version in the module path, so releasing 2.0.0 means
`go.mod` must declare `module github.com/paubox/paubox-go/v2` and every internal
import must be updated. Make that change in the same pull request as the
breaking change; a test fails the build if the module path and `Version`
disagree. Consumers stay on 1.x until they change their import paths, so weigh a
major bump carefully.

## Reporting security issues

Do **not** open a public issue for security vulnerabilities. Email security@paubox.com instead. See [SECURITY.md](SECURITY.md).

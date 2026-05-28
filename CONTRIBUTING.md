# Contributing to paubox-go

Thank you for your interest in contributing!

## Developer Certificate of Origin (DCO)

All commits must be signed off. Add `-s` to your commit command:

```bash
git commit -s -m "your message"
```

This adds a `Signed-off-by` trailer certifying that you have the right to submit the contribution under the Apache 2.0 license. See [developercertificate.org](https://developercertificate.org) for the full text.

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

## Pull request expectations

- All tests must pass: `go test -race ./...`
- Lint must be clean: `golangci-lint run`
- New endpoints require tests covering: happy path, validation errors, and at least 400/401/404 error responses
- Public API additions require documentation in README.md and CHANGELOG.md
- No external runtime dependencies — see CLAUDE.md for the full rules

## Reporting security issues

Do **not** open a public issue for security vulnerabilities. Email security@paubox.com instead. See [SECURITY.md](SECURITY.md).

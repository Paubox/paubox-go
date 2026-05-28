# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-05-28

Schema corrections for the Dynamic Templates endpoints, validated against the
live Paubox API. Contains breaking changes to template ID types and method
signatures.

### Changed (breaking)
- `Template.ID` is now `int64` (was `string`) — the API returns numeric IDs.
- `GetTemplate`, `UpdateTemplate`, and `DeleteTemplate` now take `id int64`
  (was `string`).
- `Template.CreatedAt` and `Template.UpdatedAt` are now `*time.Time` (was
  `time.Time`); they are only populated on a single-template fetch.
- `CreateTemplate` and `UpdateTemplate` now return `*TemplateMutationResponse`
  (was `*Template`). The API confirms the operation with a message and does
  **not** return the template object or its ID. To obtain a newly created
  template's ID, call `ListTemplates` and match on `Name`.

### Fixed
- `ListTemplates` now decodes the bare JSON array the API actually returns
  (previously expected a `{"templates":[…]}` object wrapper and failed to
  unmarshal).

### Added
- `Template.APICustomerID` (`int64`) and `Template.Metadata`
  (`map[string]any`) fields returned by the API.
- `TemplateMutationResponse` and `TemplateMutationParams` types.

[0.2.0]: https://github.com/paubox/paubox-go/releases/tag/v0.2.0

## [0.1.1] - 2026-05-28

Initial public release.

### Added
- Paubox Go SDK for the Email API
- `Client` with functional options: `WithBaseURL`, `WithHTTPClient`, `WithTimeout`, `WithRetry`, `WithUserAgent`
- **Messages**: `SendMessage`, `SendBatch`, `GetEmailDisposition`
- **Dynamic Templates**: `ListTemplates`, `GetTemplate`, `CreateTemplate`, `UpdateTemplate`, `DeleteTemplate`, `SendTemplatedMessage`
- `PauboxError` type with `errors.Is` / `errors.As` support and HTTP status-code sentinels
- Automatic retry with exponential backoff + jitter on GET / 429 / 5xx
- `Authorization: Token token=` header set automatically on every request
- TLS 1.2 minimum enforced on the default HTTP client
- `Ptr[T]` generic helper for optional pointer-typed fields
- Full `httptest`-based test suite — no live API calls required
- Examples: `send_single`, `send_batch`, `dynamic_template`, `send_templated`
- `.golangci.yml` strict linter configuration (golangci-lint v2)
- GitHub Actions CI: test (Go 1.23–1.26), lint, govulncheck
- `SECURITY.md` with vulnerability disclosure policy and `NOTICE` (Apache 2.0)

### Requirements
- Requires Go 1.23 or later.

[Unreleased]: https://github.com/paubox/paubox-go/compare/v0.2.0...HEAD
[0.1.1]: https://github.com/paubox/paubox-go/releases/tag/v0.1.1

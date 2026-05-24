# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `FormsClient` — unauthenticated client for the Paubox Forms API (`pb_rforms` service, base URL `https://apx.paubox.com/forms`). No API key required.
- `GetForm(ctx, formID)` — retrieves form metadata and field schema (`GET /public/form_data/:id`). Returns `*Form` including `FormJSON` with typed `[]FormField`.
- `SubmitForm(ctx, formID, FormSubmission)` — submits a form response with optional base64-encoded file attachments (`POST /api/forms/:id/submissions`).
- `FormsOption` functional options: `WithFormsBaseURL`, `WithFormsHTTPClient`, `WithFormsTimeout`, `WithFormsRetry`, `WithFormsUserAgent`.
- Example: `examples/forms` demonstrating both methods end-to-end.

### Changed
- Set minimum Go version to 1.23 (was 1.26 in 0.1.0, but the code only uses 1.22+ features). CI now tests against Go 1.23, 1.24, 1.25, and 1.26.

## [0.1.0] - 2026-05-19

### Added
- Initial release of the Paubox Go SDK (Email API)
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
- `.golangci.yml` strict linter configuration
- GitHub Actions CI: test (Go 1.22 + 1.23), lint, govulncheck
- `SECURITY.md` with vulnerability disclosure policy

[Unreleased]: https://github.com/paubox/paubox-go/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/paubox/paubox-go/releases/tag/v0.1.0

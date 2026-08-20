# Changelog

All notable changes to this project will be documented in this file.

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Entries below 1.0.0 were written by hand; from 1.0.0 onward this file is
maintained by release-please, and pending changes live in the open release pull
request rather than in an "Unreleased" section here.

## [1.0.0] - 2026-08-20

First stable release. The exported API of the `paubox` package is now covered
by the compatibility policy in [README.md](README.md#versioning--compatibility);
breaking changes from here on require a 2.x release and a new module import
path (`github.com/paubox/paubox-go/v2`).

This release also adds Paubox Forms support and completes the migration off
the legacy `api.paubox.net` host. See **Migrating from 0.2.0** below.

### Migrating from 0.2.0

1. **Drop the username argument from `New`.** The Email API no longer uses a
   username for authentication or in request URLs.

   ```go
   // before
   client, err := paubox.New(apiKey, username)
   // after
   client, err := paubox.New(apiKey)
   ```

2. **Remove any hardcoded `api.paubox.net` base URL.** The default is now
   `https://api.paubox.com/v1`. If you passed `WithBaseURL` only to pin the
   old host, delete the option entirely; if you pointed it at a staging host,
   confirm the staging equivalent no longer expects a `{username}` path
   segment.

3. **No changes are required for templates, messages, dispositions, retries,
   or error handling** — those APIs are unchanged since 0.2.0.

### Changed (breaking)
- `New` no longer takes a username: the signature is now
  `New(apiKey string, opts ...Option)` (was `New(apiKey, username string,
  opts ...Option)`). The Paubox Email API no longer uses a username for
  authentication or in URLs; the `Authorization: Token token=<key>` header is
  unchanged.
- The default Email API base URL is now `https://api.paubox.com/v1` (was
  `https://api.paubox.net/v1`).
- Email API request URLs no longer contain a `{username}` path segment —
  `endpointURL` now appends the endpoint path directly to the base URL
  (e.g. `https://api.paubox.com/v1/messages`).

### Added
- **Paubox Forms API** support via a new `FormsClient` (`paubox.NewForms(apiKey, opts...)`):
  - Authenticates with a **scoped API key** carrying the `forms` scope, sent as
    `Authorization: Bearer <key>`. The Email API's `Token token=` format is
    unchanged and remains Email-only.
  - An empty API key yields a **public-only client** on which only the
    unauthenticated endpoints (`GetPublicForm`, `SubmitForm`) work; protected
    methods fail fast with an error.
  - **Forms CRUD**: `ListForms`, `CreateForm`, `GetForm`, `UpdateForm`
  - **Lifecycle & stats**: `ArchiveForm`, `UnarchiveForm`, `CopyForm`, `GetFormStats`
  - **Public endpoints (no auth)**: `GetPublicForm`, `SubmitForm` (attachment
    contents are raw bytes; the SDK base64-encodes them on the wire)
  - **Submissions**: `ListFormSubmissions`, `ExportFormSubmissionsCSV`,
    `ExportFormSubmissionCSV`, `ExportFormSubmissionPDF` (exports return raw
    file bytes)
  - Functional options: `WithFormsBaseURL`, `WithFormsHTTPClient`,
    `WithFormsTimeout`, `WithFormsRetry`, `WithFormsUserAgent`
  - Default base URL `https://api.paubox.com/v1/forms` (assumes the production
    gateway forwards the remaining path unchanged; override with
    `WithFormsBaseURL`)
  - Same retry/backoff semantics as the Email client (GET/DELETE retry on
    429/5xx; PUT is treated as non-idempotent like POST/PATCH)
  - The Forms `{"message": "..."}` error envelope is parsed into the existing
    `*PauboxError` type, so `errors.Is` / `errors.As` and the status-code
    sentinels work unchanged across both APIs

### Added (tooling)
- Exported `paubox.Version` constant, reported in the `User-Agent` header and
  kept in lockstep with this changelog by a test.
- Releases are now cut by release-please: merging its release pull request
  bumps `paubox.Version`, updates this file, and tags `vX.Y.Z`.
- CI runs `apidiff` on every pull request and fails the build on an
  incompatible change to the public API unless the pull request is labelled
  `breaking change allowed`.
- A test asserts the go.mod module path carries the `/vN` suffix required at
  v2 and above, so a major bump cannot ship a tag the module proxy rejects.

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

[1.0.0]: https://github.com/paubox/paubox-go/compare/v0.2.0...v1.0.0
[0.1.1]: https://github.com/paubox/paubox-go/releases/tag/v0.1.1

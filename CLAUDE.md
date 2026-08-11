# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## Commands

```bash
# Test (exclude examples — they require live credentials)
PKGS=$(go list ./... | grep -v '/examples/')
go test -race -coverprofile=coverage.out -covermode=atomic $PKGS
go tool cover -func=coverage.out | grep total   # must be ≥85%

# Single test
go test -run TestSendMessage_HappyPath ./...

# Lint (golangci-lint v2)
~/go/bin/golangci-lint run ./...

# Vulnerability check
~/go/bin/govulncheck ./...
```

Install linter/vulncheck (if missing):
```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
```

---

## Architecture

**Module:** `github.com/paubox/paubox-go` — HIPAA-compliant SDK. **No external runtime dependencies** (stdlib only). Scope is the Email API (transactional messages + dynamic templates) **plus Paubox Forms** (forms CRUD, lifecycle, public submission, submission exports). The Paubox Marketing API is intentionally out of scope.

Two clients, two choke-points, one shared retry loop (`client.go:doHTTP()`):

- **Email** (`Client`): all calls flow through `client.go:do()`. It sets `Authorization: Token token=<key>` — the only place this header is written, don't add it elsewhere; the format is Email-only. `doJSON()` wraps `do()` with JSON marshal/unmarshal. `doMultipartTemplate()` in `templates.go` wraps `do()` for multipart uploads. Base URL pattern: `https://api.paubox.net/v1/{username}/{path}` — the username is embedded by `endpointURL()`.
- **Forms** (`FormsClient`): all calls flow through `forms_client.go:do()`. Auth is a **scoped API key** carrying the `forms` scope, sent as `Authorization: Bearer <key>` — set only in `forms_client.go:do()`, and only when a key is present. An empty key is legal: it yields a public-only client (only `GetPublicForm`/`SubmitForm` work; every protected method fails fast via `requireKey`). URL building is `baseURL + path` with the service routes verbatim (`/api/forms`, `/public/form_data/{id}`) — no username segment. `defaultFormsBaseURL` is `https://api.paubox.com/v1/forms`, which assumes the production gateway forwards the remaining path unchanged; `WithFormsBaseURL` overrides (e.g. a local Forms service on `http://localhost:3000`).

`doHTTP()` in `client.go` buffers the body once for replay across retries and runs the retry loop for both clients — keep behaviour changes there in sync with both.

### Forms file layout

- `forms_client.go` — `FormsClient`, `NewForms`, `WithForms*` options, `do()`/`doJSON()`
- `forms_types.go` — all public Forms request/response types
- `forms.go` — core CRUD: `ListForms`, `CreateForm`, `GetForm`, `UpdateForm`
- `forms_lifecycle.go` — `ArchiveForm`, `UnarchiveForm`, `CopyForm`, `GetFormStats`
- `forms_public.go` — public (no-auth) endpoints: `GetPublicForm`, `SubmitForm`
- `forms_submissions.go` — `ListFormSubmissions` + CSV/PDF exports (raw bytes via `exportRaw`)

### Retry logic

- GET and DELETE: retry on 429 + 5xx, up to `RetryConfig.MaxAttempts`
- POST, PUT, and PATCH: **not retried** by default (`RetryNonIdempotent: false`). The Forms service uses PUT for update — treated like POST: non-idempotent.
- Backoff: exponential × 2^(attempt-1), capped at `WaitMax`, ±20% jitter via `math/rand/v2`, honours `Retry-After` header; context cancellation respected via `select`

### Error model

The Email API returns `{"errors":[{"code":int,"title":"...","details":"..."}]}`. `parseAPIError()` in `errors.go` handles this and produces `*PauboxError`. Sentinel errors (`ErrUnauthorized`, `ErrNotFound`, etc.) match by `StatusCode` only via `errors.Is()`.

The Forms service envelope is different: `{"message":"..."}` (sometimes an empty body, sometimes plain text on the CSV/PDF export error paths). `parseFormsAPIError()` in `errors.go` tries the `{"message":...}` envelope first (→ `Title`), falls back to the Email `{"errors":[...]}` envelope, and finally to `http.StatusText`. It produces the same `*PauboxError`, so the sentinels work identically for both APIs.

### Dynamic template uploads

`CreateTemplate` and `UpdateTemplate` use `multipart/form-data`, not JSON. `internal/multipart.go:BuildTemplateForm()` builds the form with fields `data[name]` (text) and `data[body]` (binary `.hbs` file). The unexported `buildTemplateForm(io.Writer, …)` variant exists solely for error-path testing with `errWriter`.

### template_values footgun

The Paubox API requires `template_values` to be a **JSON-encoded string**, not an object:
```json
{"data": {"template_values": "{\"name\":\"Alice\"}"}}
```
`SendTemplatedMessage` accepts `map[string]any` from callers and marshals it to a string internally. This is hidden from callers entirely.

The Forms API is the opposite on write but similar on read: `SubmitForm` sends `form_data` as a **real JSON object** (no string encoding), while `FormSubmission.FormData` on the read side arrives as a **JSON-encoded string** (the server re-serializes it) — it stays a `string` in the SDK.

### MessageHeaders custom marshalling

`MessageHeaders` implements `MarshalJSON` to flatten `CustomHeaders` map entries into the top-level JSON object alongside the standard fields (`subject`, `from`, `reply-to`, etc.).

---

## Adding a new endpoint

1. Fetch the live schema from `https://docs.paubox.com/api-reference/` — don't guess from naming patterns.
2. Add public request/response types to `*_types.go`; keep unexported wire types separate.
3. Implement the method following the pattern in `messages.go` or `templates.go` (Email) or `forms.go` (Forms): validate first, then call `doJSON` / `doMultipartTemplate` / `exportRaw`. Forms endpoints take the service route verbatim against the Forms base URL.
4. Validation errors must be prefixed `"paubox: MethodName: "`.
5. Tests (in the same package, not `_test`): happy path, correct method+path, request body assertions, all validation cases, ≥400/401/404 error responses via `httptest.Server`.
6. Update `README.md` (add a `<details>` usage block) and `CHANGELOG.md`.

---

## Testing conventions

All tests use `httptest.NewServer`/`httptest.NewTLSServer` — no live API calls. Test files are in `package paubox` (same package) so unexported helpers like `validateMessage` are accessible. Standard helpers: `newTestClient(t, srv)` in `client_test.go`, `respondJSON(w, code, body)` in `messages_test.go`. Unused `http.HandlerFunc` parameters must be named `_` to satisfy `revive`.

---

## Security constraints

- The API key lives only in `c.apiKey` (on both `Client` and `FormsClient`); never log it or include it in errors.
- `PauboxError.Raw` is for debugging only; SDK code must never log it (may contain PHI).
- `InsecureSkipVerify` must never be set, including in tests.
- TLS 1.2 minimum must be maintained on the default transport.
- Never add a runtime dependency without explicit human approval.

---

## Release

1. Move `[Unreleased]` entries in `CHANGELOG.md` to a versioned section.
2. Update `defaultUserAgent` in `client.go`.
3. Run the full suite clean.
4. `git tag -s vX.Y.Z -m "Release vX.Y.Z" && git push origin vX.Y.Z`

Never bump the major version, add external runtime deps, add Marketing API endpoints, or change the `Authorization` header format without explicit human approval.

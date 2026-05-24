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

**Module:** `github.com/paubox/paubox-go` — HIPAA-compliant transactional email SDK. **No external runtime dependencies** (stdlib only). Scope is Email API + Paubox Forms (public endpoints only). The Paubox Marketing API is intentionally out of scope.

### Authenticated Email client (`client.go`)

All Email API calls flow through a single choke-point: `client.go:do()`. It buffers the body once for replay across retries, sets `Authorization: Token token=<key>` (the only place this header is written — don't add it elsewhere), and handles the retry loop. `doJSON()` wraps `do()` with JSON marshal/unmarshal. `doMultipartTemplate()` in `templates.go` wraps `do()` for multipart uploads.

Base URL pattern: `https://api.paubox.net/v1/{username}/{path}` — the username is embedded by `endpointURL()`.

### Unauthenticated Forms client (`forms.go`)

`FormsClient` mirrors the `Client` pattern but carries **no API key and no username**. Its `do()` never sets an `Authorization` header — this is enforced by design (no `apiKey` field on the struct). Base URL: `https://apx.paubox.com/forms` (separate host from the Email API). URL pattern: `baseURL + path` — no username embedded.

Public endpoints implemented:
- `GET /public/form_data/:form_id` → `GetForm`
- `POST /api/forms/:form_id/submissions` → `SubmitForm` (returns 201 No Content)

`FormsClient.backoff()` duplicates `Client.backoff()` — do not refactor them into a shared helper without explicit approval, as it would couple the two clients.

### Retry logic

- GET and DELETE: retry on 429 + 5xx, up to `RetryConfig.MaxAttempts`
- POST and PATCH: **not retried** by default (`RetryNonIdempotent: false`)
- Backoff: exponential × 2^(attempt-1), capped at `WaitMax`, ±20% jitter via `math/rand/v2`, honours `Retry-After` header; context cancellation respected via `select`

### Error model

The API returns `{"errors":[{"code":int,"title":"...","details":"..."}]}`. `parseAPIError()` in `errors.go` handles this and produces `*PauboxError`. Sentinel errors (`ErrUnauthorized`, `ErrNotFound`, etc.) match by `StatusCode` only via `errors.Is()`.

### Dynamic template uploads

`CreateTemplate` and `UpdateTemplate` use `multipart/form-data`, not JSON. `internal/multipart.go:BuildTemplateForm()` builds the form with fields `data[name]` (text) and `data[body]` (binary `.hbs` file). The unexported `buildTemplateForm(io.Writer, …)` variant exists solely for error-path testing with `errWriter`.

### template_values footgun

The Paubox API requires `template_values` to be a **JSON-encoded string**, not an object:
```json
{"data": {"template_values": "{\"name\":\"Alice\"}"}}
```
`SendTemplatedMessage` accepts `map[string]any` from callers and marshals it to a string internally. This is hidden from callers entirely.

### MessageHeaders custom marshalling

`MessageHeaders` implements `MarshalJSON` to flatten `CustomHeaders` map entries into the top-level JSON object alongside the standard fields (`subject`, `from`, `reply-to`, etc.).

---

## Adding a new endpoint

### Authenticated (Email API)
1. Fetch the live schema from `https://docs.paubox.com/api-reference/` — don't guess from naming patterns.
2. Add public request/response types to `*_types.go`; keep unexported wire types separate.
3. Implement the method following the pattern in `messages.go` or `templates.go`: validate first, then call `doJSON` or `doMultipartTemplate`.
4. Validation errors must be prefixed `"paubox: MethodName: "`.
5. Tests (in the same package, not `_test`): happy path, correct method+path, request body assertions, all validation cases, ≥400/401/404 error responses via `httptest.Server`.
6. Update `README.md` (add a `<details>` usage block), `api.md`, and `CHANGELOG.md`.

### Unauthenticated (Forms API)
Follow the same steps above but add methods to `FormsClient` in `forms.go`/`forms_types.go`. Use `c.doFormsJSON()` instead of `c.doJSON()`. Tests go in `forms_test.go`; use `newTestFormsClient(t, srv)`. Add a no-auth-header assertion for every new Forms method.

---

## Testing conventions

All tests use `httptest.NewServer`/`httptest.NewTLSServer` — no live API calls. Test files are in `package paubox` (same package) so unexported helpers like `validateMessage` are accessible. Standard helpers: `newTestClient(t, srv)` in `client_test.go`, `respondJSON(w, code, body)` in `messages_test.go`. Unused `http.HandlerFunc` parameters must be named `_` to satisfy `revive`.

---

## Security constraints

- The API key lives only in `c.apiKey`; never log it or include it in errors.
- `FormsClient` has no `apiKey` field — it must never send an `Authorization` header. Every new Forms method needs a test asserting `Authorization` is empty.
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

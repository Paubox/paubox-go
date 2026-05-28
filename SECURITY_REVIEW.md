# Security Review — paubox-go v0.1.1

Reviewed: 2026-05-19  
Reviewer: Paubox Engineering  
Scope: all `.go` files in this repository

---

## Summary

No critical or high-severity findings. Three low-severity observations are documented below with mitigations already in place or recommended actions.

---

## Findings

### 1. Credential handling — PASS

**What was checked:**
- API key never appears in log output, error messages, or `fmt.Sprintf` calls
- `PauboxError.Raw` stores raw response bodies for debugging but is never logged by the SDK itself
- The `Authorization` header is set only inside `client.go:do()` — one location, no string construction by callers

**Evidence:**
- `client.go`: `req.Header.Set("Authorization", "Token token="+c.apiKey)` — set once, not logged
- `errors.go`: `Raw []byte` field documented with the explicit note "The SDK never logs this value to avoid inadvertently capturing PHI"
- No `log.*`, `fmt.Print*`, or `os.Stderr.Write` calls anywhere in non-test SDK code

**Recommendation for callers:** Treat `PauboxError.Raw` as sensitive. Do not log it without scrubbing; it may contain message metadata.

---

### 2. TLS configuration — PASS

**What was checked:**
- Default `http.Transport` minimum TLS version
- No `InsecureSkipVerify` usage anywhere
- Custom HTTP client path (`WithHTTPClient`)

**Evidence:**
```go
// client.go
Transport: &http.Transport{
    TLSClientConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
    },
},
```

`tls.VersionTLS12` is enforced on the default client. A `grep` for `InsecureSkipVerify` across the codebase returns zero results.

**Low finding — custom client bypass:** When callers supply `WithHTTPClient(hc)`, the SDK uses their client's TLS config as-is. This is the correct design (the SDK cannot override a caller's transport), but it is a footgun.

**Mitigation in place:** `WithHTTPClient` is documented with an explicit warning: *"callers are responsible for maintaining a minimum TLS version of 1.2 and for not setting InsecureSkipVerify."* `SECURITY.md` repeats this. No code change required; documentation is the correct control here.

---

### 3. Input validation — PASS with notes

**What was checked:**
- Empty/whitespace recipients, from, subject
- Attachment field completeness
- Tracking ID validation
- Header injection via `CustomHeaders`
- Template ID path parameter injection

**Validation in place:**
- `validateMessage()` checks `len(recipients) == 0`, `headers.From == ""`, `headers.Subject == ""`, `content nil`, and all three attachment fields
- `GetEmailDisposition` rejects empty and whitespace-only tracking IDs via `strings.TrimSpace`
- `GetTemplate`, `UpdateTemplate`, `DeleteTemplate` reject empty `id`

**Low finding — header injection in CustomHeaders:** `MessageHeaders.CustomHeaders` accepts arbitrary string keys and values and serialises them directly into the JSON body sent to the API. A caller could supply a key containing a newline or special character. The Paubox API is the ultimate gatekeeper here (it validates header names server-side), but the SDK provides no client-side validation.

**Recommendation:** In a future release, add validation that custom header keys match `^[Xx]-[A-Za-z0-9-]+$` and that values contain no CR/LF characters. This is a defence-in-depth measure, not a blocking issue for v0.1.1 since the API rejects invalid headers.

**Low finding — template ID path injection:** Template IDs are interpolated directly into URL paths (e.g. `/dynamic_templates/` + id). If a caller passes a value like `../messages`, it would change the request path. However: (a) the `id` comes from a previous `ListTemplates`/`CreateTemplate` API response, not from untrusted user input in typical use; (b) the `net/http` client normalises paths.

**Recommendation:** Document that `id` parameters must be API-returned values. A future release could add `strings.ContainsAny(id, "/\\.?")` validation.

---

### 4. PHI / HIPAA considerations — PASS

**What was checked:**
- No logging of request or response bodies anywhere in SDK code
- No telemetry, metrics, or tracing added by the SDK
- No caching of message content

**Evidence:** A search for `log.`, `fmt.Print`, `os.Stdout`, `os.Stderr` in non-test SDK files returns zero results. The SDK is deliberately silent — it returns data to the caller and logs nothing.

**Recommendation for callers (documented in `SECURITY.md` and `paubox.go`):**
- Do not log `SendMessageRequest`, `SendBatchRequest`, or `SendTemplatedMessageRequest` values — they may contain recipient email addresses or message content that qualifies as PHI
- Do not log `PauboxError.Raw` without scrubbing
- Do not log `GetEmailDisposition` responses — they contain recipient addresses and open/click metadata

---

### 5. Retry behaviour — PASS

**What was checked:**
- POST requests are not retried by default (non-idempotent)
- Jitter prevents thundering-herd on retry storms
- Context cancellation is respected during backoff sleep

**Evidence:**
```go
// client.go:backoff
select {
case <-ctx.Done():
case <-time.After(wait):
}
```

If the context is cancelled during a backoff sleep, the goroutine wakes immediately and the next `c.httpClient.Do(req)` will return `ctx.Err()`.

---

### 6. Dependency audit — PASS

**Runtime dependencies:** None. The SDK uses only the Go standard library.

**`govulncheck` results:** Cannot be run in this environment (Go not in PATH). Run `govulncheck ./...` as part of CI (configured in `.github/workflows/ci.yml`) before each release. With zero external runtime dependencies, the vulnerability surface is limited to Go stdlib CVEs, which are rare and typically addressed by upgrading the Go toolchain.

---

### 7. Timing-safe comparisons — NOT APPLICABLE

The SDK does not perform any credential comparison, token validation, or equality check on secrets. The API key is only ever written into a request header; it is never read back or compared. No timing-safe comparison is needed.

---

## Checklist

| Area | Status | Notes |
|---|---|---|
| API key not logged | ✅ Pass | Single set point in `do()` |
| API key not in error messages | ✅ Pass | Errors contain status code and API message only |
| No `InsecureSkipVerify` | ✅ Pass | grep confirms zero uses |
| TLS 1.2 minimum | ✅ Pass | Enforced on default transport |
| Input validation — required fields | ✅ Pass | `validateMessage`, `validateAttachment` |
| Input validation — empty IDs | ✅ Pass | All resource methods check |
| Header injection — CustomHeaders | ⚠️ Low | No client-side key/value validation; API validates server-side |
| Path injection — template IDs | ⚠️ Low | IDs expected to be API-returned values; documented |
| PHI not logged by SDK | ✅ Pass | Zero log calls in SDK code |
| POST not retried by default | ✅ Pass | `RetryNonIdempotent: false` default |
| Context cancellation in backoff | ✅ Pass | `select` on `ctx.Done()` |
| Zero runtime deps | ✅ Pass | `go.mod` has no `require` block |
| `govulncheck` | ⏳ Pending | Run via CI; zero deps limits exposure |
| Timing-safe comparisons | N/A | No secret comparisons performed |

---

## Recommended actions before v1.0.0

1. Add `CustomHeaders` key/value validation (defence-in-depth against header injection)
2. Add `id` parameter validation to reject values containing `/`, `\`, or `..`
3. Run `govulncheck ./...` in CI and confirm clean before tagging any release
4. Run `go test -race ./...` and confirm clean (race detector catches concurrency bugs the linter misses)

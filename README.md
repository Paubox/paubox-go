# paubox-go

The official Go SDK for the [Paubox](https://www.paubox.com) Email and Forms APIs — HIPAA-compliant transactional email and form collection for healthcare developers.

[![CI](https://github.com/paubox/paubox-go/actions/workflows/ci.yml/badge.svg)](https://github.com/paubox/paubox-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/paubox/paubox-go.svg)](https://pkg.go.dev/github.com/paubox/paubox-go)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

---

## Install

```bash
go get github.com/paubox/paubox-go
```

Requires Go 1.23 or later. No external runtime dependencies.

---

## Quickstart

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/paubox/paubox-go"
)

func main() {
    client, err := paubox.New("YOUR_API_KEY")
    if err != nil {
        log.Fatal(err)
    }

    resp, err := client.SendMessage(context.Background(), &paubox.SendMessageRequest{
        Message: paubox.Message{
            Recipients: []string{"recipient@example.com"},
            Headers: paubox.MessageHeaders{
                From:    "sender@yourdomain.com",
                Subject: "Hello from Paubox",
            },
            Content: paubox.MessageContent{
                PlainText: paubox.Ptr("Hello, world!"),
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Sent! Tracking ID:", resp.SourceTrackingID)
}
```

Find your API key in the [Paubox dashboard](https://app.paubox.com).

---

## Authentication

The Paubox Email API uses a non-standard authorization header:

```
Authorization: Token token=<API_KEY>
```

The SDK sets this header automatically on every request. **Never construct it manually** — use `paubox.New` with your key and the SDK handles it correctly every time. (The [Paubox Forms API](#paubox-forms) uses a different scheme — scoped keys with `Bearer` auth — see below.)

Store your API key in an environment variable, not in source code:

```go
client, err := paubox.New(os.Getenv("PAUBOX_API_KEY"))
```

---

## Usage

<details>
<summary><strong>Send a single message</strong></summary>

```go
resp, err := client.SendMessage(ctx, &paubox.SendMessageRequest{
    Message: paubox.Message{
        Recipients: []string{"alice@example.com"},
        CC:         []string{"manager@example.com"},
        Headers: paubox.MessageHeaders{
            From:    "noreply@yourdomain.com",
            Subject: "Your results are ready",
            ReplyTo: "support@yourdomain.com",
        },
        Content: paubox.MessageContent{
            PlainText: paubox.Ptr("Your results are attached."),
            HTML:      paubox.Ptr("<p>Your results are <strong>attached</strong>.</p>"),
        },
        Attachments: []paubox.Attachment{
            {
                FileName:    "results.pdf",
                ContentType: "application/pdf",
                Content:     base64EncodedPDF, // base64-encoded string
            },
        },
    },
    OverrideOpenTracking: true,
})
```
</details>

<details>
<summary><strong>Send a batch of messages</strong></summary>

Paubox recommends batches of 50 or fewer. Responses are returned in the same order as the request.

```go
messages := []paubox.Message{
    {
        Recipients: []string{"alice@example.com"},
        Headers:    paubox.MessageHeaders{From: "f@yourdomain.com", Subject: "Hi Alice"},
        Content:    paubox.MessageContent{PlainText: paubox.Ptr("Hello Alice")},
    },
    {
        Recipients: []string{"bob@example.com"},
        Headers:    paubox.MessageHeaders{From: "f@yourdomain.com", Subject: "Hi Bob"},
        Content:    paubox.MessageContent{PlainText: paubox.Ptr("Hello Bob")},
    },
}

resp, err := client.SendBatch(ctx, &paubox.SendBatchRequest{Messages: messages})
for i, msg := range resp.Messages {
    fmt.Printf("[%d] tracking ID: %s\n", i, msg.SourceTrackingID)
}
```
</details>

<details>
<summary><strong>Get email disposition (delivery status)</strong></summary>

```go
disp, err := client.GetEmailDisposition(ctx, resp.SourceTrackingID)
if err != nil {
    log.Fatal(err)
}

for _, d := range disp.Data.Message.MessageDeliveries {
    fmt.Printf("%s → %s\n", d.Recipient, d.Status.DeliveryStatus)
}

// Possible DeliveryStatus values:
// paubox.DeliveryStatusProcessing
// paubox.DeliveryStatusDelivered
// paubox.DeliveryStatusDeliveredViaPortal
// paubox.DeliveryStatusSoftBounced
// paubox.DeliveryStatusHardBounced
// paubox.DeliveryStatusTLSNotOffered
```

For production workloads, prefer [Paubox webhooks](https://docs.paubox.com/api-reference/) over polling.
</details>

<details>
<summary><strong>Dynamic templates — CRUD</strong></summary>

Templates use [Handlebars](https://handlebarsjs.com/) syntax.

```go
// Create — returns a confirmation message, NOT the new template's ID.
created, err := client.CreateTemplate(ctx, &paubox.CreateTemplateRequest{
    Name: "appointment-confirmation",
    Body: []byte(`<p>Hello {{first_name}}, your appointment is on {{date}}.</p>`),
})
fmt.Println(created.Message) // "Template appointment-confirmation created!"

// List — also used to resolve a template's numeric ID by name.
list, err := client.ListTemplates(ctx)
var id int64
for _, t := range list.Templates {
    if t.Name == "appointment-confirmation" {
        id = t.ID // template IDs are int64
        break
    }
}

// Get
tmpl, err := client.GetTemplate(ctx, id)

// Update — supply only the fields to change
_, err = client.UpdateTemplate(ctx, id, &paubox.UpdateTemplateRequest{
    Name: "appointment-confirmation-v2",
})

// Delete
_, err = client.DeleteTemplate(ctx, id)
```
</details>

<details>
<summary><strong>Send a templated message</strong></summary>

Pass `TemplateValues` as a plain Go map — the SDK serialises it correctly. Do not pre-encode it as JSON.

```go
resp, err := client.SendTemplatedMessage(ctx, &paubox.SendTemplatedMessageRequest{
    TemplateName: "appointment-confirmation",
    TemplateValues: map[string]any{
        "first_name": "Jane",
        "date":       "2024-03-15",
        "time":       "2:00 PM",
    },
    Message: paubox.TemplatedMessage{
        Recipients: []string{"jane@example.com"},
        Headers: paubox.MessageHeaders{
            From:    "appointments@yourclinic.com",
            Subject: "Your appointment is confirmed",
        },
    },
})
```
</details>

---

## Configuration

```go
client, err := paubox.New(apiKey,
    // Override the base URL (useful for staging or tests).
    paubox.WithBaseURL("https://api.paubox.com/v1"),

    // Per-request timeout.
    paubox.WithTimeout(15*time.Second),

    // Provide a custom HTTP client (you own its TLS configuration).
    paubox.WithHTTPClient(myHTTPClient),

    // Retry behaviour.
    // Default: GET requests retry up to 3× on 429/5xx with backoff + jitter.
    // POST/PATCH are not retried unless RetryNonIdempotent is true.
    paubox.WithRetry(paubox.RetryConfig{
        MaxAttempts:        4,
        WaitMin:            200 * time.Millisecond,
        WaitMax:            5 * time.Second,
        RetryNonIdempotent: false,
    }),

    // Prepend a custom string to the User-Agent header.
    paubox.WithUserAgent("myapp/1.0"),
)
```

### Ptr helper

Use `paubox.Ptr` to set optional pointer-typed fields without declaring a named variable:

```go
Content: paubox.MessageContent{
    PlainText: paubox.Ptr("Hello"),
    HTML:      paubox.Ptr("<p>Hello</p>"),
}
```

---

## Paubox Forms

The SDK also covers the Paubox Forms API — HIPAA-compliant form hosting and submission collection. Forms use a dedicated client, created with `paubox.NewForms`.

### Forms authentication

Protected Forms endpoints authenticate with a **scoped API key** carrying the `forms` scope, sent as a standard Bearer header:

```
Authorization: Bearer <SCOPED_API_KEY>
```

This is **not** the Email API's `Token token=` format — Email keys and Forms scoped keys are not interchangeable. As with the Email client, the SDK sets the header automatically; never construct it yourself.

```go
forms, err := paubox.NewForms(os.Getenv("PAUBOX_FORMS_API_KEY"))
```

Pass an **empty key** to get a public-only client. Only the two unauthenticated endpoints — `GetPublicForm` and `SubmitForm` — work on it; every other method fails fast with an error before any request is sent. This is the right client to embed in a public-facing app that only renders and submits forms:

```go
public, err := paubox.NewForms("")
```

### Forms base URL

The default base URL is `https://api.paubox.com/v1/forms`. The SDK appends the Forms service routes verbatim (e.g. `/api/forms`, `/public/form_data/{id}`) — the default assumes the production gateway forwards paths under `https://api.paubox.com/v1/forms` unchanged. Use `WithFormsBaseURL` to point at a different mount, e.g. a staging gateway or a locally running Forms service:

```go
forms, err := paubox.NewForms(apiKey,
    paubox.WithFormsBaseURL("http://localhost:3000"),
)
```

### Forms usage

<details>
<summary><strong>Forms — list, create, get, update</strong></summary>

```go
// List — params may be nil for the server defaults
// (page 1, 50 items, ordered by created_at descending).
list, err := forms.ListForms(ctx, &paubox.ListFormsParams{
    Search: "intake",
    Active: paubox.Ptr(true),
    Page:   1,
    Items:  50,
})
for _, f := range list.Results {
    fmt.Println(f.ID, f.Title, f.SubmissionCount)
}
fmt.Println("total:", list.PageInfo.Count)

// Create — returns the new form's UUID.
// Version defaults to 1 when left zero.
created, err := forms.CreateForm(ctx, &paubox.CreateFormRequest{
    Title:      "Patient intake",
    CustomerID: 1234,
    FormJSON:   json.RawMessage(`{"fields":[{"name":"first_name","type":"text"}]}`),
    Recipient:  "intake@yourclinic.com", // comma-separated notification recipients
    Active:     true,
})
fmt.Println("new form ID:", created.ID)

// Get
form, err := forms.GetForm(ctx, created.ID)

// Update — PATCH-style: only non-nil fields are sent;
// everything else is left unchanged on the server.
updated, err := forms.UpdateForm(ctx, created.ID, &paubox.UpdateFormRequest{
    Title:  paubox.Ptr("Patient intake v2"),
    Active: paubox.Ptr(false),
})
fmt.Println(updated.Detail)
```
</details>

<details>
<summary><strong>Forms — archive, unarchive, copy, stats</strong></summary>

```go
// Archive — also deactivates the form, so it stops accepting submissions.
_, err := forms.ArchiveForm(ctx, formID)

// Unarchive — does NOT reactivate the form; toggle Active via UpdateForm.
_, err = forms.UnarchiveForm(ctx, formID)

// Copy — duplicates a form under a new title and returns the new Form.
dup, err := forms.CopyForm(ctx, &paubox.CopyFormRequest{
    FormID: formID,
    Title:  "Patient intake (copy)",
})
fmt.Println("copy ID:", dup.ID)

// Stats — nil params scopes the stats to the API key's customer.
stats, err := forms.GetFormStats(ctx, nil)
fmt.Println(stats.ActiveFormCount, stats.TotalSubmissionCount, stats.SubmissionsLast7Days)
```
</details>

<details>
<summary><strong>Public endpoints — fetch and submit a form (no API key)</strong></summary>

These two endpoints work on any Forms client, including a keyless one from `paubox.NewForms("")`.

```go
public, err := paubox.NewForms("")

// Fetch the public definition of an active form.
// Returns 404 (paubox.ErrNotFound) when the form is inactive, archived, or deleted.
form, err := public.GetPublicForm(ctx, formID)

// Submit a filled-out form. FormData keys are the form's slugified field
// names, sent as a real JSON object. Attachment contents are raw bytes —
// the SDK base64-encodes them on the wire. Success returns nil.
err = public.SubmitForm(ctx, formID, &paubox.SubmitFormRequest{
    FormData: map[string]any{
        "first_name": "Jane",
        "email":      "jane@example.com",
    },
    Attachments: []paubox.FormSubmissionAttachment{
        {Name: "insurance-card.pdf", Content: pdfBytes},
    },
})
```
</details>

<details>
<summary><strong>Submissions — list and export (CSV / PDF)</strong></summary>

```go
// List a form's submissions — params may be nil for the server defaults.
subs, err := forms.ListFormSubmissions(ctx, formID, &paubox.ListFormSubmissionsParams{
    Order: "desc",
    Page:  1,
})
for _, s := range subs.Data {
    // Note: FormData arrives as a JSON-encoded string, not an object —
    // unmarshal it yourself if you need the individual fields.
    fmt.Println(s.ID, s.CreatedAt, s.FormData)
}

// Exports return the raw file bytes. They may contain PHI —
// handle and store them accordingly.
csvAll, err := forms.ExportFormSubmissionsCSV(ctx, formID)               // all submissions
csvOne, err := forms.ExportFormSubmissionCSV(ctx, formID, submissionID)  // one submission
pdfOne, err := forms.ExportFormSubmissionPDF(ctx, formID, submissionID)  // one submission
```
</details>

### Forms configuration

The Forms client mirrors the Email client's options and defaults (30 s timeout, TLS 1.2 minimum, same retry policy):

```go
forms, err := paubox.NewForms(apiKey,
    // Override the base URL (staging, tests, or a local Forms service).
    paubox.WithFormsBaseURL("http://localhost:3000"),

    // Per-request timeout.
    paubox.WithFormsTimeout(15*time.Second),

    // Provide a custom HTTP client (you own its TLS configuration).
    paubox.WithFormsHTTPClient(myHTTPClient),

    // Retry behaviour. GET requests retry up to 3× on 429/5xx with backoff
    // + jitter. POST/PUT are not retried unless RetryNonIdempotent is true
    // (the Forms service uses PUT for updates).
    paubox.WithFormsRetry(paubox.RetryConfig{
        MaxAttempts: 4,
        WaitMin:     200 * time.Millisecond,
        WaitMax:     5 * time.Second,
    }),

    // Prepend a custom string to the User-Agent header.
    paubox.WithFormsUserAgent("myapp/1.0"),
)
```

Forms API errors are returned as the same `*paubox.PauboxError` used by the Email client, so the `errors.Is` sentinels and `errors.As` patterns in the next section apply unchanged.

---

## Error handling

All API errors are returned as `*paubox.PauboxError`. Use `errors.As` to inspect the full error and `errors.Is` to match against status-code sentinels:

```go
resp, err := client.SendMessage(ctx, req)
if err != nil {
    if errors.Is(err, paubox.ErrUnauthorized) {
        log.Fatal("check your API key")
    }
    if errors.Is(err, paubox.ErrRateLimit) {
        log.Fatal("rate limited — back off and retry")
    }

    var apiErr *paubox.PauboxError
    if errors.As(err, &apiErr) {
        fmt.Printf("HTTP %d: %s — %s (request ID: %s)\n",
            apiErr.StatusCode, apiErr.Title, apiErr.Details, apiErr.RequestID)
    }
    log.Fatal(err)
}
```

**Sentinels:**

| Sentinel | HTTP status |
|---|---|
| `paubox.ErrBadRequest` | 400 |
| `paubox.ErrUnauthorized` | 401 |
| `paubox.ErrForbidden` | 403 |
| `paubox.ErrNotFound` | 404 |
| `paubox.ErrRateLimit` | 429 |
| `paubox.ErrServerError` | 500 |

Include `apiErr.RequestID` in any support request to Paubox.

---

## HIPAA / compliance

Paubox is a HIPAA-compliant email platform. This SDK is designed for use in regulated healthcare environments:

- **The SDK never logs request bodies, response bodies, or API credentials.** It is deliberately silent.
- **Do not log** `SendMessageRequest`, response objects, or `PauboxError.Raw` in your application without scrubbing — these values may contain Protected Health Information (PHI).
- Form submissions (`FormSubmission`, `SubmitFormRequest`) and the CSV/PDF export bytes may also contain PHI — apply the same care when handling or storing them.
- `AllowNonTLS: true` on a `Message` permits delivery without TLS encryption. Consult your compliance team before enabling this.
- For a full security analysis of the SDK itself, see [SECURITY_REVIEW.md](SECURITY_REVIEW.md).

---

## Versioning & compatibility

This SDK follows [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html). As of **v1.0.0** the public API is stable.

**Covered by the compatibility promise** — the exported identifiers of the `paubox` package: `Client`, `FormsClient`, their methods, the request and response types, the functional options, `PauboxError` and its sentinels, and `Ptr`.

**Not covered** — anything under `internal/`, the programs under `examples/`, the exact wording of error strings, and the `User-Agent` value. These may change in any release.

Within 1.x, your code will keep compiling: new endpoints and new struct fields arrive as minor releases, bug fixes as patch releases. Because Go encodes the major version in the module path, a future 2.0.0 would be published as a *separate* module at `github.com/paubox/paubox-go/v2` — your existing imports keep resolving to 1.x until you deliberately migrate.

Pin the major version and take minor upgrades freely:

```bash
go get github.com/paubox/paubox-go@latest
```

The running version is available at `paubox.Version`, and each release is described in [CHANGELOG.md](CHANGELOG.md).

> **Upgrading from 0.x?** The 0.x line made no stability guarantees and did contain breaking changes. See [Migrating from 0.2.0](CHANGELOG.md#migrating-from-020) for the two changes you need to make.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Apache 2.0 — see [LICENSE](LICENSE).

## Support

- **API documentation:** https://docs.paubox.com/api-reference/
- **Security vulnerabilities:** security@paubox.com — see [SECURITY.md](SECURITY.md)
- **General support:** https://www.paubox.com/contact
## 💬 Community & support

Questions, ideas, or want to share what you built? Join the **[Paubox Community](https://github.com/Paubox/community/discussions)** — the single home for discussions across every Paubox SDK and API.

🔐 Found a security issue? Email **devops@paubox.com** — please don't post it publicly.

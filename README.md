# paubox-go

The official Go SDK for the [Paubox](https://www.paubox.com) Email API — HIPAA-compliant transactional email for healthcare developers.

[![CI](https://github.com/paubox/paubox-go/actions/workflows/ci.yml/badge.svg)](https://github.com/paubox/paubox-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/paubox/paubox-go.svg)](https://pkg.go.dev/github.com/paubox/paubox-go)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

---

## Install

```bash
go get github.com/paubox/paubox-go
```

Requires Go 1.22 or later. No external runtime dependencies.

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
    client, err := paubox.New("YOUR_API_KEY", "YOUR_API_USERNAME")
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

Find your API key and username in the [Paubox dashboard](https://app.paubox.com).

---

## Authentication

The Paubox API uses a non-standard authorization header:

```
Authorization: Token token=<API_KEY>
```

The SDK sets this header automatically on every request. **Never construct it manually** — use `paubox.New` with your key and the SDK handles it correctly every time.

Store your API key in an environment variable, not in source code:

```go
client, err := paubox.New(os.Getenv("PAUBOX_API_KEY"), os.Getenv("PAUBOX_USERNAME"))
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
// Create
tmpl, err := client.CreateTemplate(ctx, &paubox.CreateTemplateRequest{
    Name: "appointment-confirmation",
    Body: []byte(`<p>Hello {{first_name}}, your appointment is on {{date}}.</p>`),
})

// List
list, err := client.ListTemplates(ctx)

// Get
tmpl, err := client.GetTemplate(ctx, "template-id")

// Update — supply only the fields to change
tmpl, err := client.UpdateTemplate(ctx, "template-id", &paubox.UpdateTemplateRequest{
    Name: "appointment-confirmation-v2",
})

// Delete
_, err = client.DeleteTemplate(ctx, "template-id")
```
</details>

<details>
<summary><strong>Paubox Forms — get schema and submit a response (no credentials required)</strong></summary>

The `FormsClient` is separate from the Email API client and requires no API key.

```go
fc, err := paubox.NewFormsClient()
if err != nil {
    log.Fatal(err)
}

// Retrieve the form's metadata and field schema.
form, err := fc.GetForm(ctx, "your-form-uuid")
if err != nil {
    log.Fatal(err)
}
fmt.Println("Form:", form.Title)
for _, field := range form.FormJSON.Body {
    fmt.Printf("  [%s] %s\n", field.Type, field.ID)
}

// Submit a response. Keys in FormData should match the form's field names.
_, err = fc.SubmitForm(ctx, "your-form-uuid", paubox.FormSubmission{
    FormData: map[string]any{
        "name":  "Alice Example",
        "email": "alice@example.com",
    },
    // Optional: attach a file.
    Attachments: []paubox.FormAttachment{
        {Name: "consent.pdf", Content: base64EncodedPDF},
    },
})
if err != nil {
    log.Fatal(err)
}
```

`FormsClient` supports the same options as `Client` (prefixed with `WithForms`): `WithFormsBaseURL`, `WithFormsHTTPClient`, `WithFormsTimeout`, `WithFormsRetry`, `WithFormsUserAgent`.
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
client, err := paubox.New(apiKey, username,
    // Override the base URL (useful for staging or tests).
    paubox.WithBaseURL("https://api.paubox.net/v1"),

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
- `AllowNonTLS: true` on a `Message` permits delivery without TLS encryption. Consult your compliance team before enabling this.
- For a full security analysis of the SDK itself, see [SECURITY_REVIEW.md](SECURITY_REVIEW.md).

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). All contributions require a DCO sign-off (`git commit -s`).

## License

Apache 2.0 — see [LICENSE](LICENSE).

## Support

- **API documentation:** https://docs.paubox.com/api-reference/
- **Security vulnerabilities:** security@paubox.com — see [SECURITY.md](SECURITY.md)
- **General support:** https://www.paubox.com/contact

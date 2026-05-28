# API Reference — paubox-go

Full reference for the `github.com/paubox/paubox-go` SDK.

---

## Client

### `New`

```go
func New(apiKey, username string, opts ...Option) (*Client, error)
```

Creates a new client. Both `apiKey` and `username` are required and must be non-empty. Returns an error if either is blank or whitespace-only.

```go
client, err := paubox.New("your-api-key", "your-username")
```

### Options

| Option | Description |
|---|---|
| `WithBaseURL(url string)` | Override the API base URL (no trailing slash). Useful for staging environments or tests. |
| `WithHTTPClient(hc *http.Client)` | Replace the default HTTP client. Caller is responsible for TLS ≥ 1.2 and not setting `InsecureSkipVerify`. |
| `WithTimeout(d time.Duration)` | Set the per-request timeout on the default HTTP client. Ignored if `WithHTTPClient` is also used. |
| `WithRetry(cfg RetryConfig)` | Configure retry behaviour. Pass a zero value (`RetryConfig{}`) to disable retries. |
| `WithUserAgent(ua string)` | Prepend a custom token to the `User-Agent` header. The SDK identifier is always appended. |

### `RetryConfig`

```go
type RetryConfig struct {
    MaxAttempts        int           // total attempts (1 = no retries)
    WaitMin            time.Duration // minimum backoff
    WaitMax            time.Duration // maximum backoff (cap)
    RetryNonIdempotent bool          // enable retries for POST and PATCH
}
```

Default: 3 attempts, 100 ms–2 s exponential backoff with jitter, GET/DELETE only. The `Retry-After` response header is honoured when present.

---

## Messages

### `SendMessage`

```go
func (c *Client) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error)
```

Sends a single HIPAA-compliant email. `POST /messages`

**`SendMessageRequest`**

| Field | Type | Required | Description |
|---|---|---|---|
| `Message` | `Message` | ✅ | The email to send. |
| `OverrideOpenTracking` | `bool` | | Enable open tracking for this message. |
| `OverrideLinkTracking` | `bool` | | Enable click tracking for this message. |
| `UnsubscribeURL` | `string` | | URL to redirect unsubscribe requests to. |

**`Message`**

| Field | Type | Required | Description |
|---|---|---|---|
| `Recipients` | `[]string` | ✅ | To: addresses. Plain address or `"Name <addr>"` format. |
| `Headers` | `MessageHeaders` | ✅ | See below. |
| `Content` | `MessageContent` | ✅ | See below. At least one of `PlainText` or `HTML` required. |
| `BCC` | `[]string` | | Blind carbon-copy recipients. |
| `CC` | `[]string` | | Carbon-copy recipients. |
| `AllowNonTLS` | `bool` | | Permit delivery without TLS. May affect HIPAA compliance. |
| `ForceSecureNotification` | `bool` | | Force delivery via the Paubox Secure Message portal. |
| `Attachments` | `[]Attachment` | | File attachments. |

**`MessageHeaders`**

| Field | Type | Required | Description |
|---|---|---|---|
| `From` | `string` | ✅ | Sender address. Must be a verified domain in your account. |
| `Subject` | `string` | ✅ | Email subject line. |
| `ReplyTo` | `string` | | Reply-To address. |
| `ListUnsubscribe` | `string` | | `List-Unsubscribe` header (RFC 2369). |
| `ListUnsubscribePost` | `string` | | `List-Unsubscribe-Post` header (RFC 8058). |
| `CustomHeaders` | `map[string]string` | | Additional headers. Keys must begin with `x-` or `X-`. Serialised as top-level fields alongside standard headers. |

**`MessageContent`**

| Field | Type | Description |
|---|---|---|
| `PlainText` | `*string` | `text/plain` body. Use `paubox.Ptr("…")` to set. |
| `HTML` | `*string` | `text/html` body. Plain HTML or base64-encoded. |

**`Attachment`**

| Field | Type | Required | Description |
|---|---|---|---|
| `FileName` | `string` | ✅ | Filename shown to the recipient (e.g. `"report.pdf"`). |
| `ContentType` | `string` | ✅ | MIME type (e.g. `"application/pdf"`). |
| `Content` | `string` | ✅ | Base64-encoded file data. |

**`SendMessageResponse`**

| Field | Type | Description |
|---|---|---|
| `SourceTrackingID` | `string` | Tracking ID for this message. Pass to `GetEmailDisposition`. |
| `Data` | `string` | Service status message (typically `"Service OK"`). |
| `CustomHeaders` | `map[string]string` | Custom headers accepted by the API. |

---

### `SendBatch`

```go
func (c *Client) SendBatch(ctx context.Context, req *SendBatchRequest) (*SendBatchResponse, error)
```

Sends multiple emails in a single API call. Paubox recommends batches of 50 or fewer. Each message is validated individually before the request is sent. `POST /bulk_messages`

**`SendBatchRequest`**

| Field | Type | Required | Description |
|---|---|---|---|
| `Messages` | `[]Message` | ✅ | Emails to send. At least one required. |

**`SendBatchResponse`**

| Field | Type | Description |
|---|---|---|
| `Messages` | `[]SendMessageResponse` | One entry per request message, in the same order. |

---

### `GetEmailDisposition`

```go
func (c *Client) GetEmailDisposition(ctx context.Context, sourceTrackingID string) (*EmailDisposition, error)
```

Retrieves delivery status and engagement metrics for a previously sent message. `GET /message_receipt`

**`EmailDisposition`**

| Field | Type | Description |
|---|---|---|
| `SourceTrackingID` | `string` | Echoes the queried tracking ID. |
| `Data.Message` | `MessageDisposition` | Per-recipient delivery detail and aggregate metrics. |

**`MessageDisposition`**

| Field | Type | Description |
|---|---|---|
| `ID` | `string` | Internal Paubox message identifier. |
| `MessageDeliveries` | `[]MessageDelivery` | One entry per recipient. |
| `TotalOpens` | `*int` | Total open events across all recipients. |
| `DistinctOpens` | `*int` | Unique recipients who opened the message. |
| `TotalClickCount` | `*int` | Aggregate link clicks. |
| `ClicksPerLink` | `[]LinkClick` | Click counts broken down by URL. |
| `Unsubscribed` | `*bool` | Whether any recipient unsubscribed. |

**`MessageDelivery`**

| Field | Type | Description |
|---|---|---|
| `Recipient` | `string` | Email address of the recipient. |
| `Status.DeliveryStatus` | `string` | See delivery status constants below. |
| `Status.DeliveryTime` | `*string` | RFC 2822 timestamp of delivery. |
| `Status.OpenedStatus` | `*string` | `"opened"` or `"not opened"`. |
| `Status.OpenedTime` | `*string` | Timestamp of the first open event. |

**Delivery status constants**

| Constant | Value |
|---|---|
| `DeliveryStatusProcessing` | `"processing"` |
| `DeliveryStatusDelivered` | `"delivered"` |
| `DeliveryStatusDeliveredViaPortal` | `"delivered via secure portal"` |
| `DeliveryStatusTLSNotOffered` | `"TLS not offered, sending via Secure Portal"` |
| `DeliveryStatusSoftBounced` | `"soft bounced"` |
| `DeliveryStatusSoftBouncedMailboxFull` | `"soft bounced - mailbox full"` |
| `DeliveryStatusHardBounced` | `"hard bounced"` |
| `DeliveryStatusInternalError` | `"Internal error. Please check back later."` |

**Open status constants**

| Constant | Value |
|---|---|
| `OpenedStatusOpened` | `"opened"` |
| `OpenedStatusNotOpened` | `"not opened"` |

---

## Dynamic Templates

Template bodies use [Handlebars](https://handlebarsjs.com/) syntax: `{{variableName}}`.

### `ListTemplates`

```go
func (c *Client) ListTemplates(ctx context.Context) (*ListTemplatesResponse, error)
```

Returns all dynamic templates in your account. `GET /dynamic_templates`

**`ListTemplatesResponse`**

| Field | Type | Description |
|---|---|---|
| `Templates` | `[]Template` | All templates in the account. |

---

### `GetTemplate`

```go
func (c *Client) GetTemplate(ctx context.Context, id int64) (*Template, error)
```

Returns a single template by its ID. `GET /dynamic_templates/{id}`

---

### `CreateTemplate`

```go
func (c *Client) CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (*TemplateMutationResponse, error)
```

Uploads a new Handlebars template. Internally uses `multipart/form-data`. `POST /dynamic_templates`

> The API confirms creation with a message and **does not return the new template's ID**. To get it, call `ListTemplates` and match on `Name`.

**`CreateTemplateRequest`**

| Field | Type | Required | Description |
|---|---|---|---|
| `Name` | `string` | ✅ | Human-readable template name. |
| `Body` | `[]byte` | ✅ | Handlebars template content. |

**`TemplateMutationResponse`** (returned by `CreateTemplate` and `UpdateTemplate`)

| Field | Type | Description |
|---|---|---|
| `Message` | `string` | Confirmation message, e.g. `"Template welcome created!"`. |
| `Params.Name` | `string` | The template name the API recorded. |

---

### `UpdateTemplate`

```go
func (c *Client) UpdateTemplate(ctx context.Context, id int64, req *UpdateTemplateRequest) (*TemplateMutationResponse, error)
```

Modifies an existing template. Supply only the fields to change; omitted fields retain their current values. At least one field must be non-empty. Returns a confirmation message, not the updated template. `PATCH /dynamic_templates/{id}`

**`UpdateTemplateRequest`**

| Field | Type | Description |
|---|---|---|
| `Name` | `string` | New template name. Leave empty to keep current. |
| `Body` | `[]byte` | New Handlebars content. Leave nil to keep current. |

---

### `DeleteTemplate`

```go
func (c *Client) DeleteTemplate(ctx context.Context, id int64) (*DeleteTemplateResponse, error)
```

Permanently removes a template. `DELETE /dynamic_templates/{id}`

**`DeleteTemplateResponse`**

| Field | Type | Description |
|---|---|---|
| `Message` | `string` | Confirmation message from the API. |

---

### `SendTemplatedMessage`

```go
func (c *Client) SendTemplatedMessage(ctx context.Context, req *SendTemplatedMessageRequest) (*SendMessageResponse, error)
```

Sends an email rendered from a stored Handlebars template. `POST /templated_messages`

Returns a `*SendMessageResponse` (same as `SendMessage`).

**`SendTemplatedMessageRequest`**

| Field | Type | Required | Description |
|---|---|---|---|
| `TemplateName` | `string` | ✅ | Exact name of the template to use. |
| `TemplateValues` | `map[string]any` | | Variable values for Handlebars substitution. The SDK serialises this to the JSON-encoded string the API requires — do not pre-encode it. |
| `Message` | `TemplatedMessage` | ✅ | Routing and delivery options. |

**`TemplatedMessage`** — same fields as `Message` except `Content` is omitted (provided by the template).

---

**`Template`** (returned by all template methods)

| Field | Type | Description |
|---|---|---|
| `ID` | `int64` | Unique numeric template identifier. |
| `Name` | `string` | Human-readable name. |
| `APICustomerID` | `int64` | Account the template belongs to. |
| `Body` | `string` | Handlebars template content. Returned by `GetTemplate`; empty in `ListTemplates`. |
| `Metadata` | `map[string]any` | Arbitrary template metadata, when returned. |
| `CreatedAt` | `*time.Time` | Creation timestamp, when returned. |
| `UpdatedAt` | `*time.Time` | Last-modified timestamp, when returned. |

---

## Errors

All API errors are returned as `*PauboxError`. Use `errors.Is` with sentinels to match by HTTP status, or `errors.As` to access the full detail.

```go
var apiErr *paubox.PauboxError
if errors.As(err, &apiErr) {
    fmt.Printf("HTTP %d: %s — %s (request-id: %s)\n",
        apiErr.StatusCode, apiErr.Title, apiErr.Details, apiErr.RequestID)
}

if errors.Is(err, paubox.ErrUnauthorized) {
    // rotate API key
}
```

**`PauboxError` fields**

| Field | Type | Description |
|---|---|---|
| `StatusCode` | `int` | HTTP status code. |
| `Code` | `string` | Application-level error code from the API. |
| `Title` | `string` | Short error summary. |
| `Details` | `string` | Extended error description. |
| `RequestID` | `string` | `X-Request-Id` header value. Include when contacting support. |
| `Raw` | `[]byte` | Unmodified response body for debugging. Never logged by the SDK. |

**Sentinel errors**

| Sentinel | HTTP status |
|---|---|
| `ErrBadRequest` | 400 |
| `ErrUnauthorized` | 401 |
| `ErrForbidden` | 403 |
| `ErrNotFound` | 404 |
| `ErrRateLimit` | 429 |
| `ErrServerError` | 500 |

---

## Utilities

### `Ptr`

```go
func Ptr[T any](v T) *T
```

Returns a pointer to `v`. Convenience helper for setting optional pointer-typed fields inline:

```go
Content: paubox.MessageContent{
    PlainText: paubox.Ptr("Hello!"),
    HTML:      paubox.Ptr("<p>Hello!</p>"),
}
```

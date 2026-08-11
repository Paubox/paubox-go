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
    RetryNonIdempotent bool          // enable retries for POST, PUT, and PATCH
}
```

Default: 3 attempts, 100 ms–2 s exponential backoff with jitter, GET/DELETE only. The `Retry-After` response header is honoured when present. The same `RetryConfig` type configures both the Email client (`WithRetry`) and the Forms client (`WithFormsRetry`).

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

## Paubox Forms

The Forms API uses a dedicated client. Protected endpoints authenticate with a **scoped API key** carrying the `forms` scope, sent as `Authorization: Bearer <key>` — a different scheme from the Email API's `Token token=` header. Form and submission IDs are UUID strings.

### `NewForms`

```go
func NewForms(apiKey string, opts ...FormsOption) (*FormsClient, error)
```

Creates a new Forms client. `apiKey` may be empty (or whitespace-only, which is treated as empty): that yields a **public-only client** on which only `GetPublicForm` and `SubmitForm` work — every protected method fails fast with an error before any request is sent.

```go
forms, err := paubox.NewForms("your-scoped-api-key")
public, err := paubox.NewForms("") // public endpoints only
```

Defaults match the Email client: 30 s timeout, TLS 1.2 minimum, same retry policy.

### Forms options

| Option | Description |
|---|---|
| `WithFormsBaseURL(url string)` | Override the Forms base URL (trailing slash trimmed). Default `https://api.paubox.com/v1/forms` — service routes are appended verbatim, assuming the production gateway forwards the remaining path unchanged. Point at e.g. `http://localhost:3000` for a local Forms service. |
| `WithFormsHTTPClient(hc *http.Client)` | Replace the default HTTP client. Caller is responsible for TLS ≥ 1.2 and not setting `InsecureSkipVerify`. |
| `WithFormsTimeout(d time.Duration)` | Set the per-request timeout on the default HTTP client. Ignored if `WithFormsHTTPClient` is also used. |
| `WithFormsRetry(cfg RetryConfig)` | Configure retry behaviour (same `RetryConfig` as the Email client). GET/DELETE retry on 429/5xx; POST/PUT are not retried unless `RetryNonIdempotent` is true. |
| `WithFormsUserAgent(ua string)` | Prepend a custom token to the `User-Agent` header. The SDK identifier is always appended. |

---

### `ListForms`

```go
func (c *FormsClient) ListForms(ctx context.Context, params *ListFormsParams) (*ListFormsResponse, error)
```

Lists forms visible to the API key's customer. `params` may be nil for the server defaults (page 1, 50 items, ordered by `created_at` descending). `GET /api/forms`

**`ListFormsParams`** (zero values are omitted from the query string)

| Field | Type | Query param | Description |
|---|---|---|---|
| `CustomerID` | `int` | `customer_id` | Filter by customer. |
| `FormID` | `string` | `form_id` | Filter to a single form. |
| `Search` | `string` | `search` | Matches title/description (LIKE). |
| `Order` | `string` | `order` | `"asc"` or `"desc"` (server default `desc`). |
| `OrderBy` | `string` | `order_by` | Allowlisted server-side: `title`, `updated_at`, `submission_count`; default `created_at`. |
| `Archived` | `*bool` | `archived` | Filter by archived state. Use `paubox.Ptr`. |
| `Active` | `*bool` | `active` | Filter by active state. Use `paubox.Ptr`. |
| `Page` | `int` | `page` | 1-based page number. |
| `Items` | `int` | `items` | Page size. Server caps at 100, default 50. |

**`ListFormsResponse`**

| Field | Type | Description |
|---|---|---|
| `Results` | `[]Form` | The current page of forms. |
| `PageInfo` | `FormPageInfo` | Pagination metadata: `Count` (`int64`, total matches), `Pages`, `Page`, `Items`. |

---

### `CreateForm`

```go
func (c *FormsClient) CreateForm(ctx context.Context, req *CreateFormRequest) (*CreateFormResponse, error)
```

Creates a new form and returns its UUID. `POST /api/forms`

**`CreateFormRequest`**

| Field | Type | Required | Description |
|---|---|---|---|
| `Title` | `string` | ✅ | Form title. |
| `CustomerID` | `int` | ✅ | Owning customer ID (> 0). |
| `FormJSON` | `json.RawMessage` | ✅ | Form-builder JSON definition. |
| `Version` | `int` | | Required by the API; the SDK defaults 0 → 1. |
| `Description` | `string` | | Form description. |
| `FormHTML` | `string` | | Rendered HTML for the form. |
| `FormCSS` | `string` | | Custom CSS. |
| `Recipient` | `string` | | Comma-separated notification recipient emails. |
| `Signable` | `bool` | | Enable signature support. |
| `SignatureConfirmationLabel` | `string` | | Label for the signature confirmation checkbox. |
| `SubscriptionListID` | `string` | | Linked marketing subscription list. |
| `Type` | `string` | | Form type (e.g. `"marketing_form"`). |
| `Active` | `bool` | | Whether the form accepts submissions. |
| `SubmissionCount` | `int` | | Seeds the submission counter. |

**`CreateFormResponse`**

| Field | Type | Description |
|---|---|---|
| `ID` | `string` | The new form's UUID. |

---

### `GetForm`

```go
func (c *FormsClient) GetForm(ctx context.Context, id string) (*Form, error)
```

Retrieves a single form by UUID. The SDK unwraps the API's `{"data": {…}}` envelope. `GET /api/forms/{id}`

---

### `UpdateForm`

```go
func (c *FormsClient) UpdateForm(ctx context.Context, id string, req *UpdateFormRequest) (*UpdateFormResponse, error)
```

Applies a partial update. PATCH-style semantics: nil pointer fields (and a nil `FormJSON`) are omitted from the request and left unchanged on the server — use `paubox.Ptr` to set fields inline. `PUT /api/forms/{id}`

**`UpdateFormRequest`**

| Field | Type | Description |
|---|---|---|
| `Title` | `*string` | New title. |
| `Description` | `*string` | New description. |
| `FormJSON` | `json.RawMessage` | New form-builder JSON definition. |
| `VanityURL` | `*string` | New vanity URL slug. |
| `Recipient` | `*string` | New comma-separated notification recipient list. |
| `Active` | `*bool` | Toggle whether the form accepts submissions. |
| `SubscriptionListID` | `*string` | New linked subscription list. |

**`UpdateFormResponse`**

| Field | Type | Description |
|---|---|---|
| `Detail` | `string` | Human-readable result message. |
| `FormID` | `string` | UUID of the updated form. |

---

### `ArchiveForm` / `UnarchiveForm`

```go
func (c *FormsClient) ArchiveForm(ctx context.Context, id string) (*FormActionResponse, error)
func (c *FormsClient) UnarchiveForm(ctx context.Context, id string) (*FormActionResponse, error)
```

Archiving also deactivates the form; unarchiving does **not** reactivate it (toggle `Active` via `UpdateForm`). `POST /api/forms/{id}/archive`, `POST /api/forms/{id}/unarchive`

**`FormActionResponse`**

| Field | Type | Description |
|---|---|---|
| `Detail` | `string` | Human-readable result message. |

---

### `CopyForm`

```go
func (c *FormsClient) CopyForm(ctx context.Context, req *CopyFormRequest) (*Form, error)
```

Duplicates an existing form under a new title and returns the new `Form` (bare object, no envelope). The copy starts with a fresh submission counter and no vanity URL. `POST /api/forms/copy`

**`CopyFormRequest`**

| Field | Type | Required | Description |
|---|---|---|---|
| `FormID` | `string` | ✅ | UUID of the form to copy. |
| `Title` | `string` | ✅ | Title for the copy. |

---

### `GetFormStats`

```go
func (c *FormsClient) GetFormStats(ctx context.Context, params *FormStatsParams) (*FormStats, error)
```

Retrieves aggregate form statistics. `params` may be nil, scoping the stats to the API key's customer; set `FormStatsParams.CustomerID` to query another customer. `GET /api/forms/stats`

**`FormStats`**

| Field | Type | Description |
|---|---|---|
| `ActiveFormCount` | `int64` | Number of active forms. |
| `TotalSubmissionCount` | `int64` | All-time submission count. |
| `SubmissionsLast7Days` | `int64` | Submissions over the last 7 days. |

---

### `GetPublicForm` (public — no API key required)

```go
func (c *FormsClient) GetPublicForm(ctx context.Context, formID string) (*Form, error)
```

Retrieves the public definition of an active form. Works on a keyless client. Returns 404 (`ErrNotFound`) when the form is inactive, archived, or deleted. `GET /public/form_data/{formID}`

---

### `SubmitForm` (public — no API key required)

```go
func (c *FormsClient) SubmitForm(ctx context.Context, formID string, req *SubmitFormRequest) error
```

Submits a filled-out form. Works on a keyless client. A successful submission returns nil (the service responds 201 with an empty body). `POST /api/forms/{formID}/submissions`

**`SubmitFormRequest`**

| Field | Type | Required | Description |
|---|---|---|---|
| `FormData` | `map[string]any` | ✅ | Submitted field values keyed by the form's slugified field names. Sent as a real JSON object — do not pre-encode. |
| `Attachments` | `[]FormSubmissionAttachment` | | Optional file attachments. |

**`FormSubmissionAttachment`**

| Field | Type | Required | Description |
|---|---|---|---|
| `Name` | `string` | ✅ | Filename. |
| `Content` | `[]byte` | ✅ | Raw file bytes. The SDK base64-encodes them on the wire. |

---

### `ListFormSubmissions`

```go
func (c *FormsClient) ListFormSubmissions(ctx context.Context, formID string, params *ListFormSubmissionsParams) (*ListFormSubmissionsResponse, error)
```

Lists a form's submissions. `params` may be nil for the server defaults. `GET /api/forms/{formID}/submissions`

**`ListFormSubmissionsParams`** (zero values are omitted from the query string)

| Field | Type | Query param | Description |
|---|---|---|---|
| `SubmissionID` | `string` | `submission_id` | Filter to one submission. |
| `OrderBy` | `string` | `order_by` | Allowlisted server-side: `submitter_email`; default `created_at`. |
| `Order` | `string` | `order` | `"asc"` or `"desc"`. |
| `Page` | `int` | `page` | 1-based page number. |
| `Items` | `int` | `items` | Page size. Server caps at 100. |

**`ListFormSubmissionsResponse`**

| Field | Type | Description |
|---|---|---|
| `Data` | `[]FormSubmission` | The current page of submissions. |
| `Total` | `int64` | Total number of matching submissions. |
| `Page` | `int` | Current (1-based) page number. |
| `Items` | `int` | Page size used. |

**`FormSubmission`**

| Field | Type | Description |
|---|---|---|
| `ID` | `string` | Submission UUID. |
| `FormID` | `string` | UUID of the submitted form. |
| `FormData` | `string` | **JSON-encoded string** of the submitted fields (the server re-serializes the payload). Unmarshal it yourself to access individual fields. |
| `StorageType` | `string` | Where the payload is stored. |
| `StorageURL` | `*string` | Storage location URL, when present. |
| `SubmitterEmail` | `*string` | Submitter's email, when captured. |
| `Recipients` | `*string` | Comma-separated notification recipients used, when present. |
| `Attachment` | `*string` | Stored attachment reference, when present. |
| `AttachmentName` | `*string` | Attachment filename, when present. |
| `AttachmentURL` | `*string` | Attachment download URL, when present. |
| `AttachmentType` | `*string` | Attachment content type, when present. |
| `CreatedAt` | `time.Time` | When the submission was received. |

---

### Submission exports

```go
func (c *FormsClient) ExportFormSubmissionsCSV(ctx context.Context, formID string) ([]byte, error)
func (c *FormsClient) ExportFormSubmissionCSV(ctx context.Context, formID, submissionID string) ([]byte, error)
func (c *FormsClient) ExportFormSubmissionPDF(ctx context.Context, formID, submissionID string) ([]byte, error)
```

Return the raw exported file bytes. The bytes may contain PHI — handle and store them accordingly.

| Method | Endpoint |
|---|---|
| `ExportFormSubmissionsCSV` | `GET /api/forms/{formID}/submissions/submission-csv` |
| `ExportFormSubmissionCSV` | `GET /api/forms/{formID}/submissions/submission-csv/{submissionID}` |
| `ExportFormSubmissionPDF` | `GET /api/forms/{formID}/submissions/{submissionID}/submission-pdf` |

---

**`Form`** (returned by `GetForm`, `CopyForm`, `GetPublicForm`, and in `ListFormsResponse.Results`)

| Field | Type | Description |
|---|---|---|
| `ID` | `string` | Form UUID. |
| `Title` | `string` | Display title. |
| `Description` | `*string` | Optional description. |
| `FormHTML` | `*string` | Rendered HTML, when present. |
| `FormJSON` | `json.RawMessage` | Arbitrary form-builder JSON definition. |
| `FormCSS` | `*string` | Custom CSS, when present. |
| `VanityURL` | `*string` | Vanity URL slug, when set. |
| `Version` | `int` | Form definition version. |
| `Active` | `bool` | Whether the form accepts submissions. |
| `CustomerID` | `int` | Owning Paubox customer ID. |
| `OldFormID` | `*int` | Legacy form ID, when migrated. |
| `CreatedAt` / `UpdatedAt` | `time.Time` | Creation / last-update timestamps. |
| `Recipient` | `*string` | Comma-separated notification recipients, when configured. |
| `Signable` | `bool` | Whether the form supports signatures. |
| `SignatureConfirmationLabel` | `*string` | Signature confirmation checkbox label, when configured. |
| `SubmissionCount` | `int` | Number of submissions received. |
| `Type` | `*string` | Form type (e.g. `"marketing_form"`), when set. |
| `SubscriptionListID` | `*string` | Linked marketing subscription list, when set. |
| `Deleted` | `bool` | Whether the form is soft-deleted. |
| `Archived` | `bool` | Whether the form is archived. |

---

## Errors

All API errors — from both the Email `Client` and the `FormsClient` — are returned as `*PauboxError`. Use `errors.Is` with sentinels to match by HTTP status, or `errors.As` to access the full detail. The two services use different error envelopes on the wire (`{"errors":[…]}` for Email, `{"message":"…"}` for Forms); the SDK normalises both into the same type.

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

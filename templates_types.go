package paubox

import "time"

// ---------------------------------------------------------------------------
// Dynamic Template types
// ---------------------------------------------------------------------------

// Template is a dynamic email template stored in your Paubox account.
// Template bodies use Handlebars syntax: {{variableName}}.
type Template struct {
	// ID is the unique numeric identifier for this template.
	ID int64 `json:"id"`

	// Name is the human-readable template name.
	Name string `json:"name"`

	// APICustomerID is the account identifier the template belongs to.
	APICustomerID int64 `json:"api_customer_id,omitempty"`

	// Body is the Handlebars template content. Present on a single-template
	// fetch; not returned by ListTemplates.
	Body string `json:"body,omitempty"`

	// Metadata holds arbitrary template metadata returned by the API.
	Metadata map[string]any `json:"metadata,omitempty"`

	// CreatedAt is the time the template was created, when returned.
	CreatedAt *time.Time `json:"created_at,omitempty"`

	// UpdatedAt is the time the template was last modified, when returned.
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// ListTemplatesResponse is the response from [Client.ListTemplates].
type ListTemplatesResponse struct {
	// Templates is the list of all templates in your Paubox account.
	Templates []Template `json:"templates"`
}

// DeleteTemplateResponse is the response from [Client.DeleteTemplate].
type DeleteTemplateResponse struct {
	// Message is the confirmation message returned by the API.
	Message string `json:"message"`
}

// TemplateMutationResponse is returned by [Client.CreateTemplate] and
// [Client.UpdateTemplate].
//
// The Paubox API confirms the operation with a human-readable message and
// echoes the submitted name; it does NOT return the template's ID, body, or
// timestamps. To obtain a newly created template's ID, call
// [Client.ListTemplates] and match on Name.
type TemplateMutationResponse struct {
	// Message is the confirmation message returned by the API,
	// e.g. "Template welcome created!".
	Message string `json:"message"`

	// Params echoes the accepted template parameters.
	Params TemplateMutationParams `json:"params"`
}

// TemplateMutationParams holds the template parameters echoed back by the API
// on create and update.
type TemplateMutationParams struct {
	// Name is the template name the API recorded.
	Name string `json:"name"`
}

// ---------------------------------------------------------------------------
// Create / Update template requests
// ---------------------------------------------------------------------------

// CreateTemplateRequest is the request for [Client.CreateTemplate].
type CreateTemplateRequest struct {
	// Name is the human-readable name for the template. Required.
	Name string

	// Body is the Handlebars template content. Required.
	// Uploaded as a binary .hbs file via multipart/form-data internally.
	Body []byte
}

// UpdateTemplateRequest is the request for [Client.UpdateTemplate].
// Both fields are optional — supply only the ones you want to change.
// At least one must be non-empty.
type UpdateTemplateRequest struct {
	// Name is the updated template name. Leave empty to keep the current value.
	Name string

	// Body is the updated Handlebars content. Leave nil to keep the current value.
	Body []byte
}

// ---------------------------------------------------------------------------
// Send templated message
// ---------------------------------------------------------------------------

// SendTemplatedMessageRequest is the request for [Client.SendTemplatedMessage].
type SendTemplatedMessageRequest struct {
	// TemplateName is the exact name of the template to use. Required.
	TemplateName string

	// TemplateValues is a map of Handlebars variable names to their values.
	// The SDK serialises this map to a JSON string before sending — do not
	// pre-encode it yourself. The Paubox API requires a JSON-encoded string,
	// not a JSON object, as the value of template_values.
	TemplateValues map[string]any

	// Message contains routing and delivery options. Content is provided by
	// the template and must not be set here.
	Message TemplatedMessage
}

// TemplatedMessage is the message envelope for a templated send. It mirrors
// [Message] but omits Content (which the template provides).
type TemplatedMessage struct {
	// Recipients is the list of To: addresses. Required; minimum one entry.
	Recipients []string `json:"recipients"`

	// BCC is the list of blind carbon-copy recipients.
	BCC []string `json:"bcc,omitempty"`

	// CC is the list of carbon-copy recipients.
	CC []string `json:"cc,omitempty"`

	// Headers contains the required and optional email header fields.
	Headers MessageHeaders `json:"headers"`

	// AllowNonTLS permits delivery without TLS. Defaults to false.
	AllowNonTLS bool `json:"allowNonTLS,omitempty"`

	// ForceSecureNotification forces delivery via the Paubox Secure Message portal.
	ForceSecureNotification bool `json:"forceSecureNotification,omitempty"`

	// Attachments is an optional list of file attachments.
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Wire types — not exported.

type sendTemplatedMessageWire struct {
	Data sendTemplatedMessageData `json:"data"`
}

type sendTemplatedMessageData struct {
	TemplateName   string           `json:"template_name"`
	TemplateValues string           `json:"template_values"` // JSON-encoded string, not object
	Message        TemplatedMessage `json:"message"`
}

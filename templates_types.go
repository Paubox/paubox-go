package paubox

import "time"

// ---------------------------------------------------------------------------
// Dynamic Template types
// ---------------------------------------------------------------------------

// Template is a dynamic email template stored in your Paubox account.
// Template bodies use Handlebars syntax: {{variableName}}.
type Template struct {
	// ID is the unique identifier for this template.
	ID string `json:"id"`

	// Name is the human-readable template name.
	Name string `json:"name"`

	// Body is the Handlebars template content.
	Body string `json:"body"`

	// CreatedAt is the time the template was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the time the template was last modified.
	UpdatedAt time.Time `json:"updated_at"`
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

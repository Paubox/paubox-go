package paubox

import (
	"encoding/json"
	"time"
)

// Form is the metadata and field schema for a Paubox Form,
// as returned by [FormsClient.GetForm].
type Form struct {
	ID                         string      `json:"id"`
	Title                      string      `json:"title"`
	Description                *string     `json:"description"`
	FormHTML                   *string     `json:"form_html"`
	FormJSON                   *FormJSON   `json:"form_json"`
	FormCSS                    *string     `json:"form_css"`
	VanityURL                  *string     `json:"vanity_url"`
	Version                    json.Number `json:"version"` // API may return int or quoted string
	Active                     bool        `json:"active"`
	CustomerID                 json.Number `json:"customer_id"` // API may return int or quoted string
	OldFormID                  *int64      `json:"old_form_id"`
	CreatedAt                  time.Time   `json:"created_at"`
	UpdatedAt                  time.Time   `json:"updated_at"`
	Recipient                  *string     `json:"recipient"`
	Signable                   bool        `json:"signable"`
	SignatureConfirmationLabel *string     `json:"signature_confirmation_label"`
	SubmissionCount            int         `json:"submission_count"`
	Type                       *string     `json:"type"`
	SubscriptionListID         *string     `json:"subscription_list_id"`
	Deleted                    bool        `json:"deleted"`
	Archived                   bool        `json:"archived"`
}

// FormJSON is the form builder schema embedded in [Form.FormJSON].
type FormJSON struct {
	Body []FormField `json:"body"`
}

// FormField is one field definition within [FormJSON.Body].
// The Properties schema varies by Type ("Text", "Signature", etc.),
// so it is kept as raw JSON for caller inspection.
type FormField struct {
	Type       string          `json:"type"`
	ID         string          `json:"id"`
	Properties json.RawMessage `json:"properties,omitempty"`
}

// FormSubmission is the request body for [FormsClient.SubmitForm].
type FormSubmission struct {
	FormData    map[string]any   `json:"form_data"`
	Attachments []FormAttachment `json:"attachments,omitempty"`
}

// FormAttachment is a file included with a form submission.
type FormAttachment struct {
	Name    string `json:"name"`
	Content string `json:"content"` // base64-encoded
}

// FormSubmitResponse is the success acknowledgement from [FormsClient.SubmitForm].
// The pb_rforms service returns 201 No Content on success; this struct is empty
// and reserved for a future API response body.
type FormSubmitResponse struct{}

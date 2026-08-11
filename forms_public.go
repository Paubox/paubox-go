package paubox

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// submitFormWire is the wire body for FormsClient.SubmitForm:
//
//	{"form_data": {...}, "attachments": [{"name": "...", "content": "<base64>"}]}
//
// The attachments key is omitted entirely when there are none.
type submitFormWire struct {
	FormData    map[string]any             `json:"form_data"`
	Attachments []submitFormAttachmentWire `json:"attachments,omitempty"`
}

type submitFormAttachmentWire struct {
	Name    string `json:"name"`
	Content string `json:"content"` // unpadded standard base64 — the service rejects '=' padding
}

// GetPublicForm retrieves the public definition of an active form. This is a
// public endpoint: no API key is required, so it works on a public-only
// (keyless) client.
//
// The Forms service returns 404 ([ErrNotFound]) when the form is inactive,
// archived, or deleted.
//
// API: GET /public/form_data/{formID}
func (c *FormsClient) GetPublicForm(ctx context.Context, formID string) (*Form, error) {
	if strings.TrimSpace(formID) == "" {
		return nil, fmt.Errorf("paubox: GetPublicForm: formID must not be empty")
	}

	var form Form
	if err := c.doJSON(ctx, http.MethodGet, "/public/form_data/"+url.PathEscape(formID), nil, &form); err != nil {
		return nil, err
	}
	return &form, nil
}

// SubmitForm submits a filled-out form. This is a public endpoint: no API key
// is required, so it works on a public-only (keyless) client.
//
// req.FormData holds the submitted field values keyed by the form's
// slugified field names and is sent as a real JSON object (unlike the Email
// API's template_values, which travels as a JSON-encoded string). Attachment
// contents are encoded by the SDK on the wire as unpadded standard base64
// (the Forms service rejects padded input).
//
// A successful submission returns nil (the service responds 201 with an
// empty body).
//
// API: POST /api/forms/{formID}/submissions
func (c *FormsClient) SubmitForm(ctx context.Context, formID string, req *SubmitFormRequest) error {
	if strings.TrimSpace(formID) == "" {
		return fmt.Errorf("paubox: SubmitForm: formID must not be empty")
	}
	if req == nil {
		return fmt.Errorf("paubox: SubmitForm: request must not be nil")
	}
	if len(req.FormData) == 0 {
		return fmt.Errorf("paubox: SubmitForm: form data must not be empty")
	}
	for i, a := range req.Attachments {
		if strings.TrimSpace(a.Name) == "" {
			return fmt.Errorf("paubox: SubmitForm: attachment[%d]: name must not be empty", i)
		}
		if len(a.Content) == 0 {
			return fmt.Errorf("paubox: SubmitForm: attachment %q: content must not be empty", a.Name)
		}
	}

	wire := submitFormWire{FormData: req.FormData}
	if len(req.Attachments) > 0 {
		wire.Attachments = make([]submitFormAttachmentWire, 0, len(req.Attachments))
		for _, a := range req.Attachments {
			wire.Attachments = append(wire.Attachments, submitFormAttachmentWire{
				Name:    a.Name,
				Content: base64.RawStdEncoding.EncodeToString(a.Content),
			})
		}
	}

	return c.doJSON(ctx, http.MethodPost, "/api/forms/"+url.PathEscape(formID)+"/submissions", wire, nil)
}

package paubox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/paubox/paubox-go/internal"
)

// ListTemplates retrieves all dynamic templates in your Paubox account.
//
// API: GET /dynamic_templates
func (c *Client) ListTemplates(ctx context.Context) (*ListTemplatesResponse, error) {
	// The API returns a bare JSON array of templates, not an object wrapper.
	var templates []Template
	if err := c.doJSON(ctx, http.MethodGet, "/dynamic_templates", nil, &templates); err != nil {
		return nil, err
	}
	return &ListTemplatesResponse{Templates: templates}, nil
}

// GetTemplate retrieves a single dynamic template by its ID.
//
// API: GET /dynamic_templates/{id}
func (c *Client) GetTemplate(ctx context.Context, id int64) (*Template, error) {
	if id <= 0 {
		return nil, fmt.Errorf("paubox: GetTemplate: id must be a positive template ID")
	}
	var tmpl Template
	if err := c.doJSON(ctx, http.MethodGet, "/dynamic_templates/"+strconv.FormatInt(id, 10), nil, &tmpl); err != nil {
		return nil, err
	}
	return &tmpl, nil
}

// CreateTemplate uploads a new Handlebars dynamic template to your account.
//
// Both Name and Body are required. The body is uploaded as a binary .hbs file
// using multipart/form-data — callers pass raw bytes and the SDK handles
// the encoding.
//
// The API confirms creation with a message but does not return the new
// template's ID. To obtain it, call [Client.ListTemplates] and match on Name.
//
// API: POST /dynamic_templates
func (c *Client) CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (*TemplateMutationResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("paubox: CreateTemplate: request must not be nil")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("paubox: CreateTemplate: Name must not be empty")
	}
	if len(req.Body) == 0 {
		return nil, fmt.Errorf("paubox: CreateTemplate: Body must not be empty")
	}

	formBody, ct, err := internal.BuildTemplateForm(req.Name, req.Body)
	if err != nil {
		return nil, fmt.Errorf("paubox: CreateTemplate: %w", err)
	}
	return c.doMultipartTemplateMutation(ctx, http.MethodPost, "/dynamic_templates", formBody, ct)
}

// UpdateTemplate modifies an existing dynamic template.
//
// Provide only the fields you want to change; omitted fields retain their
// current values. At least one of Name or Body must be non-empty.
//
// The API confirms the update with a message and does not return the updated
// template object.
//
// API: PATCH /dynamic_templates/{id}
func (c *Client) UpdateTemplate(ctx context.Context, id int64, req *UpdateTemplateRequest) (*TemplateMutationResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("paubox: UpdateTemplate: id must be a positive template ID")
	}
	if req == nil {
		return nil, fmt.Errorf("paubox: UpdateTemplate: request must not be nil")
	}
	if req.Name == "" && len(req.Body) == 0 {
		return nil, fmt.Errorf("paubox: UpdateTemplate: at least one of Name or Body must be provided")
	}

	formBody, ct, err := internal.BuildTemplateForm(req.Name, req.Body)
	if err != nil {
		return nil, fmt.Errorf("paubox: UpdateTemplate: %w", err)
	}
	return c.doMultipartTemplateMutation(ctx, http.MethodPatch, "/dynamic_templates/"+strconv.FormatInt(id, 10), formBody, ct)
}

// DeleteTemplate permanently removes a dynamic template by its ID.
//
// API: DELETE /dynamic_templates/{id}
func (c *Client) DeleteTemplate(ctx context.Context, id int64) (*DeleteTemplateResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("paubox: DeleteTemplate: id must be a positive template ID")
	}
	var resp DeleteTemplateResponse
	if err := c.doJSON(ctx, http.MethodDelete, "/dynamic_templates/"+strconv.FormatInt(id, 10), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendTemplatedMessage sends an email rendered from a dynamic Handlebars
// template stored in your Paubox account.
//
// TemplateValues is a plain Go map; the SDK serialises it to the JSON-encoded
// string that the API requires. Do not pre-encode the values yourself.
//
// API: POST /templated_messages
func (c *Client) SendTemplatedMessage(ctx context.Context, req *SendTemplatedMessageRequest) (*SendMessageResponse, error) {
	if err := validateTemplatedMessageRequest(req); err != nil {
		return nil, err
	}

	// The API requires template_values to be a JSON-encoded string, not a
	// JSON object. We marshal the map here and pass the resulting string.
	tvJSON, err := json.Marshal(req.TemplateValues)
	if err != nil {
		return nil, fmt.Errorf("paubox: SendTemplatedMessage: encoding TemplateValues: %w", err)
	}

	wire := sendTemplatedMessageWire{
		Data: sendTemplatedMessageData{
			TemplateName:   req.TemplateName,
			TemplateValues: string(tvJSON),
			Message:        req.Message,
		},
	}

	var resp SendMessageResponse
	if err := c.doJSON(ctx, http.MethodPost, "/templated_messages", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// doMultipartTemplateMutation executes a multipart/form-data request and
// decodes the message-style response the API returns for template create and
// update. Used by CreateTemplate and UpdateTemplate.
func (c *Client) doMultipartTemplateMutation(ctx context.Context, method, path string, body []byte, contentType string) (*TemplateMutationResponse, error) {
	resp, err := c.do(ctx, method, path, bytes.NewReader(body), contentType)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // close-on-defer; read errors already reported by ReadAll above

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("paubox: reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseAPIError(resp.StatusCode, resp.Header.Get("X-Request-Id"), raw)
	}

	var out TemplateMutationResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("paubox: decoding template response: %w", err)
	}
	return &out, nil
}

func validateTemplatedMessageRequest(req *SendTemplatedMessageRequest) error {
	if req == nil {
		return fmt.Errorf("paubox: SendTemplatedMessage: request must not be nil")
	}
	if req.TemplateName == "" {
		return fmt.Errorf("paubox: SendTemplatedMessage: TemplateName must not be empty")
	}
	if len(req.Message.Recipients) == 0 {
		return fmt.Errorf("paubox: SendTemplatedMessage: message.recipients must not be empty")
	}
	if req.Message.Headers.From == "" {
		return fmt.Errorf("paubox: SendTemplatedMessage: message.headers.from must not be empty")
	}
	if req.Message.Headers.Subject == "" {
		return fmt.Errorf("paubox: SendTemplatedMessage: message.headers.subject must not be empty")
	}
	return nil
}

package paubox

// Core CRUD methods for the Paubox Forms API: ListForms, CreateForm, GetForm
// and UpdateForm. The transport lives in forms_client.go and the
// request/response types in forms_types.go.

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// encode renders the params as a URL query string. Zero values are omitted;
// a nil receiver yields an empty string.
func (p *ListFormsParams) encode() string {
	if p == nil {
		return ""
	}
	q := url.Values{}
	if p.CustomerID > 0 {
		q.Set("customer_id", strconv.Itoa(p.CustomerID))
	}
	if p.FormID != "" {
		q.Set("form_id", p.FormID)
	}
	if p.Search != "" {
		q.Set("search", p.Search)
	}
	if p.Order != "" {
		q.Set("order", p.Order)
	}
	if p.OrderBy != "" {
		q.Set("order_by", p.OrderBy)
	}
	if p.Archived != nil {
		q.Set("archived", strconv.FormatBool(*p.Archived))
	}
	if p.Active != nil {
		q.Set("active", strconv.FormatBool(*p.Active))
	}
	if p.Page > 0 {
		q.Set("page", strconv.Itoa(p.Page))
	}
	if p.Items > 0 {
		q.Set("items", strconv.Itoa(p.Items))
	}
	return q.Encode()
}

// ListForms lists forms visible to the API key's customer.
//
// When authenticating with a scoped API key, params.CustomerID must be set
// to the key's customer ID (or a customer related to it) — the service
// authorizes the listing against that value and responds 403 [ErrForbidden]
// when it is omitted. params may still be nil (accepted for forward
// compatibility with other auth modes); once authorization passes, the
// server defaults apply (page 1, 50 items per page, ordered by created_at
// descending).
//
// API: GET /api/forms
func (c *FormsClient) ListForms(ctx context.Context, params *ListFormsParams) (*ListFormsResponse, error) {
	if err := c.requireKey("ListForms"); err != nil {
		return nil, err
	}

	path := "/api/forms"
	if q := params.encode(); q != "" {
		path += "?" + q
	}

	var resp ListFormsResponse
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateForm creates a new form and returns its ID.
//
// Title, CustomerID and FormJSON are required. The API requires an explicit
// Version; when the zero value is passed the SDK defaults it to 1 (the
// caller's request struct is not mutated).
//
// API: POST /api/forms
func (c *FormsClient) CreateForm(ctx context.Context, req *CreateFormRequest) (*CreateFormResponse, error) {
	if err := c.requireKey("CreateForm"); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, fmt.Errorf("paubox: CreateForm: request must not be nil")
	}
	if strings.TrimSpace(req.Title) == "" {
		return nil, fmt.Errorf("paubox: CreateForm: title must not be empty")
	}
	if req.CustomerID <= 0 {
		return nil, fmt.Errorf("paubox: CreateForm: customer_id must be greater than zero")
	}
	if len(req.FormJSON) == 0 {
		return nil, fmt.Errorf("paubox: CreateForm: form_json must not be empty")
	}

	body := *req
	if body.Version == 0 {
		body.Version = 1
	}

	var resp CreateFormResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/forms", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// getFormWire is the {"data": {...}} envelope returned by GET /api/forms/{id}.
type getFormWire struct {
	Data Form `json:"data"`
}

// GetForm retrieves a single form by its UUID.
//
// API: GET /api/forms/{id}
func (c *FormsClient) GetForm(ctx context.Context, id string) (*Form, error) {
	if err := c.requireKey("GetForm"); err != nil {
		return nil, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("paubox: GetForm: id must not be empty")
	}

	var wire getFormWire
	if err := c.doJSON(ctx, http.MethodGet, "/api/forms/"+url.PathEscape(id), nil, &wire); err != nil {
		return nil, err
	}
	return &wire.Data, nil
}

// UpdateForm applies a partial update to a form.
//
// Only fields set on req are sent; nil pointer fields (and a nil FormJSON)
// are omitted from the request body and left unchanged on the server. Use
// [Ptr] to set pointer fields inline.
//
// API: PUT /api/forms/{id}
func (c *FormsClient) UpdateForm(ctx context.Context, id string, req *UpdateFormRequest) (*UpdateFormResponse, error) {
	if err := c.requireKey("UpdateForm"); err != nil {
		return nil, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("paubox: UpdateForm: id must not be empty")
	}
	if req == nil {
		return nil, fmt.Errorf("paubox: UpdateForm: request must not be nil")
	}

	var resp UpdateFormResponse
	if err := c.doJSON(ctx, http.MethodPut, "/api/forms/"+url.PathEscape(id), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

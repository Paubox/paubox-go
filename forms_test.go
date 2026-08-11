package paubox

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Shared fixtures for the core CRUD tests
// ---------------------------------------------------------------------------

const crudFormJSON = `{"fields":[{"name":"email","type":"text"}]}`

func validCreateFormRequest() *CreateFormRequest {
	return &CreateFormRequest{
		Title:      "Intake Form",
		CustomerID: 42,
		FormJSON:   json.RawMessage(crudFormJSON),
	}
}

// crudFormBody is a full Form JSON object as the Forms service serializes it.
const crudFormBody = `{
	"id": "f-123",
	"title": "Intake Form",
	"description": "Patient intake",
	"form_html": null,
	"form_json": {"fields":[]},
	"form_css": null,
	"vanity_url": "intake",
	"version": 2,
	"active": true,
	"customer_id": 42,
	"old_form_id": null,
	"created_at": "2026-01-02T03:04:05Z",
	"updated_at": "2026-01-03T06:07:08Z",
	"recipient": "a@example.com,b@example.com",
	"signable": false,
	"signature_confirmation_label": null,
	"submission_count": 7,
	"type": "marketing_form",
	"subscription_list_id": null,
	"deleted": false,
	"archived": false
}`

// ---------------------------------------------------------------------------
// ListForms
// ---------------------------------------------------------------------------

func TestListForms_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"results":[`+crudFormBody+`],"page_info":{"count":1,"pages":1,"page":1,"items":50}}`)
	}))
	defer srv.Close()

	resp, err := newTestFormsClient(t, srv).ListForms(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListForms() error: %v", err)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("len(Results) = %d, want 1", len(resp.Results))
	}
	f := resp.Results[0]
	if f.ID != "f-123" || f.Title != "Intake Form" || f.CustomerID != 42 {
		t.Errorf("form = %+v", f)
	}
	if resp.PageInfo.Count != 1 || resp.PageInfo.Pages != 1 || resp.PageInfo.Page != 1 || resp.PageInfo.Items != 50 {
		t.Errorf("PageInfo = %+v", resp.PageInfo)
	}
}

func TestListForms_MethodAndPath_NilParams(t *testing.T) {
	var gotMethod, gotPath, gotRawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotRawQuery = r.URL.RawQuery
		respondJSON(w, http.StatusOK, `{"results":[],"page_info":{"count":0,"pages":0,"page":1,"items":50}}`)
	}))
	defer srv.Close()

	if _, err := newTestFormsClient(t, srv).ListForms(context.Background(), nil); err != nil {
		t.Fatalf("ListForms() error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/api/forms" {
		t.Errorf("request = %s %s, want GET /api/forms", gotMethod, gotPath)
	}
	if gotRawQuery != "" {
		t.Errorf("query = %q, want empty for nil params", gotRawQuery)
	}
}

func TestListForms_QueryParams(t *testing.T) {
	var gotQuery map[string][]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		respondJSON(w, http.StatusOK, `{"results":[],"page_info":{"count":0,"pages":0,"page":2,"items":25}}`)
	}))
	defer srv.Close()

	_, err := newTestFormsClient(t, srv).ListForms(context.Background(), &ListFormsParams{
		CustomerID: 42,
		FormID:     "f-1",
		Search:     "intake",
		Order:      "asc",
		OrderBy:    "title",
		Archived:   Ptr(false),
		Active:     Ptr(true),
		Page:       2,
		Items:      25,
	})
	if err != nil {
		t.Fatalf("ListForms() error: %v", err)
	}

	want := map[string]string{
		"customer_id": "42",
		"form_id":     "f-1",
		"search":      "intake",
		"order":       "asc",
		"order_by":    "title",
		"archived":    "false",
		"active":      "true",
		"page":        "2",
		"items":       "25",
	}
	for key, val := range want {
		got := gotQuery[key]
		if len(got) != 1 || got[0] != val {
			t.Errorf("query[%q] = %v, want [%q]", key, got, val)
		}
	}
	if len(gotQuery) != len(want) {
		t.Errorf("query has %d keys, want %d: %v", len(gotQuery), len(want), gotQuery)
	}
}

func TestListForms_ZeroValueParamsOmitted(t *testing.T) {
	var gotRawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRawQuery = r.URL.RawQuery
		respondJSON(w, http.StatusOK, `{"results":[],"page_info":{"count":0,"pages":0,"page":1,"items":50}}`)
	}))
	defer srv.Close()

	if _, err := newTestFormsClient(t, srv).ListForms(context.Background(), &ListFormsParams{}); err != nil {
		t.Fatalf("ListForms() error: %v", err)
	}
	if gotRawQuery != "" {
		t.Errorf("query = %q, want empty for zero-value params", gotRawQuery)
	}
}

// ---------------------------------------------------------------------------
// CreateForm
// ---------------------------------------------------------------------------

func TestCreateForm_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"id":"new-form-uuid"}`)
	}))
	defer srv.Close()

	resp, err := newTestFormsClient(t, srv).CreateForm(context.Background(), validCreateFormRequest())
	if err != nil {
		t.Fatalf("CreateForm() error: %v", err)
	}
	if resp.ID != "new-form-uuid" {
		t.Errorf("ID = %q, want new-form-uuid", resp.ID)
	}
}

func TestCreateForm_MethodAndPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"id":"x"}`)
	}))
	defer srv.Close()

	_, _ = newTestFormsClient(t, srv).CreateForm(context.Background(), validCreateFormRequest())
	if gotMethod != http.MethodPost || gotPath != "/api/forms" {
		t.Errorf("request = %s %s, want POST /api/forms", gotMethod, gotPath)
	}
}

func TestCreateForm_RequestBody_AllFields(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		respondJSON(w, http.StatusOK, `{"id":"x"}`)
	}))
	defer srv.Close()

	req := &CreateFormRequest{
		Title:                      "Intake Form",
		CustomerID:                 42,
		FormJSON:                   json.RawMessage(crudFormJSON),
		Version:                    3,
		Description:                "Patient intake",
		FormHTML:                   "<form></form>",
		FormCSS:                    "body{}",
		Recipient:                  "a@example.com,b@example.com",
		Signable:                   true,
		SignatureConfirmationLabel: "I agree",
		SubscriptionListID:         "list-1",
		Type:                       "marketing_form",
		Active:                     true,
		SubmissionCount:            5,
	}
	if _, err := newTestFormsClient(t, srv).CreateForm(context.Background(), req); err != nil {
		t.Fatalf("CreateForm() error: %v", err)
	}

	wantStrings := map[string]string{
		"title":                        "Intake Form",
		"description":                  "Patient intake",
		"form_html":                    "<form></form>",
		"form_css":                     "body{}",
		"recipient":                    "a@example.com,b@example.com",
		"signature_confirmation_label": "I agree",
		"subscription_list_id":         "list-1",
		"type":                         "marketing_form",
	}
	for key, val := range wantStrings {
		if got, _ := gotBody[key].(string); got != val {
			t.Errorf("body[%q] = %v, want %q", key, gotBody[key], val)
		}
	}
	if got, _ := gotBody["customer_id"].(float64); got != 42 {
		t.Errorf("body[customer_id] = %v, want 42", gotBody["customer_id"])
	}
	if got, _ := gotBody["version"].(float64); got != 3 {
		t.Errorf("body[version] = %v, want 3", gotBody["version"])
	}
	if got, _ := gotBody["submission_count"].(float64); got != 5 {
		t.Errorf("body[submission_count] = %v, want 5", gotBody["submission_count"])
	}
	if got, _ := gotBody["signable"].(bool); !got {
		t.Errorf("body[signable] = %v, want true", gotBody["signable"])
	}
	if got, _ := gotBody["active"].(bool); !got {
		t.Errorf("body[active] = %v, want true", gotBody["active"])
	}
	// form_json must be a real JSON object on the wire, not a string.
	fj, ok := gotBody["form_json"].(map[string]any)
	if !ok {
		t.Fatalf("body[form_json] is %T, want JSON object", gotBody["form_json"])
	}
	if _, ok := fj["fields"]; !ok {
		t.Errorf("body[form_json] = %v, want fields key", fj)
	}
}

func TestCreateForm_VersionDefaultsToOne(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		respondJSON(w, http.StatusOK, `{"id":"x"}`)
	}))
	defer srv.Close()

	req := validCreateFormRequest() // Version is the zero value
	if _, err := newTestFormsClient(t, srv).CreateForm(context.Background(), req); err != nil {
		t.Fatalf("CreateForm() error: %v", err)
	}
	if got, _ := gotBody["version"].(float64); got != 1 {
		t.Errorf("body[version] = %v, want 1 (defaulted from 0)", gotBody["version"])
	}
	if req.Version != 0 {
		t.Errorf("caller's req.Version = %d, want 0 (must not be mutated)", req.Version)
	}
}

func TestCreateForm_VersionPreservedWhenSet(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		respondJSON(w, http.StatusOK, `{"id":"x"}`)
	}))
	defer srv.Close()

	req := validCreateFormRequest()
	req.Version = 7
	if _, err := newTestFormsClient(t, srv).CreateForm(context.Background(), req); err != nil {
		t.Fatalf("CreateForm() error: %v", err)
	}
	if got, _ := gotBody["version"].(float64); got != 7 {
		t.Errorf("body[version] = %v, want 7", gotBody["version"])
	}
}

func TestCreateForm_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateFormRequest
		wantErr string
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: "paubox: CreateForm: request must not be nil",
		},
		{
			name: "empty title",
			req: &CreateFormRequest{
				CustomerID: 42,
				FormJSON:   json.RawMessage(crudFormJSON),
			},
			wantErr: "paubox: CreateForm: title must not be empty",
		},
		{
			name: "whitespace title",
			req: &CreateFormRequest{
				Title:      "   ",
				CustomerID: 42,
				FormJSON:   json.RawMessage(crudFormJSON),
			},
			wantErr: "paubox: CreateForm: title must not be empty",
		},
		{
			name: "zero customer_id",
			req: &CreateFormRequest{
				Title:    "t",
				FormJSON: json.RawMessage(crudFormJSON),
			},
			wantErr: "paubox: CreateForm: customer_id must be greater than zero",
		},
		{
			name: "negative customer_id",
			req: &CreateFormRequest{
				Title:      "t",
				CustomerID: -1,
				FormJSON:   json.RawMessage(crudFormJSON),
			},
			wantErr: "paubox: CreateForm: customer_id must be greater than zero",
		},
		{
			name: "empty form_json",
			req: &CreateFormRequest{
				Title:      "t",
				CustomerID: 42,
			},
			wantErr: "paubox: CreateForm: form_json must not be empty",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("validation failure must not reach the server")
	}))
	defer srv.Close()
	c := newTestFormsClient(t, srv)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.CreateForm(context.Background(), tt.req)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetForm
// ---------------------------------------------------------------------------

func TestGetForm_HappyPath_UnwrapsDataEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"data":`+crudFormBody+`}`)
	}))
	defer srv.Close()

	form, err := newTestFormsClient(t, srv).GetForm(context.Background(), "f-123")
	if err != nil {
		t.Fatalf("GetForm() error: %v", err)
	}
	if form.ID != "f-123" {
		t.Errorf("ID = %q, want f-123", form.ID)
	}
	if form.Title != "Intake Form" {
		t.Errorf("Title = %q", form.Title)
	}
	if form.Description == nil || *form.Description != "Patient intake" {
		t.Errorf("Description = %v, want Patient intake", form.Description)
	}
	if form.FormHTML != nil {
		t.Errorf("FormHTML = %v, want nil", form.FormHTML)
	}
	if string(form.FormJSON) != `{"fields":[]}` {
		t.Errorf("FormJSON = %s", form.FormJSON)
	}
	if form.VanityURL == nil || *form.VanityURL != "intake" {
		t.Errorf("VanityURL = %v, want intake", form.VanityURL)
	}
	if form.Version != 2 || !form.Active || form.CustomerID != 42 {
		t.Errorf("Version/Active/CustomerID = %d/%v/%d", form.Version, form.Active, form.CustomerID)
	}
	wantCreated := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if !form.CreatedAt.Equal(wantCreated) {
		t.Errorf("CreatedAt = %v, want %v", form.CreatedAt, wantCreated)
	}
	if form.Recipient == nil || *form.Recipient != "a@example.com,b@example.com" {
		t.Errorf("Recipient = %v", form.Recipient)
	}
	if form.SubmissionCount != 7 {
		t.Errorf("SubmissionCount = %d, want 7", form.SubmissionCount)
	}
	if form.Type == nil || *form.Type != "marketing_form" {
		t.Errorf("Type = %v, want marketing_form", form.Type)
	}
	if form.Deleted || form.Archived {
		t.Errorf("Deleted/Archived = %v/%v, want false/false", form.Deleted, form.Archived)
	}
}

func TestGetForm_MethodAndPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"data":`+crudFormBody+`}`)
	}))
	defer srv.Close()

	_, _ = newTestFormsClient(t, srv).GetForm(context.Background(), "f-123")
	if gotMethod != http.MethodGet || gotPath != "/api/forms/f-123" {
		t.Errorf("request = %s %s, want GET /api/forms/f-123", gotMethod, gotPath)
	}
}

func TestGetForm_TrimsID(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"data":`+crudFormBody+`}`)
	}))
	defer srv.Close()

	_, _ = newTestFormsClient(t, srv).GetForm(context.Background(), "  f-123  ")
	if gotPath != "/api/forms/f-123" {
		t.Errorf("path = %q, want /api/forms/f-123", gotPath)
	}
}

func TestGetForm_Validation(t *testing.T) {
	c, _ := NewForms("key")
	for _, id := range []string{"", "   \t"} {
		_, err := c.GetForm(context.Background(), id)
		if err == nil {
			t.Fatalf("GetForm(%q) = nil error, want validation error", id)
		}
		const want = "paubox: GetForm: id must not be empty"
		if err.Error() != want {
			t.Errorf("GetForm(%q) error = %q, want %q", id, err.Error(), want)
		}
	}
}

// ---------------------------------------------------------------------------
// UpdateForm
// ---------------------------------------------------------------------------

func TestUpdateForm_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"detail":"Form updated successfully","form_id":"f-123"}`)
	}))
	defer srv.Close()

	resp, err := newTestFormsClient(t, srv).UpdateForm(context.Background(), "f-123", &UpdateFormRequest{
		Title: Ptr("Renamed"),
	})
	if err != nil {
		t.Fatalf("UpdateForm() error: %v", err)
	}
	if resp.Detail != "Form updated successfully" {
		t.Errorf("Detail = %q", resp.Detail)
	}
	if resp.FormID != "f-123" {
		t.Errorf("FormID = %q, want f-123", resp.FormID)
	}
}

func TestUpdateForm_MethodAndPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"detail":"ok","form_id":"f-123"}`)
	}))
	defer srv.Close()

	_, _ = newTestFormsClient(t, srv).UpdateForm(context.Background(), "f-123", &UpdateFormRequest{Title: Ptr("t")})
	if gotMethod != http.MethodPut || gotPath != "/api/forms/f-123" {
		t.Errorf("request = %s %s, want PUT /api/forms/f-123", gotMethod, gotPath)
	}
}

func TestUpdateForm_RequestBody_OmitsUnsetFields(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		respondJSON(w, http.StatusOK, `{"detail":"ok","form_id":"f-123"}`)
	}))
	defer srv.Close()

	_, err := newTestFormsClient(t, srv).UpdateForm(context.Background(), "f-123", &UpdateFormRequest{
		Title:  Ptr("Renamed"),
		Active: Ptr(false),
	})
	if err != nil {
		t.Fatalf("UpdateForm() error: %v", err)
	}

	if got, _ := gotBody["title"].(string); got != "Renamed" {
		t.Errorf("body[title] = %v, want Renamed", gotBody["title"])
	}
	if got, ok := gotBody["active"].(bool); !ok || got {
		t.Errorf("body[active] = %v, want false", gotBody["active"])
	}
	// PATCH semantics: unset fields must be absent, not null — a null would
	// still be omitted server-side, but absence proves omitempty works.
	for _, key := range []string{"description", "form_json", "vanity_url", "recipient", "subscription_list_id"} {
		if _, present := gotBody[key]; present {
			t.Errorf("body contains %q, want it omitted", key)
		}
	}
	if len(gotBody) != 2 {
		t.Errorf("body has %d keys, want 2: %v", len(gotBody), gotBody)
	}
}

func TestUpdateForm_RequestBody_AllFields(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		respondJSON(w, http.StatusOK, `{"detail":"ok","form_id":"f-123"}`)
	}))
	defer srv.Close()

	_, err := newTestFormsClient(t, srv).UpdateForm(context.Background(), "f-123", &UpdateFormRequest{
		Title:              Ptr("Renamed"),
		Description:        Ptr("New description"),
		FormJSON:           json.RawMessage(crudFormJSON),
		VanityURL:          Ptr("renamed"),
		Recipient:          Ptr("c@example.com"),
		Active:             Ptr(true),
		SubscriptionListID: Ptr("list-9"),
	})
	if err != nil {
		t.Fatalf("UpdateForm() error: %v", err)
	}

	wantStrings := map[string]string{
		"title":                "Renamed",
		"description":          "New description",
		"vanity_url":           "renamed",
		"recipient":            "c@example.com",
		"subscription_list_id": "list-9",
	}
	for key, val := range wantStrings {
		if got, _ := gotBody[key].(string); got != val {
			t.Errorf("body[%q] = %v, want %q", key, gotBody[key], val)
		}
	}
	if got, _ := gotBody["active"].(bool); !got {
		t.Errorf("body[active] = %v, want true", gotBody["active"])
	}
	if _, ok := gotBody["form_json"].(map[string]any); !ok {
		t.Errorf("body[form_json] is %T, want JSON object", gotBody["form_json"])
	}
}

func TestUpdateForm_Validation(t *testing.T) {
	c, _ := NewForms("key")

	_, err := c.UpdateForm(context.Background(), "  ", &UpdateFormRequest{Title: Ptr("t")})
	if err == nil || err.Error() != "paubox: UpdateForm: id must not be empty" {
		t.Errorf("empty id error = %v, want paubox: UpdateForm: id must not be empty", err)
	}

	_, err = c.UpdateForm(context.Background(), "f-123", nil)
	if err == nil || err.Error() != "paubox: UpdateForm: request must not be nil" {
		t.Errorf("nil request error = %v, want paubox: UpdateForm: request must not be nil", err)
	}
}

// ---------------------------------------------------------------------------
// Cross-endpoint: Authorization header
// ---------------------------------------------------------------------------

func TestFormsCRUD_BearerAuthHeader(t *testing.T) {
	calls := map[string]func(c *FormsClient, ctx context.Context) error{
		"ListForms": func(c *FormsClient, ctx context.Context) error {
			_, err := c.ListForms(ctx, nil)
			return err
		},
		"CreateForm": func(c *FormsClient, ctx context.Context) error {
			_, err := c.CreateForm(ctx, validCreateFormRequest())
			return err
		},
		"GetForm": func(c *FormsClient, ctx context.Context) error {
			_, err := c.GetForm(ctx, "f-123")
			return err
		},
		"UpdateForm": func(c *FormsClient, ctx context.Context) error {
			_, err := c.UpdateForm(ctx, "f-123", &UpdateFormRequest{Title: Ptr("t")})
			return err
		},
	}

	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			var gotAuth string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotAuth = r.Header.Get("Authorization")
				respondJSON(w, http.StatusOK, `{"data":`+crudFormBody+`,"results":[],"page_info":{},"id":"x","detail":"ok","form_id":"f-123"}`)
			}))
			defer srv.Close()

			if err := call(newTestFormsClient(t, srv), context.Background()); err != nil {
				t.Fatalf("%s error: %v", name, err)
			}
			if gotAuth != "Bearer test-key" {
				t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-key")
			}
			if strings.Contains(gotAuth, "Token token=") {
				t.Errorf("Forms request must not use the Email API auth format, got %q", gotAuth)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Cross-endpoint: keyless-client rejection
// ---------------------------------------------------------------------------

func TestFormsCRUD_KeylessClientRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("keyless client must fail fast without an HTTP request")
	}))
	defer srv.Close()
	c := newTestPublicFormsClient(t, srv)

	tests := []struct {
		method string
		call   func() error
	}{
		{"ListForms", func() error { _, err := c.ListForms(context.Background(), nil); return err }},
		{"CreateForm", func() error { _, err := c.CreateForm(context.Background(), validCreateFormRequest()); return err }},
		{"GetForm", func() error { _, err := c.GetForm(context.Background(), "f-123"); return err }},
		{"UpdateForm", func() error {
			_, err := c.UpdateForm(context.Background(), "f-123", &UpdateFormRequest{Title: Ptr("t")})
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			err := tt.call()
			if err == nil {
				t.Fatal("expected error from keyless client, got nil")
			}
			want := "paubox: " + tt.method + ": an API key with the forms scope is required"
			if err.Error() != want {
				t.Errorf("error = %q, want %q", err.Error(), want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Cross-endpoint: 400/401/404 error mapping
// ---------------------------------------------------------------------------

func TestFormsCRUD_ErrorMapping(t *testing.T) {
	endpoints := []struct {
		name string
		call func(c *FormsClient) error
	}{
		{"ListForms", func(c *FormsClient) error {
			_, err := c.ListForms(context.Background(), nil)
			return err
		}},
		{"CreateForm", func(c *FormsClient) error {
			_, err := c.CreateForm(context.Background(), validCreateFormRequest())
			return err
		}},
		{"GetForm", func(c *FormsClient) error {
			_, err := c.GetForm(context.Background(), "f-123")
			return err
		}},
		{"UpdateForm", func(c *FormsClient) error {
			_, err := c.UpdateForm(context.Background(), "f-123", &UpdateFormRequest{Title: Ptr("t")})
			return err
		}},
	}

	statuses := []struct {
		code     int
		sentinel *PauboxError
	}{
		{http.StatusBadRequest, ErrBadRequest},
		{http.StatusUnauthorized, ErrUnauthorized},
		{http.StatusNotFound, ErrNotFound},
	}

	for _, ep := range endpoints {
		for _, st := range statuses {
			t.Run(ep.name+"/"+http.StatusText(st.code), func(t *testing.T) {
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					respondJSON(w, st.code, `{"message":"boom"}`)
				}))
				defer srv.Close()

				err := ep.call(newTestFormsClient(t, srv))
				if err == nil {
					t.Fatalf("expected error for HTTP %d, got nil", st.code)
				}
				if !errors.Is(err, st.sentinel) {
					t.Errorf("errors.Is(err, sentinel %d) = false; err = %v", st.code, err)
				}
				var apiErr *PauboxError
				if !errors.As(err, &apiErr) {
					t.Fatalf("error is %T, want *PauboxError", err)
				}
				if apiErr.StatusCode != st.code {
					t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, st.code)
				}
				if apiErr.Title != "boom" {
					t.Errorf("Title = %q, want boom", apiErr.Title)
				}
			})
		}
	}
}

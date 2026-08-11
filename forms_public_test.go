package paubox

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// publicFormJSON is a full Form body as serialized by the Forms service's
// get_public_form handler (a bare Form object, no wrapper).
const publicFormJSON = `{
	"id": "11111111-2222-3333-4444-555555555555",
	"title": "Patient Intake",
	"description": "New patient intake form",
	"form_html": "<form></form>",
	"form_json": {"body": [{"properties": {"field_name": "first_name"}}]},
	"form_css": null,
	"vanity_url": "patient-intake",
	"version": 2,
	"active": true,
	"customer_id": 42,
	"old_form_id": null,
	"created_at": "2025-06-01T12:00:00Z",
	"updated_at": "2025-06-02T08:30:00Z",
	"recipient": "clinic@example.com,admin@example.com",
	"signable": true,
	"signature_confirmation_label": "I agree",
	"submission_count": 7,
	"type": null,
	"subscription_list_id": null,
	"deleted": false,
	"archived": false
}`

// ---------------------------------------------------------------------------
// GetPublicForm
// ---------------------------------------------------------------------------

func TestGetPublicForm_HappyPath(t *testing.T) {
	var gotMethod, gotPath string
	authPresent := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_, authPresent = r.Header["Authorization"]
		respondJSON(w, http.StatusOK, publicFormJSON)
	}))
	defer srv.Close()

	c := newTestPublicFormsClient(t, srv)
	form, err := c.GetPublicForm(context.Background(), "11111111-2222-3333-4444-555555555555")
	if err != nil {
		t.Fatalf("GetPublicForm() error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if want := "/public/form_data/11111111-2222-3333-4444-555555555555"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	if authPresent {
		t.Error("keyless client must not send an Authorization header")
	}

	if form.ID != "11111111-2222-3333-4444-555555555555" {
		t.Errorf("ID = %q", form.ID)
	}
	if form.Title != "Patient Intake" {
		t.Errorf("Title = %q, want %q", form.Title, "Patient Intake")
	}
	if form.Description == nil || *form.Description != "New patient intake form" {
		t.Errorf("Description = %v, want %q", form.Description, "New patient intake form")
	}
	if form.FormCSS != nil {
		t.Errorf("FormCSS = %v, want nil", form.FormCSS)
	}
	if form.VanityURL == nil || *form.VanityURL != "patient-intake" {
		t.Errorf("VanityURL = %v, want %q", form.VanityURL, "patient-intake")
	}
	if form.Version != 2 {
		t.Errorf("Version = %d, want 2", form.Version)
	}
	if !form.Active {
		t.Error("Active = false, want true")
	}
	if form.CustomerID != 42 {
		t.Errorf("CustomerID = %d, want 42", form.CustomerID)
	}
	if form.Recipient == nil || *form.Recipient != "clinic@example.com,admin@example.com" {
		t.Errorf("Recipient = %v", form.Recipient)
	}
	if !form.Signable {
		t.Error("Signable = false, want true")
	}
	if form.SubmissionCount != 7 {
		t.Errorf("SubmissionCount = %d, want 7", form.SubmissionCount)
	}
	if form.Deleted || form.Archived {
		t.Errorf("Deleted = %v, Archived = %v, want false/false", form.Deleted, form.Archived)
	}
	if len(form.FormJSON) == 0 || !strings.Contains(string(form.FormJSON), "first_name") {
		t.Errorf("FormJSON = %q, want form-builder JSON preserved", form.FormJSON)
	}
	if form.CreatedAt.IsZero() || form.UpdatedAt.IsZero() {
		t.Error("CreatedAt/UpdatedAt should be parsed, got zero values")
	}
}

func TestGetPublicForm_SendsBearerWhenKeyed(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		respondJSON(w, http.StatusOK, publicFormJSON)
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	if _, err := c.GetPublicForm(context.Background(), "form-1"); err != nil {
		t.Fatalf("GetPublicForm() error: %v", err)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-key")
	}
	if strings.Contains(gotAuth, "Token token=") {
		t.Errorf("Forms client must not use the Email API auth format, got %q", gotAuth)
	}
}

func TestGetPublicForm_EscapesFormID(t *testing.T) {
	var gotEscapedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotEscapedPath = r.URL.EscapedPath()
		respondJSON(w, http.StatusOK, publicFormJSON)
	}))
	defer srv.Close()

	c := newTestPublicFormsClient(t, srv)
	if _, err := c.GetPublicForm(context.Background(), "id with spaces"); err != nil {
		t.Fatalf("GetPublicForm() error: %v", err)
	}
	if want := "/public/form_data/id%20with%20spaces"; gotEscapedPath != want {
		t.Errorf("escaped path = %q, want %q", gotEscapedPath, want)
	}
}

func TestGetPublicForm_Validation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("no HTTP request should be made when validation fails")
	}))
	defer srv.Close()

	c := newTestPublicFormsClient(t, srv)
	const want = "paubox: GetPublicForm: formID must not be empty"

	for _, formID := range []string{"", "   ", "\t\n"} {
		form, err := c.GetPublicForm(context.Background(), formID)
		if err == nil {
			t.Fatalf("GetPublicForm(%q) expected error, got nil", formID)
		}
		if form != nil {
			t.Errorf("GetPublicForm(%q) form = %+v, want nil", formID, form)
		}
		if err.Error() != want {
			t.Errorf("GetPublicForm(%q) error = %q, want %q", formID, err.Error(), want)
		}
	}
}

func TestGetPublicForm_ErrorMapping(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		sentinel error
	}{
		// The service 404s an inactive/archived/deleted form with an empty
		// message; the parser falls back to status text.
		{"not found (inactive form)", http.StatusNotFound, `{"message":""}`, ErrNotFound},
		{"not found with message", http.StatusNotFound, `{"message":"Form not found"}`, ErrNotFound},
		{"bad request", http.StatusBadRequest, `{"message":"Bad request"}`, ErrBadRequest},
		{"unauthorized", http.StatusUnauthorized, `{"message":"Unauthorized"}`, ErrUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("X-Request-Id", "req-public-1")
				respondJSON(w, tt.status, tt.body)
			}))
			defer srv.Close()

			c := newTestPublicFormsClient(t, srv)
			form, err := c.GetPublicForm(context.Background(), "missing-form")
			if err == nil {
				t.Fatalf("expected error for %d response, got nil", tt.status)
			}
			if form != nil {
				t.Errorf("form = %+v, want nil on error", form)
			}
			if !errors.Is(err, tt.sentinel) {
				t.Errorf("errors.Is(err, sentinel %d) = false; err = %v", tt.status, err)
			}

			var apiErr *PauboxError
			if !errors.As(err, &apiErr) {
				t.Fatalf("error is %T, want *PauboxError", err)
			}
			if apiErr.StatusCode != tt.status {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.status)
			}
			if apiErr.RequestID != "req-public-1" {
				t.Errorf("RequestID = %q, want %q", apiErr.RequestID, "req-public-1")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// SubmitForm
// ---------------------------------------------------------------------------

func TestSubmitForm_HappyPath(t *testing.T) {
	var gotMethod, gotPath, gotContentType string
	var gotBody []byte
	authPresent := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		_, authPresent = r.Header["Authorization"]
		gotBody, _ = io.ReadAll(r.Body)
		// The service responds 201 Created with an empty body.
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := newTestPublicFormsClient(t, srv)
	err := c.SubmitForm(context.Background(), "form-abc", &SubmitFormRequest{
		FormData: map[string]any{
			"first_name":    "Alice",
			"email_address": "alice@example.com",
		},
	})
	if err != nil {
		t.Fatalf("SubmitForm() error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if want := "/api/forms/form-abc/submissions"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotContentType)
	}
	if authPresent {
		t.Error("keyless client must not send an Authorization header")
	}

	var wire struct {
		FormData map[string]any `json:"form_data"`
	}
	if err := json.Unmarshal(gotBody, &wire); err != nil {
		t.Fatalf("request body is not valid JSON: %v", err)
	}
	if wire.FormData["first_name"] != "Alice" {
		t.Errorf("form_data.first_name = %v, want %q", wire.FormData["first_name"], "Alice")
	}
	if wire.FormData["email_address"] != "alice@example.com" {
		t.Errorf("form_data.email_address = %v", wire.FormData["email_address"])
	}

	// form_data must travel as a real JSON object, not a JSON-encoded string
	// (the Email API's template_values footgun does NOT apply here).
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(gotBody, &raw); err != nil {
		t.Fatalf("request body is not a JSON object: %v", err)
	}
	if fd, ok := raw["form_data"]; !ok || !strings.HasPrefix(strings.TrimSpace(string(fd)), "{") {
		t.Errorf("form_data = %s, want a JSON object", raw["form_data"])
	}
	// attachments must be omitted entirely when there are none.
	if _, ok := raw["attachments"]; ok {
		t.Errorf("attachments key present in body %s, want omitted when none", gotBody)
	}
}

func TestSubmitForm_WithAttachments(t *testing.T) {
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	pdfBytes := []byte("%PDF-1.4 fake insurance card")
	pngBytes := []byte{0x89, 0x50, 0x4E, 0x47}

	c := newTestPublicFormsClient(t, srv)
	err := c.SubmitForm(context.Background(), "form-abc", &SubmitFormRequest{
		FormData: map[string]any{"first_name": "Bob"},
		Attachments: []FormSubmissionAttachment{
			{Name: "insurance.pdf", Content: pdfBytes},
			{Name: "photo.png", Content: pngBytes},
		},
	})
	if err != nil {
		t.Fatalf("SubmitForm() error: %v", err)
	}

	var wire struct {
		FormData    map[string]any `json:"form_data"`
		Attachments []struct {
			Name    string `json:"name"`
			Content string `json:"content"`
		} `json:"attachments"`
	}
	if err := json.Unmarshal(gotBody, &wire); err != nil {
		t.Fatalf("request body is not valid JSON: %v", err)
	}
	if len(wire.Attachments) != 2 {
		t.Fatalf("attachments length = %d, want 2", len(wire.Attachments))
	}
	if wire.Attachments[0].Name != "insurance.pdf" {
		t.Errorf("attachments[0].name = %q, want %q", wire.Attachments[0].Name, "insurance.pdf")
	}
	if want := base64.RawStdEncoding.EncodeToString(pdfBytes); wire.Attachments[0].Content != want {
		t.Errorf("attachments[0].content = %q, want unpadded base64 %q", wire.Attachments[0].Content, want)
	}
	if strings.Contains(wire.Attachments[0].Content, "=") {
		t.Errorf("attachments[0].content %q contains '=' padding; the Forms service requires unpadded base64", wire.Attachments[0].Content)
	}
	if wire.Attachments[1].Name != "photo.png" {
		t.Errorf("attachments[1].name = %q, want %q", wire.Attachments[1].Name, "photo.png")
	}
	if want := base64.RawStdEncoding.EncodeToString(pngBytes); wire.Attachments[1].Content != want {
		t.Errorf("attachments[1].content = %q, want unpadded base64 %q", wire.Attachments[1].Content, want)
	}
	if strings.Contains(wire.Attachments[1].Content, "=") {
		t.Errorf("attachments[1].content %q contains '=' padding; the Forms service requires unpadded base64", wire.Attachments[1].Content)
	}
}

func TestSubmitForm_SendsBearerWhenKeyed(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	err := c.SubmitForm(context.Background(), "form-abc", &SubmitFormRequest{
		FormData: map[string]any{"a": "b"},
	})
	if err != nil {
		t.Fatalf("SubmitForm() error: %v", err)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-key")
	}
	if strings.Contains(gotAuth, "Token token=") {
		t.Errorf("Forms client must not use the Email API auth format, got %q", gotAuth)
	}
}

func TestSubmitForm_EscapesFormID(t *testing.T) {
	var gotEscapedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotEscapedPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := newTestPublicFormsClient(t, srv)
	err := c.SubmitForm(context.Background(), "id with spaces", &SubmitFormRequest{
		FormData: map[string]any{"a": "b"},
	})
	if err != nil {
		t.Fatalf("SubmitForm() error: %v", err)
	}
	if want := "/api/forms/id%20with%20spaces/submissions"; gotEscapedPath != want {
		t.Errorf("escaped path = %q, want %q", gotEscapedPath, want)
	}
}

func TestSubmitForm_Validation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("no HTTP request should be made when validation fails")
	}))
	defer srv.Close()

	c := newTestPublicFormsClient(t, srv)
	validData := map[string]any{"first_name": "Alice"}

	tests := []struct {
		name    string
		formID  string
		req     *SubmitFormRequest
		wantErr string
	}{
		{
			name:    "empty formID",
			formID:  "",
			req:     &SubmitFormRequest{FormData: validData},
			wantErr: "paubox: SubmitForm: formID must not be empty",
		},
		{
			name:    "whitespace formID",
			formID:  "   ",
			req:     &SubmitFormRequest{FormData: validData},
			wantErr: "paubox: SubmitForm: formID must not be empty",
		},
		{
			name:    "nil request",
			formID:  "form-abc",
			req:     nil,
			wantErr: "paubox: SubmitForm: request must not be nil",
		},
		{
			name:    "nil form data",
			formID:  "form-abc",
			req:     &SubmitFormRequest{},
			wantErr: "paubox: SubmitForm: form data must not be empty",
		},
		{
			name:    "empty form data",
			formID:  "form-abc",
			req:     &SubmitFormRequest{FormData: map[string]any{}},
			wantErr: "paubox: SubmitForm: form data must not be empty",
		},
		{
			name:   "attachment missing name",
			formID: "form-abc",
			req: &SubmitFormRequest{
				FormData:    validData,
				Attachments: []FormSubmissionAttachment{{Name: "  ", Content: []byte("x")}},
			},
			wantErr: "paubox: SubmitForm: attachment[0]: name must not be empty",
		},
		{
			name:   "attachment missing content",
			formID: "form-abc",
			req: &SubmitFormRequest{
				FormData:    validData,
				Attachments: []FormSubmissionAttachment{{Name: "a.pdf"}},
			},
			wantErr: `paubox: SubmitForm: attachment "a.pdf": content must not be empty`,
		},
		{
			name:   "second attachment invalid",
			formID: "form-abc",
			req: &SubmitFormRequest{
				FormData: validData,
				Attachments: []FormSubmissionAttachment{
					{Name: "ok.pdf", Content: []byte("x")},
					{Name: "", Content: []byte("y")},
				},
			},
			wantErr: "paubox: SubmitForm: attachment[1]: name must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.SubmitForm(context.Background(), tt.formID, tt.req)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestSubmitForm_ErrorMapping(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		sentinel error
	}{
		{"bad request (missing form_data)", http.StatusBadRequest, `{"message":"Bad Request"}`, ErrBadRequest},
		{"unauthorized", http.StatusUnauthorized, `{"message":"Unauthorized"}`, ErrUnauthorized},
		{"form not found", http.StatusNotFound, `{"message":"Form not found"}`, ErrNotFound},
		{"server error, empty body", http.StatusInternalServerError, ``, ErrServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				respondJSON(w, tt.status, tt.body)
			}))
			defer srv.Close()

			c := newTestPublicFormsClient(t, srv)
			err := c.SubmitForm(context.Background(), "form-abc", &SubmitFormRequest{
				FormData: map[string]any{"a": "b"},
			})
			if err == nil {
				t.Fatalf("expected error for %d response, got nil", tt.status)
			}
			if !errors.Is(err, tt.sentinel) {
				t.Errorf("errors.Is(err, sentinel %d) = false; err = %v", tt.status, err)
			}

			var apiErr *PauboxError
			if !errors.As(err, &apiErr) {
				t.Fatalf("error is %T, want *PauboxError", err)
			}
			if apiErr.StatusCode != tt.status {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.status)
			}
			if tt.body != "" {
				var msg struct {
					Message string `json:"message"`
				}
				if json.Unmarshal([]byte(tt.body), &msg) == nil && msg.Message != "" && apiErr.Title != msg.Message {
					t.Errorf("Title = %q, want %q", apiErr.Title, msg.Message)
				}
			}
		})
	}
}

// TestSubmitForm_SucceedsOn201EmptyBody pins the success contract: the
// service replies 201 Created with no body and SubmitForm must return nil
// without attempting to decode anything.
func TestSubmitForm_SucceedsOn201EmptyBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated) // no body at all
	}))
	defer srv.Close()

	c := newTestPublicFormsClient(t, srv)
	err := c.SubmitForm(context.Background(), "form-abc", &SubmitFormRequest{
		FormData: map[string]any{"a": "b"},
	})
	if err != nil {
		t.Fatalf("SubmitForm() = %v, want nil on 201 empty body", err)
	}
}

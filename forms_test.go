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

// newTestFormsClient wires a FormsClient to the given httptest.Server.
func newTestFormsClient(t *testing.T, srv *httptest.Server) *FormsClient {
	t.Helper()
	fc, err := NewFormsClient(
		WithFormsBaseURL(srv.URL),
		WithFormsTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("NewFormsClient() error: %v", err)
	}
	return fc
}

// ---------------------------------------------------------------------------
// NewFormsClient
// ---------------------------------------------------------------------------

func TestNewFormsClient_Defaults(t *testing.T) {
	fc, err := NewFormsClient()
	if err != nil {
		t.Fatalf("NewFormsClient() error: %v", err)
	}
	if fc.baseURL != defaultFormsBaseURL {
		t.Errorf("baseURL = %q, want %q", fc.baseURL, defaultFormsBaseURL)
	}
	if fc.userAgent != defaultUserAgent {
		t.Errorf("userAgent = %q, want %q", fc.userAgent, defaultUserAgent)
	}
	if fc.retry.MaxAttempts != defaultRetryConfig.MaxAttempts {
		t.Errorf("MaxAttempts = %d, want %d", fc.retry.MaxAttempts, defaultRetryConfig.MaxAttempts)
	}
}

func TestNewFormsClient_WithOptions(t *testing.T) {
	fc, err := NewFormsClient(
		WithFormsBaseURL("https://staging.example.com"),
		WithFormsTimeout(10*time.Second),
		WithFormsUserAgent("myapp/1.0"),
		WithFormsRetry(RetryConfig{MaxAttempts: 5, WaitMin: 50 * time.Millisecond, WaitMax: 1 * time.Second}),
	)
	if err != nil {
		t.Fatalf("NewFormsClient() error: %v", err)
	}
	if fc.baseURL != "https://staging.example.com" {
		t.Errorf("baseURL = %q", fc.baseURL)
	}
	if !strings.Contains(fc.userAgent, "myapp/1.0") {
		t.Errorf("userAgent %q should contain myapp/1.0", fc.userAgent)
	}
	if !strings.Contains(fc.userAgent, defaultUserAgent) {
		t.Errorf("userAgent %q should contain SDK identifier", fc.userAgent)
	}
	if fc.retry.MaxAttempts != 5 {
		t.Errorf("MaxAttempts = %d, want 5", fc.retry.MaxAttempts)
	}
}

func TestNewFormsClient_WithBaseURL_TrailingSlash(t *testing.T) {
	fc, _ := NewFormsClient(WithFormsBaseURL("https://example.com/"))
	if strings.HasSuffix(fc.baseURL, "/") {
		t.Errorf("baseURL should not have trailing slash, got %q", fc.baseURL)
	}
}

// ---------------------------------------------------------------------------
// GetForm — validation
// ---------------------------------------------------------------------------

func TestGetForm_Validation(t *testing.T) {
	tests := []struct {
		name    string
		formID  string
		wantErr string
	}{
		{"empty", "", "formID must not be empty"},
		{"whitespace", "   ", "formID must not be empty"},
	}

	fc, _ := NewFormsClient()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fc.GetForm(context.Background(), tc.formID)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetForm — HTTP responses
// ---------------------------------------------------------------------------

func TestGetForm_HappyPath(t *testing.T) {
	const body = `{
		"id": "form-uuid-123",
		"title": "Patient Intake Form",
		"description": "Please fill out prior to your appointment.",
		"active": true,
		"signable": false,
		"deleted": false,
		"archived": false,
		"submission_count": 42,
		"form_json": {
			"body": [
				{"type": "Text", "id": "field-1", "properties": {"field_name": "name", "text": "Full Name"}},
				{"type": "Text", "id": "field-2", "properties": {"field_name": "dob",  "text": "Date of Birth"}}
			]
		},
		"created_at": "2024-01-01T00:00:00Z",
		"updated_at": "2024-06-01T12:00:00Z"
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, body)
	}))
	defer srv.Close()

	form, err := newTestFormsClient(t, srv).GetForm(context.Background(), "form-uuid-123")
	if err != nil {
		t.Fatalf("GetForm() error: %v", err)
	}

	if form.ID != "form-uuid-123" {
		t.Errorf("ID = %q, want form-uuid-123", form.ID)
	}
	if form.Title != "Patient Intake Form" {
		t.Errorf("Title = %q", form.Title)
	}
	if !form.Active {
		t.Error("Active = false, want true")
	}
	if form.SubmissionCount != 42 {
		t.Errorf("SubmissionCount = %d, want 42", form.SubmissionCount)
	}
	if form.FormJSON == nil {
		t.Fatal("FormJSON is nil")
	}
	if len(form.FormJSON.Body) != 2 {
		t.Errorf("FormJSON.Body len = %d, want 2", len(form.FormJSON.Body))
	}
	if form.FormJSON.Body[0].Type != "Text" {
		t.Errorf("FormJSON.Body[0].Type = %q, want Text", form.FormJSON.Body[0].Type)
	}
	if form.FormJSON.Body[0].ID != "field-1" {
		t.Errorf("FormJSON.Body[0].ID = %q, want field-1", form.FormJSON.Body[0].ID)
	}
}

func TestGetForm_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusNotFound, `{"errors":[{"code":404,"title":"Not Found","details":"form not found"}]}`)
	}))
	defer srv.Close()

	_, err := newTestFormsClient(t, srv).GetForm(context.Background(), "bad-uuid")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	var apiErr *PauboxError
	if errors.As(err, &apiErr) && apiErr.Details == "" {
		t.Error("Details should be populated from wire error")
	}
}

func TestGetForm_MalformedResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not valid json {{{"))
	}))
	defer srv.Close()

	_, err := newTestFormsClient(t, srv).GetForm(context.Background(), "form-uuid-123")
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
	if !strings.Contains(err.Error(), "decoding response") {
		t.Errorf("error %q should mention decoding", err.Error())
	}
}

func TestGetForm_SendsCorrectMethodAndPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"id":"abc","title":"t"}`)
	}))
	defer srv.Close()

	_, _ = newTestFormsClient(t, srv).GetForm(context.Background(), "abc")

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/public/form_data/abc" {
		t.Errorf("path = %q, want /public/form_data/abc", gotPath)
	}
}

// TestGetForm_NoAuthorizationHeader verifies that FormsClient never sends
// credentials, keeping it safe to call for public forms on behalf of any user.
func TestGetForm_NoAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		respondJSON(w, http.StatusOK, `{"id":"x","title":"t"}`)
	}))
	defer srv.Close()

	_, _ = newTestFormsClient(t, srv).GetForm(context.Background(), "x")
	if gotAuth != "" {
		t.Errorf("Authorization header must be empty, got %q", gotAuth)
	}
}

// ---------------------------------------------------------------------------
// SubmitForm — validation
// ---------------------------------------------------------------------------

func TestSubmitForm_Validation(t *testing.T) {
	tests := []struct {
		name    string
		formID  string
		sub     FormSubmission
		wantErr string
	}{
		{
			name:    "empty formID",
			formID:  "",
			sub:     FormSubmission{FormData: map[string]any{"k": "v"}},
			wantErr: "formID must not be empty",
		},
		{
			name:    "whitespace formID",
			formID:  "   ",
			sub:     FormSubmission{FormData: map[string]any{"k": "v"}},
			wantErr: "formID must not be empty",
		},
		{
			name:    "nil FormData",
			formID:  "form-uuid-123",
			sub:     FormSubmission{FormData: nil},
			wantErr: "FormData must not be nil",
		},
	}

	fc, _ := NewFormsClient()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fc.SubmitForm(context.Background(), tc.formID, tc.sub)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// SubmitForm — HTTP responses
// ---------------------------------------------------------------------------

func TestSubmitForm_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	resp, err := newTestFormsClient(t, srv).SubmitForm(context.Background(), "form-uuid-123", FormSubmission{
		FormData: map[string]any{"name": "Alice", "email": "alice@example.com"},
	})
	if err != nil {
		t.Fatalf("SubmitForm() error: %v", err)
	}
	if resp == nil {
		t.Error("expected non-nil response")
	}
}

func TestSubmitForm_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusNotFound, `{"errors":[{"code":404,"title":"Not Found","details":"form not found"}]}`)
	}))
	defer srv.Close()

	_, err := newTestFormsClient(t, srv).SubmitForm(context.Background(), "bad-uuid", FormSubmission{
		FormData: map[string]any{"k": "v"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSubmitForm_SendsCorrectMethodAndPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	_, _ = newTestFormsClient(t, srv).SubmitForm(context.Background(), "form-uuid-123", FormSubmission{
		FormData: map[string]any{"k": "v"},
	})

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/api/forms/form-uuid-123/submissions" {
		t.Errorf("path = %q, want /api/forms/form-uuid-123/submissions", gotPath)
	}
}

func TestSubmitForm_SendsFormData(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	_, err := newTestFormsClient(t, srv).SubmitForm(context.Background(), "form-uuid-123", FormSubmission{
		FormData: map[string]any{"name": "Alice", "dob": "1990-01-01"},
		Attachments: []FormAttachment{
			{Name: "consent.pdf", Content: "base64data=="},
		},
	})
	if err != nil {
		t.Fatalf("SubmitForm() error: %v", err)
	}

	fd, ok := gotBody["form_data"].(map[string]any)
	if !ok {
		t.Fatalf("form_data missing or wrong type in %v", gotBody)
	}
	if fd["name"] != "Alice" {
		t.Errorf("form_data.name = %v, want Alice", fd["name"])
	}

	atts, ok := gotBody["attachments"].([]any)
	if !ok || len(atts) != 1 {
		t.Fatalf("attachments = %v, want 1 element", gotBody["attachments"])
	}
	att := atts[0].(map[string]any)
	if att["name"] != "consent.pdf" {
		t.Errorf("attachment name = %v, want consent.pdf", att["name"])
	}
}

// TestSubmitForm_NoAuthorizationHeader verifies that FormsClient never
// sends credentials, even for the POST submission endpoint.
func TestSubmitForm_NoAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	_, _ = newTestFormsClient(t, srv).SubmitForm(context.Background(), "form-uuid-123", FormSubmission{
		FormData: map[string]any{"k": "v"},
	})
	if gotAuth != "" {
		t.Errorf("Authorization header must be empty, got %q", gotAuth)
	}
}

// ---------------------------------------------------------------------------
// FormsClient — retry behaviour
// ---------------------------------------------------------------------------

func TestFormsClient_RetryOn5xx_GET(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		respondJSON(w, http.StatusOK, `{"id":"x","title":"t"}`)
	}))
	defer srv.Close()

	fc, _ := NewFormsClient(
		WithFormsBaseURL(srv.URL),
		WithFormsRetry(RetryConfig{MaxAttempts: 3, WaitMin: 1 * time.Millisecond, WaitMax: 5 * time.Millisecond}),
	)

	_, err := fc.GetForm(context.Background(), "x")
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestFormsClient_NoRetryOnPOST_ByDefault(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	fc, _ := NewFormsClient(
		WithFormsBaseURL(srv.URL),
		WithFormsRetry(RetryConfig{
			MaxAttempts:        3,
			WaitMin:            1 * time.Millisecond,
			WaitMax:            5 * time.Millisecond,
			RetryNonIdempotent: false,
		}),
	)

	_, _ = fc.SubmitForm(context.Background(), "form-uuid-123", FormSubmission{
		FormData: map[string]any{"k": "v"},
	})
	if calls != 1 {
		t.Errorf("POST was retried: calls = %d, want 1", calls)
	}
}

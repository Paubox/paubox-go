package paubox

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// ArchiveForm
// ---------------------------------------------------------------------------

func TestArchiveForm_HappyPath(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		respondJSON(w, http.StatusOK, `{"detail":"Form archived."}`)
	}))
	defer srv.Close()

	resp, err := newTestFormsClient(t, srv).ArchiveForm(context.Background(), "form-123")
	if err != nil {
		t.Fatalf("ArchiveForm() error: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/api/forms/form-123/archive" {
		t.Errorf("request = %s %s, want POST /api/forms/form-123/archive", gotMethod, gotPath)
	}
	if len(gotBody) != 0 {
		t.Errorf("request body = %q, want empty", gotBody)
	}
	if resp.Detail != "Form archived." {
		t.Errorf("Detail = %q, want %q", resp.Detail, "Form archived.")
	}
}

func TestArchiveForm_SendsBearerAuth(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		respondJSON(w, http.StatusOK, `{"detail":"Form archived."}`)
	}))
	defer srv.Close()

	if _, err := newTestFormsClient(t, srv).ArchiveForm(context.Background(), "form-123"); err != nil {
		t.Fatalf("ArchiveForm() error: %v", err)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-key")
	}
	if strings.Contains(gotAuth, "Token token=") {
		t.Errorf("Forms request must not use the Email API auth format, got %q", gotAuth)
	}
}

func TestArchiveForm_Validation_EmptyID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("no HTTP request should be made on validation failure")
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	for _, id := range []string{"", "   \t"} {
		_, err := c.ArchiveForm(context.Background(), id)
		if err == nil {
			t.Fatalf("ArchiveForm(%q) expected validation error, got nil", id)
		}
		if !strings.HasPrefix(err.Error(), "paubox: ArchiveForm: ") {
			t.Errorf("error = %q, want prefix %q", err.Error(), "paubox: ArchiveForm: ")
		}
	}
}

func TestArchiveForm_ErrorMapping(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		sentinel error
	}{
		{"bad request", http.StatusBadRequest, `{"message":"bad request"}`, ErrBadRequest},
		{"unauthorized", http.StatusUnauthorized, `{"message":"Invalid API key"}`, ErrUnauthorized},
		{"not found", http.StatusNotFound, `{"message":"Form not found"}`, ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				respondJSON(w, tt.status, tt.body)
			}))
			defer srv.Close()

			_, err := newTestFormsClient(t, srv).ArchiveForm(context.Background(), "form-123")
			if err == nil {
				t.Fatalf("expected error for %d response", tt.status)
			}
			if !errors.Is(err, tt.sentinel) {
				t.Errorf("errors.Is(err, sentinel %d) = false; err = %v", tt.status, err)
			}
		})
	}
}

func TestArchiveForm_KeylessClientRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("no HTTP request should be made by a keyless client")
	}))
	defer srv.Close()

	_, err := newTestPublicFormsClient(t, srv).ArchiveForm(context.Background(), "form-123")
	if err == nil {
		t.Fatal("expected error from keyless client")
	}
	const want = "paubox: ArchiveForm: an API key with the forms scope is required"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

// ---------------------------------------------------------------------------
// UnarchiveForm
// ---------------------------------------------------------------------------

func TestUnarchiveForm_HappyPath(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		respondJSON(w, http.StatusOK, `{"detail":"Form unarchived."}`)
	}))
	defer srv.Close()

	resp, err := newTestFormsClient(t, srv).UnarchiveForm(context.Background(), "form-123")
	if err != nil {
		t.Fatalf("UnarchiveForm() error: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/api/forms/form-123/unarchive" {
		t.Errorf("request = %s %s, want POST /api/forms/form-123/unarchive", gotMethod, gotPath)
	}
	if len(gotBody) != 0 {
		t.Errorf("request body = %q, want empty", gotBody)
	}
	if resp.Detail != "Form unarchived." {
		t.Errorf("Detail = %q, want %q", resp.Detail, "Form unarchived.")
	}
}

func TestUnarchiveForm_SendsBearerAuth(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		respondJSON(w, http.StatusOK, `{"detail":"Form unarchived."}`)
	}))
	defer srv.Close()

	if _, err := newTestFormsClient(t, srv).UnarchiveForm(context.Background(), "form-123"); err != nil {
		t.Fatalf("UnarchiveForm() error: %v", err)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-key")
	}
	if strings.Contains(gotAuth, "Token token=") {
		t.Errorf("Forms request must not use the Email API auth format, got %q", gotAuth)
	}
}

func TestUnarchiveForm_Validation_EmptyID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("no HTTP request should be made on validation failure")
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	for _, id := range []string{"", "  "} {
		_, err := c.UnarchiveForm(context.Background(), id)
		if err == nil {
			t.Fatalf("UnarchiveForm(%q) expected validation error, got nil", id)
		}
		if !strings.HasPrefix(err.Error(), "paubox: UnarchiveForm: ") {
			t.Errorf("error = %q, want prefix %q", err.Error(), "paubox: UnarchiveForm: ")
		}
	}
}

func TestUnarchiveForm_ErrorMapping(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		sentinel error
	}{
		{"bad request", http.StatusBadRequest, `{"message":"bad request"}`, ErrBadRequest},
		{"unauthorized", http.StatusUnauthorized, `{"message":"Invalid API key"}`, ErrUnauthorized},
		{"not found", http.StatusNotFound, `{"message":"Form not found"}`, ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				respondJSON(w, tt.status, tt.body)
			}))
			defer srv.Close()

			_, err := newTestFormsClient(t, srv).UnarchiveForm(context.Background(), "form-123")
			if err == nil {
				t.Fatalf("expected error for %d response", tt.status)
			}
			if !errors.Is(err, tt.sentinel) {
				t.Errorf("errors.Is(err, sentinel %d) = false; err = %v", tt.status, err)
			}
		})
	}
}

func TestUnarchiveForm_KeylessClientRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("no HTTP request should be made by a keyless client")
	}))
	defer srv.Close()

	_, err := newTestPublicFormsClient(t, srv).UnarchiveForm(context.Background(), "form-123")
	if err == nil {
		t.Fatal("expected error from keyless client")
	}
	const want = "paubox: UnarchiveForm: an API key with the forms scope is required"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

// ---------------------------------------------------------------------------
// CopyForm
// ---------------------------------------------------------------------------

func TestCopyForm_HappyPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		// Bare Form object — no {"data": ...} wrapper.
		respondJSON(w, http.StatusOK, `{
			"id": "new-form-id",
			"title": "Intake (copy)",
			"description": "desc",
			"form_html": null,
			"form_json": {"fields":[]},
			"form_css": null,
			"vanity_url": null,
			"version": 2,
			"active": true,
			"customer_id": 42,
			"old_form_id": null,
			"created_at": "2026-08-11T10:00:00Z",
			"updated_at": "2026-08-11T10:00:00Z",
			"recipient": "ops@example.com",
			"signable": false,
			"signature_confirmation_label": null,
			"submission_count": 0,
			"type": null,
			"subscription_list_id": null,
			"deleted": false,
			"archived": false
		}`)
	}))
	defer srv.Close()

	form, err := newTestFormsClient(t, srv).CopyForm(context.Background(), &CopyFormRequest{
		FormID: "orig-form-id",
		Title:  "Intake (copy)",
	})
	if err != nil {
		t.Fatalf("CopyForm() error: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/api/forms/copy" {
		t.Errorf("request = %s %s, want POST /api/forms/copy", gotMethod, gotPath)
	}
	if form.ID != "new-form-id" {
		t.Errorf("ID = %q, want %q", form.ID, "new-form-id")
	}
	if form.Title != "Intake (copy)" {
		t.Errorf("Title = %q, want %q", form.Title, "Intake (copy)")
	}
	if form.Version != 2 || !form.Active || form.CustomerID != 42 {
		t.Errorf("form = %+v", form)
	}
	if form.SubmissionCount != 0 || form.Archived || form.Deleted {
		t.Errorf("copy should start clean, got %+v", form)
	}
	if form.Recipient == nil || *form.Recipient != "ops@example.com" {
		t.Errorf("Recipient = %v, want ops@example.com", form.Recipient)
	}
	if string(form.FormJSON) != `{"fields":[]}` {
		t.Errorf("FormJSON = %s", form.FormJSON)
	}
}

func TestCopyForm_SendsRequestBody(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding request body: %v", err)
		}
		respondJSON(w, http.StatusOK, `{"id":"new-id","title":"Copy"}`)
	}))
	defer srv.Close()

	_, err := newTestFormsClient(t, srv).CopyForm(context.Background(), &CopyFormRequest{
		FormID: "orig-id",
		Title:  "Copy",
	})
	if err != nil {
		t.Fatalf("CopyForm() error: %v", err)
	}
	if gotBody["form_id"] != "orig-id" {
		t.Errorf("form_id = %v, want orig-id", gotBody["form_id"])
	}
	if gotBody["title"] != "Copy" {
		t.Errorf("title = %v, want Copy", gotBody["title"])
	}
	if len(gotBody) != 2 {
		t.Errorf("body has %d keys, want 2 (form_id, title): %v", len(gotBody), gotBody)
	}
}

func TestCopyForm_SendsBearerAuth(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		respondJSON(w, http.StatusOK, `{"id":"x"}`)
	}))
	defer srv.Close()

	if _, err := newTestFormsClient(t, srv).CopyForm(context.Background(), &CopyFormRequest{
		FormID: "orig-id",
		Title:  "Copy",
	}); err != nil {
		t.Fatalf("CopyForm() error: %v", err)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-key")
	}
	if strings.Contains(gotAuth, "Token token=") {
		t.Errorf("Forms request must not use the Email API auth format, got %q", gotAuth)
	}
}

func TestCopyForm_Validation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("no HTTP request should be made on validation failure")
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	tests := []struct {
		name    string
		req     *CopyFormRequest
		wantSub string
	}{
		{"nil request", nil, "request must not be nil"},
		{"empty form_id", &CopyFormRequest{Title: "Copy"}, "form_id must not be empty"},
		{"whitespace form_id", &CopyFormRequest{FormID: "  ", Title: "Copy"}, "form_id must not be empty"},
		{"empty title", &CopyFormRequest{FormID: "orig-id"}, "title must not be empty"},
		{"whitespace title", &CopyFormRequest{FormID: "orig-id", Title: " \t"}, "title must not be empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.CopyForm(context.Background(), tt.req)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !strings.HasPrefix(err.Error(), "paubox: CopyForm: ") {
				t.Errorf("error = %q, want prefix %q", err.Error(), "paubox: CopyForm: ")
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantSub)
			}
		})
	}
}

func TestCopyForm_ErrorMapping(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		sentinel error
	}{
		{"bad request", http.StatusBadRequest, `{"message":"bad request"}`, ErrBadRequest},
		{"unauthorized", http.StatusUnauthorized, `{"message":"Invalid API key"}`, ErrUnauthorized},
		{"not found", http.StatusNotFound, `{"message":"Form not found"}`, ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				respondJSON(w, tt.status, tt.body)
			}))
			defer srv.Close()

			_, err := newTestFormsClient(t, srv).CopyForm(context.Background(), &CopyFormRequest{
				FormID: "orig-id",
				Title:  "Copy",
			})
			if err == nil {
				t.Fatalf("expected error for %d response", tt.status)
			}
			if !errors.Is(err, tt.sentinel) {
				t.Errorf("errors.Is(err, sentinel %d) = false; err = %v", tt.status, err)
			}
		})
	}
}

func TestCopyForm_KeylessClientRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("no HTTP request should be made by a keyless client")
	}))
	defer srv.Close()

	_, err := newTestPublicFormsClient(t, srv).CopyForm(context.Background(), &CopyFormRequest{
		FormID: "orig-id",
		Title:  "Copy",
	})
	if err == nil {
		t.Fatal("expected error from keyless client")
	}
	const want = "paubox: CopyForm: an API key with the forms scope is required"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

// ---------------------------------------------------------------------------
// GetFormStats
// ---------------------------------------------------------------------------

func TestGetFormStats_HappyPath_NilParams(t *testing.T) {
	var gotMethod, gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		respondJSON(w, http.StatusOK, `{"active_form_count":7,"total_submission_count":250,"submissions_last_7_days":31}`)
	}))
	defer srv.Close()

	stats, err := newTestFormsClient(t, srv).GetFormStats(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetFormStats() error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/api/forms/stats" {
		t.Errorf("request = %s %s, want GET /api/forms/stats", gotMethod, gotPath)
	}
	if gotQuery != "" {
		t.Errorf("query = %q, want empty for nil params", gotQuery)
	}
	if stats.ActiveFormCount != 7 {
		t.Errorf("ActiveFormCount = %d, want 7", stats.ActiveFormCount)
	}
	if stats.TotalSubmissionCount != 250 {
		t.Errorf("TotalSubmissionCount = %d, want 250", stats.TotalSubmissionCount)
	}
	if stats.SubmissionsLast7Days != 31 {
		t.Errorf("SubmissionsLast7Days = %d, want 31", stats.SubmissionsLast7Days)
	}
}

func TestGetFormStats_CustomerIDQueryParam(t *testing.T) {
	var gotCustomerID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCustomerID = r.URL.Query().Get("customer_id")
		respondJSON(w, http.StatusOK, `{"active_form_count":0,"total_submission_count":0,"submissions_last_7_days":0}`)
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	if _, err := c.GetFormStats(context.Background(), &FormStatsParams{CustomerID: 42}); err != nil {
		t.Fatalf("GetFormStats() error: %v", err)
	}
	if gotCustomerID != "42" {
		t.Errorf("customer_id = %q, want %q", gotCustomerID, "42")
	}
}

func TestGetFormStats_ZeroCustomerIDOmitted(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		respondJSON(w, http.StatusOK, `{"active_form_count":0,"total_submission_count":0,"submissions_last_7_days":0}`)
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	if _, err := c.GetFormStats(context.Background(), &FormStatsParams{}); err != nil {
		t.Fatalf("GetFormStats() error: %v", err)
	}
	if gotQuery != "" {
		t.Errorf("query = %q, want empty when CustomerID is zero", gotQuery)
	}
}

func TestGetFormStats_SendsBearerAuth(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		respondJSON(w, http.StatusOK, `{"active_form_count":0,"total_submission_count":0,"submissions_last_7_days":0}`)
	}))
	defer srv.Close()

	if _, err := newTestFormsClient(t, srv).GetFormStats(context.Background(), nil); err != nil {
		t.Fatalf("GetFormStats() error: %v", err)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-key")
	}
	if strings.Contains(gotAuth, "Token token=") {
		t.Errorf("Forms request must not use the Email API auth format, got %q", gotAuth)
	}
}

func TestGetFormStats_ErrorMapping(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		sentinel error
	}{
		{"bad request", http.StatusBadRequest, `{"message":"bad request"}`, ErrBadRequest},
		{"unauthorized", http.StatusUnauthorized, `{"message":"Invalid API key"}`, ErrUnauthorized},
		{"not found", http.StatusNotFound, `{"message":"not found"}`, ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				respondJSON(w, tt.status, tt.body)
			}))
			defer srv.Close()

			_, err := newTestFormsClient(t, srv).GetFormStats(context.Background(), nil)
			if err == nil {
				t.Fatalf("expected error for %d response", tt.status)
			}
			if !errors.Is(err, tt.sentinel) {
				t.Errorf("errors.Is(err, sentinel %d) = false; err = %v", tt.status, err)
			}
		})
	}
}

func TestGetFormStats_KeylessClientRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("no HTTP request should be made by a keyless client")
	}))
	defer srv.Close()

	_, err := newTestPublicFormsClient(t, srv).GetFormStats(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error from keyless client")
	}
	const want = "paubox: GetFormStats: an API key with the forms scope is required"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

package paubox

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// ListFormSubmissions
// ---------------------------------------------------------------------------

const listSubmissionsBody = `{
	"data": [
		{
			"id": "sub-1",
			"form_id": "form-1",
			"form_data": "{\"first_name\":\"Alice\"}",
			"storage_type": "S3",
			"storage_url": null,
			"submitter_email": "alice@example.com",
			"recipients": "ops@clinic.example",
			"attachment": null,
			"attachment_name": "consent.pdf",
			"attachment_url": null,
			"attachment_type": null,
			"created_at": "2026-08-01T12:00:00Z"
		},
		{
			"id": "sub-2",
			"form_id": "form-1",
			"form_data": "{}",
			"storage_type": "S3",
			"storage_url": null,
			"submitter_email": null,
			"recipients": null,
			"attachment": null,
			"attachment_name": null,
			"attachment_url": null,
			"attachment_type": null,
			"created_at": "2026-08-02T09:30:00Z"
		}
	],
	"total": 2,
	"page": 1,
	"items": 50
}`

func TestListFormSubmissions_HappyPath(t *testing.T) {
	var gotMethod, gotPath, gotQuery, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		respondJSON(w, http.StatusOK, listSubmissionsBody)
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	resp, err := c.ListFormSubmissions(context.Background(), "form-1", nil)
	if err != nil {
		t.Fatalf("ListFormSubmissions() error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/api/forms/form-1/submissions" {
		t.Errorf("path = %q, want /api/forms/form-1/submissions", gotPath)
	}
	if gotQuery != "" {
		t.Errorf("query = %q, want empty for nil params", gotQuery)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-key")
	}
	if strings.Contains(gotAuth, "Token token=") {
		t.Errorf("Forms request must not use the Email API auth format, got %q", gotAuth)
	}

	if resp.Total != 2 || resp.Page != 1 || resp.Items != 50 {
		t.Errorf("pagination = total %d page %d items %d, want 2/1/50", resp.Total, resp.Page, resp.Items)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("len(Data) = %d, want 2", len(resp.Data))
	}

	first := resp.Data[0]
	if first.ID != "sub-1" || first.FormID != "form-1" {
		t.Errorf("Data[0] = id %q form_id %q", first.ID, first.FormID)
	}
	if first.FormData != `{"first_name":"Alice"}` {
		t.Errorf("Data[0].FormData = %q, want the JSON-encoded string", first.FormData)
	}
	if first.StorageType != "S3" {
		t.Errorf("Data[0].StorageType = %q, want S3", first.StorageType)
	}
	if first.SubmitterEmail == nil || *first.SubmitterEmail != "alice@example.com" {
		t.Errorf("Data[0].SubmitterEmail = %v, want alice@example.com", first.SubmitterEmail)
	}
	if first.AttachmentName == nil || *first.AttachmentName != "consent.pdf" {
		t.Errorf("Data[0].AttachmentName = %v, want consent.pdf", first.AttachmentName)
	}
	if first.CreatedAt.IsZero() {
		t.Error("Data[0].CreatedAt should be parsed, got zero time")
	}

	second := resp.Data[1]
	if second.SubmitterEmail != nil || second.Recipients != nil || second.AttachmentName != nil {
		t.Errorf("Data[1] nullable fields should be nil, got %+v", second)
	}
}

func TestListFormSubmissions_QueryParams(t *testing.T) {
	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		respondJSON(w, http.StatusOK, `{"data":[],"total":0,"page":2,"items":10}`)
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	_, err := c.ListFormSubmissions(context.Background(), "form-1", &ListFormSubmissionsParams{
		SubmissionID: "sub-9",
		OrderBy:      "submitter_email",
		Order:        "asc",
		Page:         2,
		Items:        10,
	})
	if err != nil {
		t.Fatalf("ListFormSubmissions() error: %v", err)
	}

	want := map[string]string{
		"submission_id": "sub-9",
		"order_by":      "submitter_email",
		"order":         "asc",
		"page":          "2",
		"items":         "10",
	}
	for k, v := range want {
		if got := gotQuery.Get(k); got != v {
			t.Errorf("query[%q] = %q, want %q", k, got, v)
		}
	}
	if len(gotQuery) != len(want) {
		t.Errorf("query has %d params, want %d: %v", len(gotQuery), len(want), gotQuery)
	}
}

func TestListFormSubmissions_ZeroValueParamsOmitted(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		respondJSON(w, http.StatusOK, `{"data":[],"total":0,"page":1,"items":50}`)
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	if _, err := c.ListFormSubmissions(context.Background(), "form-1", &ListFormSubmissionsParams{}); err != nil {
		t.Fatalf("ListFormSubmissions() error: %v", err)
	}
	if gotQuery != "" {
		t.Errorf("query = %q, want empty for zero-value params", gotQuery)
	}
}

func TestListFormSubmissions_Validation(t *testing.T) {
	c, _ := NewForms("key") // never dials: validation fails first

	tests := []struct {
		name    string
		formID  string
		wantErr string
	}{
		{"empty form id", "", "paubox: ListFormSubmissions: form id must not be empty"},
		{"whitespace form id", "   ", "paubox: ListFormSubmissions: form id must not be empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.ListFormSubmissions(context.Background(), tt.formID, nil)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestListFormSubmissions_KeylessClientRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("keyless client must not reach the server")
	}))
	defer srv.Close()

	c := newTestPublicFormsClient(t, srv)
	_, err := c.ListFormSubmissions(context.Background(), "form-1", nil)
	if err == nil {
		t.Fatal("expected error from keyless client")
	}
	const want = "paubox: ListFormSubmissions: an API key with the forms scope is required"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestListFormSubmissions_ErrorMapping(t *testing.T) {
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

			c := newTestFormsClient(t, srv)
			_, err := c.ListFormSubmissions(context.Background(), "form-1", nil)
			if err == nil {
				t.Fatalf("expected error for %d response", tt.status)
			}
			if !errors.Is(err, tt.sentinel) {
				t.Errorf("errors.Is(err, sentinel %d) = false; err = %v", tt.status, err)
			}
		})
	}
}

func TestListFormSubmissions_NotFoundTitleFromMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusNotFound, `{"message":"Form not found"}`)
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	_, err := c.ListFormSubmissions(context.Background(), "missing", nil)

	var apiErr *PauboxError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is %T, want *PauboxError", err)
	}
	if apiErr.Title != "Form not found" {
		t.Errorf("Title = %q, want %q", apiErr.Title, "Form not found")
	}
}

// ---------------------------------------------------------------------------
// ExportFormSubmissionsCSV
// ---------------------------------------------------------------------------

func TestExportFormSubmissionsCSV_HappyPath(t *testing.T) {
	csvBody := []byte("Created At,First name\n2026-08-01 12:00:00 PM,Alice\n")
	var gotMethod, gotPath, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=form_data.csv")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(csvBody)
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	got, err := c.ExportFormSubmissionsCSV(context.Background(), "form-1")
	if err != nil {
		t.Fatalf("ExportFormSubmissionsCSV() error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/api/forms/form-1/submissions/submission-csv" {
		t.Errorf("path = %q, want /api/forms/form-1/submissions/submission-csv", gotPath)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-key")
	}
	if strings.Contains(gotAuth, "Token token=") {
		t.Errorf("Forms request must not use the Email API auth format, got %q", gotAuth)
	}
	if !bytes.Equal(got, csvBody) {
		t.Errorf("body = %q, want %q", got, csvBody)
	}
}

func TestExportFormSubmissionsCSV_Validation(t *testing.T) {
	c, _ := NewForms("key")

	for _, formID := range []string{"", "  \t"} {
		_, err := c.ExportFormSubmissionsCSV(context.Background(), formID)
		if err == nil {
			t.Fatalf("expected validation error for formID %q", formID)
		}
		const want = "paubox: ExportFormSubmissionsCSV: form id must not be empty"
		if err.Error() != want {
			t.Errorf("error = %q, want %q", err.Error(), want)
		}
	}
}

func TestExportFormSubmissionsCSV_KeylessClientRejected(t *testing.T) {
	c, _ := NewForms("")
	_, err := c.ExportFormSubmissionsCSV(context.Background(), "form-1")
	if err == nil {
		t.Fatal("expected error from keyless client")
	}
	const want = "paubox: ExportFormSubmissionsCSV: an API key with the forms scope is required"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

// The service's CSV error path answers 404 with a text/plain body and a
// Content-Disposition header — the error parser must fall back to the HTTP
// status text and still match ErrNotFound.
func TestExportFormSubmissionsCSV_PlainText404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", "attachment; filename=form_data.csv")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Form not found"))
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	_, err := c.ExportFormSubmissionsCSV(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("errors.Is(err, ErrNotFound) = false; err = %v", err)
	}

	var apiErr *PauboxError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is %T, want *PauboxError", err)
	}
	if apiErr.Title != http.StatusText(http.StatusNotFound) {
		t.Errorf("Title = %q, want status text %q", apiErr.Title, http.StatusText(http.StatusNotFound))
	}
	if string(apiErr.Raw) != "Form not found" {
		t.Errorf("Raw = %q, want the plain-text body", apiErr.Raw)
	}
}

func TestExportFormSubmissionsCSV_ErrorMapping(t *testing.T) {
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

			c := newTestFormsClient(t, srv)
			_, err := c.ExportFormSubmissionsCSV(context.Background(), "form-1")
			if !errors.Is(err, tt.sentinel) {
				t.Errorf("errors.Is(err, sentinel %d) = false; err = %v", tt.status, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ExportFormSubmissionCSV
// ---------------------------------------------------------------------------

func TestExportFormSubmissionCSV_HappyPath(t *testing.T) {
	csvBody := []byte("Created At,First name\n2026-08-01 12:00:00 PM,Alice\n")
	var gotMethod, gotPath, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(csvBody)
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	got, err := c.ExportFormSubmissionCSV(context.Background(), "form-1", "sub-1")
	if err != nil {
		t.Fatalf("ExportFormSubmissionCSV() error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/api/forms/form-1/submissions/submission-csv/sub-1" {
		t.Errorf("path = %q, want /api/forms/form-1/submissions/submission-csv/sub-1", gotPath)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-key")
	}
	if !bytes.Equal(got, csvBody) {
		t.Errorf("body = %q, want %q", got, csvBody)
	}
}

func TestExportFormSubmissionCSV_Validation(t *testing.T) {
	c, _ := NewForms("key")

	tests := []struct {
		name         string
		formID       string
		submissionID string
		wantErr      string
	}{
		{"empty form id", "", "sub-1", "paubox: ExportFormSubmissionCSV: form id must not be empty"},
		{"whitespace form id", " ", "sub-1", "paubox: ExportFormSubmissionCSV: form id must not be empty"},
		{"empty submission id", "form-1", "", "paubox: ExportFormSubmissionCSV: submission id must not be empty"},
		{"whitespace submission id", "form-1", "\t ", "paubox: ExportFormSubmissionCSV: submission id must not be empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.ExportFormSubmissionCSV(context.Background(), tt.formID, tt.submissionID)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestExportFormSubmissionCSV_KeylessClientRejected(t *testing.T) {
	c, _ := NewForms("")
	_, err := c.ExportFormSubmissionCSV(context.Background(), "form-1", "sub-1")
	if err == nil {
		t.Fatal("expected error from keyless client")
	}
	const want = "paubox: ExportFormSubmissionCSV: an API key with the forms scope is required"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

// "Submission not found" is also served as plain text with a
// Content-Disposition header on the single-submission CSV route.
func TestExportFormSubmissionCSV_PlainText404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", "attachment; filename=form_data.csv")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Submission not found"))
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	_, err := c.ExportFormSubmissionCSV(context.Background(), "form-1", "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("errors.Is(err, ErrNotFound) = false; err = %v", err)
	}
}

func TestExportFormSubmissionCSV_ErrorMapping(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		sentinel error
	}{
		{"bad request", http.StatusBadRequest, `{"message":"bad request"}`, ErrBadRequest},
		{"unauthorized", http.StatusUnauthorized, `{"message":"Invalid API key"}`, ErrUnauthorized},
		{"not found", http.StatusNotFound, `{"message":"Submission not found"}`, ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				respondJSON(w, tt.status, tt.body)
			}))
			defer srv.Close()

			c := newTestFormsClient(t, srv)
			_, err := c.ExportFormSubmissionCSV(context.Background(), "form-1", "sub-1")
			if !errors.Is(err, tt.sentinel) {
				t.Errorf("errors.Is(err, sentinel %d) = false; err = %v", tt.status, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ExportFormSubmissionPDF
// ---------------------------------------------------------------------------

func TestExportFormSubmissionPDF_HappyPath(t *testing.T) {
	pdfBody := []byte("%PDF-1.7 fake pdf bytes")
	var gotMethod, gotPath, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=form_data.pdf")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(pdfBody)
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	got, err := c.ExportFormSubmissionPDF(context.Background(), "form-1", "sub-1")
	if err != nil {
		t.Fatalf("ExportFormSubmissionPDF() error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/api/forms/form-1/submissions/sub-1/submission-pdf" {
		t.Errorf("path = %q, want /api/forms/form-1/submissions/sub-1/submission-pdf", gotPath)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-key")
	}
	if strings.Contains(gotAuth, "Token token=") {
		t.Errorf("Forms request must not use the Email API auth format, got %q", gotAuth)
	}
	if !bytes.Equal(got, pdfBody) {
		t.Errorf("body = %q, want %q", got, pdfBody)
	}
}

func TestExportFormSubmissionPDF_Validation(t *testing.T) {
	c, _ := NewForms("key")

	tests := []struct {
		name         string
		formID       string
		submissionID string
		wantErr      string
	}{
		{"empty form id", "", "sub-1", "paubox: ExportFormSubmissionPDF: form id must not be empty"},
		{"whitespace form id", "  ", "sub-1", "paubox: ExportFormSubmissionPDF: form id must not be empty"},
		{"empty submission id", "form-1", "", "paubox: ExportFormSubmissionPDF: submission id must not be empty"},
		{"whitespace submission id", "form-1", "  ", "paubox: ExportFormSubmissionPDF: submission id must not be empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.ExportFormSubmissionPDF(context.Background(), tt.formID, tt.submissionID)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestExportFormSubmissionPDF_KeylessClientRejected(t *testing.T) {
	c, _ := NewForms("")
	_, err := c.ExportFormSubmissionPDF(context.Background(), "form-1", "sub-1")
	if err == nil {
		t.Fatal("expected error from keyless client")
	}
	const want = "paubox: ExportFormSubmissionPDF: an API key with the forms scope is required"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestExportFormSubmissionPDF_ErrorMapping(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		sentinel error
	}{
		{"bad request", http.StatusBadRequest, `{"message":"bad request"}`, ErrBadRequest},
		{"unauthorized", http.StatusUnauthorized, `{"message":"Invalid API key"}`, ErrUnauthorized},
		{"not found", http.StatusNotFound, `{"message":"Submission not found"}`, ErrNotFound},
		{"server error", http.StatusInternalServerError, `{"message":"Form is missing JSON definition"}`, ErrServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				respondJSON(w, tt.status, tt.body)
			}))
			defer srv.Close()

			c := newTestFormsClient(t, srv)
			_, err := c.ExportFormSubmissionPDF(context.Background(), "form-1", "sub-1")
			if err == nil {
				t.Fatalf("expected error for %d response", tt.status)
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
		})
	}
}

func TestExportFormSubmissionPDF_RequestIDCaptured(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Request-Id", "req-pdf-7")
		respondJSON(w, http.StatusNotFound, `{"message":"Submission not found"}`)
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	_, err := c.ExportFormSubmissionPDF(context.Background(), "form-1", "missing")

	var apiErr *PauboxError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is %T, want *PauboxError", err)
	}
	if apiErr.RequestID != "req-pdf-7" {
		t.Errorf("RequestID = %q, want %q", apiErr.RequestID, "req-pdf-7")
	}
	if apiErr.Title != "Submission not found" {
		t.Errorf("Title = %q, want %q", apiErr.Title, "Submission not found")
	}
}

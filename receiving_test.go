package paubox

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// ListReceivingDomains
// ---------------------------------------------------------------------------

func TestListReceivingDomains_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `[{"id":1,"slug":"test","domain":"test.paubox.net","status":"active","created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}]`)
	}))
	defer srv.Close()

	domains, err := newTestClient(t, srv).ListReceivingDomains(context.Background())
	if err != nil {
		t.Fatalf("ListReceivingDomains() error: %v", err)
	}
	if len(domains) != 1 {
		t.Fatalf("len(domains) = %d, want 1", len(domains))
	}
	if domains[0].Slug != "test" {
		t.Errorf("Slug = %q, want test", domains[0].Slug)
	}
}

func TestListReceivingDomains_SendsCorrectMethodAndPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `[]`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).ListReceivingDomains(context.Background())
	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/receiving/domains" {
		t.Errorf("path = %q, want /receiving/domains", gotPath)
	}
}

func TestListReceivingDomains_401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 401, `{"errors":[{"code":401,"title":"Unauthorized","details":"bad key"}]}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).ListReceivingDomains(context.Background())
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// CreateReceivingDomain
// ---------------------------------------------------------------------------

func TestCreateReceivingDomain_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"id":2,"slug":"myslug","domain":"myslug.paubox.net","status":"pending","created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}`)
	}))
	defer srv.Close()

	domain, err := newTestClient(t, srv).CreateReceivingDomain(context.Background(), &CreateReceivingDomainRequest{Slug: "myslug"})
	if err != nil {
		t.Fatalf("CreateReceivingDomain() error: %v", err)
	}
	if domain.Slug != "myslug" {
		t.Errorf("Slug = %q, want myslug", domain.Slug)
	}
}

func TestCreateReceivingDomain_NilRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"id":3,"slug":"auto","domain":"auto.paubox.net","status":"pending","created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}`)
	}))
	defer srv.Close()

	domain, err := newTestClient(t, srv).CreateReceivingDomain(context.Background(), nil)
	if err != nil {
		t.Fatalf("CreateReceivingDomain(nil) error: %v", err)
	}
	if domain.ID != 3 {
		t.Errorf("ID = %d, want 3", domain.ID)
	}
}

func TestCreateReceivingDomain_SendsCorrectMethodAndPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"id":1}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).CreateReceivingDomain(context.Background(), &CreateReceivingDomainRequest{Slug: "x"})
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/receiving/domains" {
		t.Errorf("path = %q, want /receiving/domains", gotPath)
	}
}

func TestCreateReceivingDomain_SendsSlugInBody(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		respondJSON(w, http.StatusOK, `{"id":1}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).CreateReceivingDomain(context.Background(), &CreateReceivingDomainRequest{Slug: "myslug"})
	if gotBody["slug"] != "myslug" {
		t.Errorf("slug = %v, want myslug", gotBody["slug"])
	}
}

// ---------------------------------------------------------------------------
// GetReceivingDomain
// ---------------------------------------------------------------------------

func TestGetReceivingDomain_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"id":5,"slug":"test","domain":"test.paubox.net","status":"active","created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}`)
	}))
	defer srv.Close()

	domain, err := newTestClient(t, srv).GetReceivingDomain(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetReceivingDomain() error: %v", err)
	}
	if domain.ID != 5 {
		t.Errorf("ID = %d, want 5", domain.ID)
	}
}

func TestGetReceivingDomain_SendsCorrectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"id":5}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).GetReceivingDomain(context.Background(), 5)
	if gotPath != "/receiving/domains/5" {
		t.Errorf("path = %q, want /receiving/domains/5", gotPath)
	}
}

func TestGetReceivingDomain_InvalidID(t *testing.T) {
	c, _ := New("k")
	_, err := c.GetReceivingDomain(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error for zero domainID")
	}
	if !strings.Contains(err.Error(), "domainID") {
		t.Errorf("error %q should mention domainID", err.Error())
	}
}

func TestGetReceivingDomain_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 404, `{"errors":[{"code":404,"title":"Not Found","details":"domain not found"}]}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).GetReceivingDomain(context.Background(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// DeleteReceivingDomain
// ---------------------------------------------------------------------------

func TestDeleteReceivingDomain_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := newTestClient(t, srv).DeleteReceivingDomain(context.Background(), 5)
	if err != nil {
		t.Fatalf("DeleteReceivingDomain() error: %v", err)
	}
}

func TestDeleteReceivingDomain_SendsCorrectMethodAndPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	_ = newTestClient(t, srv).DeleteReceivingDomain(context.Background(), 7)
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	if gotPath != "/receiving/domains/7" {
		t.Errorf("path = %q, want /receiving/domains/7", gotPath)
	}
}

func TestDeleteReceivingDomain_InvalidID(t *testing.T) {
	c, _ := New("k")
	err := c.DeleteReceivingDomain(context.Background(), -1)
	if err == nil {
		t.Fatal("expected error for negative domainID")
	}
}

// ---------------------------------------------------------------------------
// ListMailboxes
// ---------------------------------------------------------------------------

func TestListMailboxes_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `[{"id":10,"domain_id":1,"name":"info","email":"info@test.paubox.net","created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}]`)
	}))
	defer srv.Close()

	mailboxes, err := newTestClient(t, srv).ListMailboxes(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListMailboxes() error: %v", err)
	}
	if len(mailboxes) != 1 {
		t.Fatalf("len(mailboxes) = %d, want 1", len(mailboxes))
	}
	if mailboxes[0].Name != "info" {
		t.Errorf("Name = %q, want info", mailboxes[0].Name)
	}
}

func TestListMailboxes_SendsCorrectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `[]`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).ListMailboxes(context.Background(), 3)
	if gotPath != "/receiving/domains/3/mailboxes" {
		t.Errorf("path = %q, want /receiving/domains/3/mailboxes", gotPath)
	}
}

func TestListMailboxes_InvalidDomainID(t *testing.T) {
	c, _ := New("k")
	_, err := c.ListMailboxes(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error for zero domainID")
	}
}

// ---------------------------------------------------------------------------
// CreateMailbox
// ---------------------------------------------------------------------------

func TestCreateMailbox_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"id":20,"domain_id":1,"name":"support","email":"support@test.paubox.net","created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}`)
	}))
	defer srv.Close()

	mb, err := newTestClient(t, srv).CreateMailbox(context.Background(), 1, &CreateMailboxRequest{
		Name:     "support",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("CreateMailbox() error: %v", err)
	}
	if mb.Name != "support" {
		t.Errorf("Name = %q, want support", mb.Name)
	}
}

func TestCreateMailbox_SendsCorrectMethodAndPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"id":1}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).CreateMailbox(context.Background(), 2, &CreateMailboxRequest{
		Name: "x", Password: "p",
	})
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/receiving/domains/2/mailboxes" {
		t.Errorf("path = %q, want /receiving/domains/2/mailboxes", gotPath)
	}
}

func TestCreateMailbox_SendsBody(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		respondJSON(w, http.StatusOK, `{"id":1}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).CreateMailbox(context.Background(), 1, &CreateMailboxRequest{
		Name:       "inbox",
		Password:   "pass",
		QuotaBytes: Ptr(int64(1048576)),
	})
	if gotBody["name"] != "inbox" {
		t.Errorf("name = %v, want inbox", gotBody["name"])
	}
	if gotBody["password"] != "pass" {
		t.Errorf("password = %v, want pass", gotBody["password"])
	}
	if gotBody["quota_bytes"] != float64(1048576) {
		t.Errorf("quota_bytes = %v, want 1048576", gotBody["quota_bytes"])
	}
}

func TestCreateMailbox_Validation(t *testing.T) {
	tests := []struct {
		name     string
		domainID int
		req      *CreateMailboxRequest
		wantErr  string
	}{
		{"zero domainID", 0, &CreateMailboxRequest{Name: "x", Password: "p"}, "domainID"},
		{"nil request", 1, nil, "nil"},
		{"empty name", 1, &CreateMailboxRequest{Name: "", Password: "p"}, "name"},
		{"whitespace name", 1, &CreateMailboxRequest{Name: "   ", Password: "p"}, "name"},
		{"empty password", 1, &CreateMailboxRequest{Name: "x", Password: ""}, "password"},
	}

	c, _ := New("k")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := c.CreateMailbox(context.Background(), tc.domainID, tc.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetMailbox
// ---------------------------------------------------------------------------

func TestGetMailbox_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"id":10,"domain_id":1,"name":"info","email":"info@test.paubox.net","created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}`)
	}))
	defer srv.Close()

	mb, err := newTestClient(t, srv).GetMailbox(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("GetMailbox() error: %v", err)
	}
	if mb.ID != 10 {
		t.Errorf("ID = %d, want 10", mb.ID)
	}
}

func TestGetMailbox_SendsCorrectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"id":10}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).GetMailbox(context.Background(), 3, 10)
	if gotPath != "/receiving/domains/3/mailboxes/10" {
		t.Errorf("path = %q, want /receiving/domains/3/mailboxes/10", gotPath)
	}
}

func TestGetMailbox_InvalidIDs(t *testing.T) {
	c, _ := New("k")

	_, err := c.GetMailbox(context.Background(), 0, 1)
	if err == nil {
		t.Fatal("expected error for zero domainID")
	}

	_, err = c.GetMailbox(context.Background(), 1, 0)
	if err == nil {
		t.Fatal("expected error for zero mailboxID")
	}
}

// ---------------------------------------------------------------------------
// DeleteMailbox
// ---------------------------------------------------------------------------

func TestDeleteMailbox_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := newTestClient(t, srv).DeleteMailbox(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("DeleteMailbox() error: %v", err)
	}
}

func TestDeleteMailbox_SendsCorrectMethodAndPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	_ = newTestClient(t, srv).DeleteMailbox(context.Background(), 2, 15)
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	if gotPath != "/receiving/domains/2/mailboxes/15" {
		t.Errorf("path = %q, want /receiving/domains/2/mailboxes/15", gotPath)
	}
}

func TestDeleteMailbox_InvalidIDs(t *testing.T) {
	c, _ := New("k")

	err := c.DeleteMailbox(context.Background(), 0, 1)
	if err == nil {
		t.Fatal("expected error for zero domainID")
	}

	err = c.DeleteMailbox(context.Background(), 1, 0)
	if err == nil {
		t.Fatal("expected error for zero mailboxID")
	}
}

// ---------------------------------------------------------------------------
// ListReceivedEmails
// ---------------------------------------------------------------------------

func TestListReceivedEmails_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"emails":[{"id":"em-1","from":"sender@example.com","to":["r@example.com"],"subject":"Hi","received_at":"2025-01-01T00:00:00Z"}]}`)
	}))
	defer srv.Close()

	resp, err := newTestClient(t, srv).ListReceivedEmails(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListReceivedEmails() error: %v", err)
	}
	if len(resp.Emails) != 1 {
		t.Fatalf("len(Emails) = %d, want 1", len(resp.Emails))
	}
	if resp.Emails[0].ID != "em-1" {
		t.Errorf("ID = %q, want em-1", resp.Emails[0].ID)
	}
}

func TestListReceivedEmails_SendsCorrectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"emails":[]}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).ListReceivedEmails(context.Background(), nil)
	if gotPath != "/receiving" {
		t.Errorf("path = %q, want /receiving", gotPath)
	}
}

func TestListReceivedEmails_QueryParams(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		respondJSON(w, http.StatusOK, `{"emails":[]}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).ListReceivedEmails(context.Background(), &ListReceivedEmailsRequest{
		Limit:  Ptr(25),
		After:  Ptr("cursor-abc"),
		Before: Ptr("cursor-xyz"),
	})
	if !strings.Contains(gotQuery, "limit=25") {
		t.Errorf("query %q missing limit=25", gotQuery)
	}
	if !strings.Contains(gotQuery, "after=cursor-abc") {
		t.Errorf("query %q missing after=cursor-abc", gotQuery)
	}
	if !strings.Contains(gotQuery, "before=cursor-xyz") {
		t.Errorf("query %q missing before=cursor-xyz", gotQuery)
	}
}

// ---------------------------------------------------------------------------
// GetReceivedEmail
// ---------------------------------------------------------------------------

func TestGetReceivedEmail_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"id":"em-5","from":"s@example.com","to":["r@example.com"],"subject":"Test","body":{"text/plain":"hello"},"received_at":"2025-01-01T00:00:00Z"}`)
	}))
	defer srv.Close()

	email, err := newTestClient(t, srv).GetReceivedEmail(context.Background(), "em-5")
	if err != nil {
		t.Fatalf("GetReceivedEmail() error: %v", err)
	}
	if email.ID != "em-5" {
		t.Errorf("ID = %q, want em-5", email.ID)
	}
	if email.Body == nil || email.Body.PlainText == nil || *email.Body.PlainText != "hello" {
		t.Error("Body.PlainText not populated")
	}
}

func TestGetReceivedEmail_SendsCorrectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"id":"em-5"}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).GetReceivedEmail(context.Background(), "em-5")
	if gotPath != "/receiving/em-5" {
		t.Errorf("path = %q, want /receiving/em-5", gotPath)
	}
}

func TestGetReceivedEmail_EmptyID(t *testing.T) {
	c, _ := New("k")
	_, err := c.GetReceivedEmail(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty emailID")
	}
}

func TestGetReceivedEmail_WhitespaceID(t *testing.T) {
	c, _ := New("k")
	_, err := c.GetReceivedEmail(context.Background(), "   ")
	if err == nil {
		t.Fatal("expected error for whitespace emailID")
	}
}

func TestGetReceivedEmail_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 404, `{"errors":[{"code":404,"title":"Not Found","details":"email not found"}]}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).GetReceivedEmail(context.Background(), "bad-id")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// DownloadAttachment
// ---------------------------------------------------------------------------

func TestDownloadAttachment_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"blob_id":"blob-1","file_name":"report.pdf","content_type":"application/pdf","data":"base64data"}`)
	}))
	defer srv.Close()

	dl, err := newTestClient(t, srv).DownloadAttachment(context.Background(), "em-5", "blob-1")
	if err != nil {
		t.Fatalf("DownloadAttachment() error: %v", err)
	}
	if dl.BlobID != "blob-1" {
		t.Errorf("BlobID = %q, want blob-1", dl.BlobID)
	}
	if dl.FileName != "report.pdf" {
		t.Errorf("FileName = %q, want report.pdf", dl.FileName)
	}
}

func TestDownloadAttachment_SendsCorrectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"blob_id":"b"}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).DownloadAttachment(context.Background(), "em-5", "blob-1")
	if gotPath != "/receiving/em-5/attachments/blob-1" {
		t.Errorf("path = %q, want /receiving/em-5/attachments/blob-1", gotPath)
	}
}

func TestDownloadAttachment_Validation(t *testing.T) {
	c, _ := New("k")

	_, err := c.DownloadAttachment(context.Background(), "", "blob-1")
	if err == nil {
		t.Fatal("expected error for empty emailID")
	}
	if !strings.Contains(err.Error(), "emailID") {
		t.Errorf("error %q should mention emailID", err.Error())
	}

	_, err = c.DownloadAttachment(context.Background(), "em-5", "")
	if err == nil {
		t.Fatal("expected error for empty blobID")
	}
	if !strings.Contains(err.Error(), "blobID") {
		t.Errorf("error %q should mention blobID", err.Error())
	}
}

func TestDownloadAttachment_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 404, `{"errors":[{"code":404,"title":"Not Found","details":"attachment not found"}]}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).DownloadAttachment(context.Background(), "em-5", "bad-blob")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

package paubox

import (
	"bytes"
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
		respondJSON(w, http.StatusOK, `{"data":[{"id":1,"slug":"test","domain":"test.paubox.net","state":"active","created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}]}`)
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
		respondJSON(w, http.StatusOK, `{"data":[]}`)
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
		respondJSON(w, http.StatusOK, `{"data":{"id":2,"slug":"myslug","domain":"myslug.paubox.net","state":"pending","created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}}`)
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
		respondJSON(w, http.StatusOK, `{"data":{"id":3,"slug":"auto","domain":"auto.paubox.net","state":"pending","created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}}`)
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
		respondJSON(w, http.StatusOK, `{"data":{"id":1}}`)
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
		respondJSON(w, http.StatusOK, `{"data":{"id":1}}`)
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
		respondJSON(w, http.StatusOK, `{"data":{"id":5,"slug":"test","domain":"test.paubox.net","state":"active","created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}}`)
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
		respondJSON(w, http.StatusOK, `{"data":{"id":5}}`)
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
		respondJSON(w, http.StatusOK, `{"data":[{"id":10,"domain_id":1,"name":"info","email_address":"info@test.paubox.net","created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}]}`)
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
		respondJSON(w, http.StatusOK, `{"data":[]}`)
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
		respondJSON(w, http.StatusOK, `{"data":{"id":20,"domain_id":1,"name":"support","email_address":"support@test.paubox.net","created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}}`)
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
		respondJSON(w, http.StatusOK, `{"data":{"id":1}}`)
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
		respondJSON(w, http.StatusOK, `{"data":{"id":1}}`)
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
		respondJSON(w, http.StatusOK, `{"data":{"id":10,"domain_id":1,"name":"info","email_address":"info@test.paubox.net","created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}}`)
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
		respondJSON(w, http.StatusOK, `{"data":{"id":10}}`)
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

const (
	testEmailUUID      = "0b6f3c2e-6a4f-4f7e-9d0a-2f4b8c1d9e10"
	testAttachmentUUID = "5d2a9e41-3c7b-4e8f-a1b6-7c0d2e9f4a83"
)

func TestListReceivedEmails_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"object":"list","data":[{"email_id":"`+testEmailUUID+`","from":[{"name":"Sender","address":"sender@example.com"}],"to":[{"name":null,"address":"r@example.com"}],"subject":"Hi","received_at":"2025-01-01T00:00:00Z","has_attachment":true,"spam":false,"size":2048,"domain":"test.paubox.net"}],"has_more":true}`)
	}))
	defer srv.Close()

	resp, err := newTestClient(t, srv).ListReceivedEmails(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListReceivedEmails() error: %v", err)
	}
	if resp.Object != "list" || !resp.HasMore {
		t.Errorf("Object = %q, HasMore = %v, want list, true", resp.Object, resp.HasMore)
	}
	if len(resp.Emails) != 1 {
		t.Fatalf("len(Emails) = %d, want 1", len(resp.Emails))
	}
	e := resp.Emails[0]
	if e.EmailID != testEmailUUID {
		t.Errorf("EmailID = %q, want %s", e.EmailID, testEmailUUID)
	}
	if e.From[0].Name == nil || *e.From[0].Name != "Sender" || e.From[0].Address != "sender@example.com" {
		t.Errorf("From = %+v, want Sender <sender@example.com>", e.From[0])
	}
	if e.To[0].Name != nil {
		t.Errorf("To[0].Name = %v, want nil", *e.To[0].Name)
	}
	if !e.HasAttachment || e.Spam || e.Size != 2048 || e.Domain != "test.paubox.net" || e.Subject != "Hi" {
		t.Errorf("unexpected list item: %+v", e)
	}
}

func TestListReceivedEmails_NullableFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"object":"list","data":[{"email_id":"`+testEmailUUID+`","from":[{"name":null,"address":null}],"to":[],"subject":null,"received_at":null,"has_attachment":null,"spam":false,"size":null,"domain":"test.paubox.net"}],"has_more":false}`)
	}))
	defer srv.Close()

	resp, err := newTestClient(t, srv).ListReceivedEmails(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListReceivedEmails() error: %v", err)
	}
	e := resp.Emails[0]
	if e.Subject != "" || e.ReceivedAt != "" || e.HasAttachment || e.Size != 0 || e.From[0].Address != "" {
		t.Errorf("null fields should decode to zero values, got %+v", e)
	}
}

func TestListReceivedEmails_SendsCorrectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"data":[],"has_more":false,"object":"list"}`)
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
		respondJSON(w, http.StatusOK, `{"data":[],"has_more":false,"object":"list"}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).ListReceivedEmails(context.Background(), &ListReceivedEmailsRequest{
		Limit:     Ptr(25),
		After:     Ptr(testEmailUUID),
		Before:    Ptr(testAttachmentUUID),
		Search:    Ptr("lab results"),
		Sort:      Ptr("received_at"),
		Ascending: Ptr(true),
	})
	for _, want := range []string{
		"limit=25",
		"after=" + testEmailUUID,
		"before=" + testAttachmentUUID,
		"search=lab+results",
		"sort=received_at",
		"ascending=true",
	} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("query %q missing %s", gotQuery, want)
		}
	}
}

func TestListReceivedEmails_EmptyRequestSendsNoQuery(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		respondJSON(w, http.StatusOK, `{"data":[],"has_more":false,"object":"list"}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).ListReceivedEmails(context.Background(), &ListReceivedEmailsRequest{})
	if gotQuery != "" {
		t.Errorf("query = %q, want empty", gotQuery)
	}
}

// ---------------------------------------------------------------------------
// GetReceivedEmail
// ---------------------------------------------------------------------------

const receivedEmailDetailJSON = `{"data":{
	"email_id":"` + testEmailUUID + `",
	"from":[{"name":"Sender","address":"s@example.com"}],
	"to":[{"name":null,"address":"r@example.com"}],
	"cc":[{"name":"Copy","address":"cc@example.com"}],
	"subject":"Test",
	"date":"Wed, 01 Jan 2025 00:00:00 +0000",
	"received_at":"2025-01-01T00:00:01Z",
	"message_id":["<abc@example.com>"],
	"in_reply_to":["<parent@example.com>"],
	"references":["<root@example.com>","<parent@example.com>"],
	"spam":false,
	"spam_score":1.5,
	"text_body":"hello",
	"html_body":"<p>hello</p>",
	"attachments":[{"id":"` + testAttachmentUUID + `","filename":"report.pdf","content_type":"application/pdf","size":1024,"content_id":"img1@example.com","download_url":"https://api.paubox.com/v1/email/receiving/` + testEmailUUID + `/attachments/` + testAttachmentUUID + `"}],
	"size":4096,
	"authentication":{"spf":"pass","dkim":"pass","dmarc":"fail"},
	"domain":"test.paubox.net",
	"headers":[{"name":"X-Test","value":"1"}]
}}`

func TestGetReceivedEmail_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, receivedEmailDetailJSON)
	}))
	defer srv.Close()

	email, err := newTestClient(t, srv).GetReceivedEmail(context.Background(), testEmailUUID)
	if err != nil {
		t.Fatalf("GetReceivedEmail() error: %v", err)
	}
	if email.EmailID != testEmailUUID {
		t.Errorf("EmailID = %q, want %s", email.EmailID, testEmailUUID)
	}
	if email.TextBody != "hello" || email.HTMLBody != "<p>hello</p>" {
		t.Errorf("TextBody/HTMLBody = %q/%q", email.TextBody, email.HTMLBody)
	}
	if len(email.CC) != 1 || email.CC[0].Address != "cc@example.com" {
		t.Errorf("CC = %+v", email.CC)
	}
	if email.Date != "Wed, 01 Jan 2025 00:00:00 +0000" || email.ReceivedAt != "2025-01-01T00:00:01Z" {
		t.Errorf("Date/ReceivedAt = %q/%q", email.Date, email.ReceivedAt)
	}
	if len(email.MessageID) != 1 || email.MessageID[0] != "<abc@example.com>" {
		t.Errorf("MessageID = %v", email.MessageID)
	}
	if len(email.InReplyTo) != 1 || email.InReplyTo[0] != "<parent@example.com>" {
		t.Errorf("InReplyTo = %v", email.InReplyTo)
	}
	if len(email.References) != 2 || email.References[0] != "<root@example.com>" {
		t.Errorf("References = %v", email.References)
	}
	if email.SpamScore == nil || *email.SpamScore != 1.5 {
		t.Errorf("SpamScore = %v, want 1.5", email.SpamScore)
	}
	if email.Size != 4096 || email.Domain != "test.paubox.net" {
		t.Errorf("Size/Domain = %d/%q", email.Size, email.Domain)
	}
	if email.Authentication == nil {
		t.Fatal("Authentication = nil")
	}
	if *email.Authentication != (ReceivedEmailAuthentication{SPF: "pass", DKIM: "pass", DMARC: "fail"}) {
		t.Errorf("Authentication = %+v", *email.Authentication)
	}
	if len(email.Headers) != 1 || email.Headers[0] != (ReceivedEmailHeader{Name: "X-Test", Value: "1"}) {
		t.Errorf("Headers = %+v", email.Headers)
	}
	if email.AccountID != "" {
		t.Errorf("AccountID = %q, want empty", email.AccountID)
	}

	if len(email.Attachments) != 1 {
		t.Fatalf("len(Attachments) = %d, want 1", len(email.Attachments))
	}
	a := email.Attachments[0]
	if a.ID != testAttachmentUUID {
		t.Errorf("Attachment ID = %q, want %s", a.ID, testAttachmentUUID)
	}
	if a.FileName != "report.pdf" || a.ContentType != "application/pdf" || a.Size != 1024 {
		t.Errorf("attachment metadata = %+v", a)
	}
	if a.ContentID == nil || *a.ContentID != "img1@example.com" {
		t.Errorf("ContentID = %v, want img1@example.com", a.ContentID)
	}
	if !strings.HasSuffix(a.DownloadURL, "/attachments/"+testAttachmentUUID) {
		t.Errorf("DownloadURL = %q", a.DownloadURL)
	}
	if a.BlobID != "" {
		t.Errorf("BlobID = %q, want empty", a.BlobID)
	}
}

func TestGetReceivedEmail_NullableFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"data":{"email_id":"`+testEmailUUID+`","from":[],"to":[],"cc":[],"subject":null,"date":null,"received_at":null,"message_id":null,"in_reply_to":null,"references":null,"spam":true,"spam_score":null,"text_body":null,"html_body":null,"attachments":[{"id":"`+testAttachmentUUID+`","filename":null,"content_type":null,"size":null,"content_id":null,"download_url":"https://example.com/a"}],"size":null,"authentication":{"spf":"none","dkim":"none","dmarc":"none"},"domain":"test.paubox.net","headers":null}}`)
	}))
	defer srv.Close()

	email, err := newTestClient(t, srv).GetReceivedEmail(context.Background(), testEmailUUID)
	if err != nil {
		t.Fatalf("GetReceivedEmail() error: %v", err)
	}
	if !email.Spam || email.SpamScore != nil || email.MessageID != nil || email.Headers != nil {
		t.Errorf("unexpected decode of null fields: %+v", email)
	}
	a := email.Attachments[0]
	if a.FileName != "" || a.ContentType != "" || a.Size != 0 || a.ContentID != nil {
		t.Errorf("null attachment fields should decode to zero values, got %+v", a)
	}
}

func TestGetReceivedEmail_SendsCorrectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"data":{"email_id":"`+testEmailUUID+`"}}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).GetReceivedEmail(context.Background(), testEmailUUID)
	if gotPath != "/receiving/"+testEmailUUID {
		t.Errorf("path = %q, want /receiving/%s", gotPath, testEmailUUID)
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
	body := []byte("%PDF-1.7\x00\xff\xfe binary")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", `attachment; filename="lab report.pdf"`)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	dl, err := newTestClient(t, srv).DownloadAttachment(context.Background(), testEmailUUID, testAttachmentUUID)
	if err != nil {
		t.Fatalf("DownloadAttachment() error: %v", err)
	}
	if !bytes.Equal(dl.Content, body) {
		t.Errorf("Content = %q, want %q", dl.Content, body)
	}
	if dl.ContentType != "application/pdf" {
		t.Errorf("ContentType = %q, want application/pdf", dl.ContentType)
	}
	if dl.FileName != "lab report.pdf" {
		t.Errorf("FileName = %q, want lab report.pdf", dl.FileName)
	}
	if dl.BlobID != "" || dl.Data != nil {
		t.Errorf("deprecated fields should be zero, got BlobID=%q Data=%q", dl.BlobID, dl.Data)
	}
}

func TestDownloadAttachment_JSONContentIsNotDecoded(t *testing.T) {
	body := `{"not":"an envelope"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, body)
	}))
	defer srv.Close()

	dl, err := newTestClient(t, srv).DownloadAttachment(context.Background(), testEmailUUID, testAttachmentUUID)
	if err != nil {
		t.Fatalf("DownloadAttachment() error: %v", err)
	}
	if string(dl.Content) != body {
		t.Errorf("Content = %q, want %q", dl.Content, body)
	}
	if dl.ContentType != "application/json" {
		t.Errorf("ContentType = %q, want application/json", dl.ContentType)
	}
}

func TestDownloadAttachment_NoContentDisposition(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte("data"))
	}))
	defer srv.Close()

	dl, err := newTestClient(t, srv).DownloadAttachment(context.Background(), testEmailUUID, testAttachmentUUID)
	if err != nil {
		t.Fatalf("DownloadAttachment() error: %v", err)
	}
	if dl.FileName != "" {
		t.Errorf("FileName = %q, want empty", dl.FileName)
	}
	if string(dl.Content) != "data" {
		t.Errorf("Content = %q, want data", dl.Content)
	}
}

func TestDownloadAttachment_SendsCorrectMethodAndPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_, _ = w.Write([]byte("x"))
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).DownloadAttachment(context.Background(), testEmailUUID, testAttachmentUUID)
	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	want := "/receiving/" + testEmailUUID + "/attachments/" + testAttachmentUUID
	if gotPath != want {
		t.Errorf("path = %q, want %s", gotPath, want)
	}
}

func TestDownloadAttachment_Validation(t *testing.T) {
	c, _ := New("k")

	_, err := c.DownloadAttachment(context.Background(), "", testAttachmentUUID)
	if err == nil {
		t.Fatal("expected error for empty emailID")
	}
	if !strings.Contains(err.Error(), "emailID") {
		t.Errorf("error %q should mention emailID", err.Error())
	}

	_, err = c.DownloadAttachment(context.Background(), testEmailUUID, "  ")
	if err == nil {
		t.Fatal("expected error for empty attachmentID")
	}
	if !strings.Contains(err.Error(), "attachmentID") {
		t.Errorf("error %q should mention attachmentID", err.Error())
	}
}

func TestDownloadAttachment_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 404, `{"errors":[{"code":404,"title":"Not Found","details":"attachment not found"}]}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).DownloadAttachment(context.Background(), testEmailUUID, "legacy-blob-id")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	var apiErr *PauboxError
	if !errors.As(err, &apiErr) || apiErr.Details != "attachment not found" {
		t.Errorf("expected parsed PauboxError, got %v", err)
	}
}

func TestDownloadAttachment_401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 401, `{"errors":[{"code":401,"title":"Unauthorized","details":"bad key"}]}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).DownloadAttachment(context.Background(), testEmailUUID, testAttachmentUUID)
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestDownloadAttachment_ContextCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("x"))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := newTestClient(t, srv).DownloadAttachment(ctx, testEmailUUID, testAttachmentUUID)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

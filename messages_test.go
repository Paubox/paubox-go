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
// Shared test helpers
// ---------------------------------------------------------------------------

func validMessage() Message {
	return Message{
		Recipients: []string{"recipient@example.com"},
		Headers: MessageHeaders{
			From:    "sender@verified.com",
			Subject: "Test subject",
		},
		Content: MessageContent{PlainText: Ptr("Hello")},
	}
}

func respondJSON(w http.ResponseWriter, statusCode int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(body))
}

// ---------------------------------------------------------------------------
// SendMessage — happy path
// ---------------------------------------------------------------------------

func TestSendMessage_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"sourceTrackingId":"tid-abc","customHeaders":{},"data":"Service OK"}`)
	}))
	defer srv.Close()

	resp, err := newTestClient(t, srv).SendMessage(context.Background(), &SendMessageRequest{
		Message: validMessage(),
	})
	if err != nil {
		t.Fatalf("SendMessage() error: %v", err)
	}
	if resp.SourceTrackingID != "tid-abc" {
		t.Errorf("SourceTrackingID = %q, want tid-abc", resp.SourceTrackingID)
	}
	if resp.Data != "Service OK" {
		t.Errorf("Data = %q, want 'Service OK'", resp.Data)
	}
}

func TestSendMessage_SendsCorrectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		respondJSON(w, 200, `{"sourceTrackingId":"x","data":"Service OK"}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).SendMessage(context.Background(), &SendMessageRequest{Message: validMessage()})
	if gotPath != "/testuser/messages" {
		t.Errorf("path = %q, want /testuser/messages", gotPath)
	}
}

func TestSendMessage_SendsCorrectMethod(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		respondJSON(w, 200, `{"sourceTrackingId":"x","data":"Service OK"}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).SendMessage(context.Background(), &SendMessageRequest{Message: validMessage()})
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
}

func TestSendMessage_AllFields(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		respondJSON(w, 200, `{"sourceTrackingId":"x","data":"Service OK"}`)
	}))
	defer srv.Close()

	req := &SendMessageRequest{
		OverrideOpenTracking: true,
		OverrideLinkTracking: true,
		UnsubscribeURL:       "https://example.com/unsub",
		Message: Message{
			Recipients: []string{"r1@example.com", "r2@example.com"},
			BCC:        []string{"bcc@example.com"},
			CC:         []string{"cc@example.com"},
			Headers: MessageHeaders{
				From:                "f@example.com",
				Subject:             "Hello",
				ReplyTo:             "reply@example.com",
				ListUnsubscribe:     "<mailto:u@example.com>",
				ListUnsubscribePost: "List-Unsubscribe=One-Click",
				CustomHeaders:       map[string]string{"x-my-header": "myval"},
			},
			AllowNonTLS:             true,
			ForceSecureNotification: true,
			Content: MessageContent{
				PlainText: Ptr("Plain body"),
				HTML:      Ptr("<p>HTML body</p>"),
			},
			Attachments: []Attachment{
				{FileName: "f.pdf", ContentType: "application/pdf", Content: "base64data"},
			},
		},
	}

	_, err := newTestClient(t, srv).SendMessage(context.Background(), req)
	if err != nil {
		t.Fatalf("SendMessage() error: %v", err)
	}

	data := gotBody["data"].(map[string]any)
	msg := data["message"].(map[string]any)
	headers := msg["headers"].(map[string]any)

	if headers["x-my-header"] != "myval" {
		t.Errorf("x-my-header = %v, want myval", headers["x-my-header"])
	}
	if headers["reply-to"] != "reply@example.com" {
		t.Errorf("reply-to = %v", headers["reply-to"])
	}
	if headers["List-Unsubscribe"] != "<mailto:u@example.com>" {
		t.Errorf("List-Unsubscribe = %v", headers["List-Unsubscribe"])
	}
	if data["override_open_tracking"] != true {
		t.Errorf("override_open_tracking = %v", data["override_open_tracking"])
	}
}

// ---------------------------------------------------------------------------
// SendMessage — validation
// ---------------------------------------------------------------------------

func TestSendMessage_NilRequest(t *testing.T) {
	c, _ := New("k", "u")
	_, err := c.SendMessage(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
}

func TestSendMessage_Validation(t *testing.T) {
	tests := []struct {
		name    string
		msg     Message
		wantErr string
	}{
		{
			name:    "no recipients",
			msg:     Message{Headers: MessageHeaders{From: "f@x.com", Subject: "s"}, Content: MessageContent{PlainText: Ptr("b")}},
			wantErr: "recipients",
		},
		{
			name:    "empty recipients slice",
			msg:     Message{Recipients: []string{}, Headers: MessageHeaders{From: "f@x.com", Subject: "s"}, Content: MessageContent{PlainText: Ptr("b")}},
			wantErr: "recipients",
		},
		{
			name:    "no from",
			msg:     Message{Recipients: []string{"r@x.com"}, Headers: MessageHeaders{Subject: "s"}, Content: MessageContent{PlainText: Ptr("b")}},
			wantErr: "from",
		},
		{
			name:    "no subject",
			msg:     Message{Recipients: []string{"r@x.com"}, Headers: MessageHeaders{From: "f@x.com"}, Content: MessageContent{PlainText: Ptr("b")}},
			wantErr: "subject",
		},
		{
			name:    "no content",
			msg:     Message{Recipients: []string{"r@x.com"}, Headers: MessageHeaders{From: "f@x.com", Subject: "s"}, Content: MessageContent{}},
			wantErr: "content",
		},
		{
			name: "attachment missing fileName",
			msg: Message{
				Recipients:  []string{"r@x.com"},
				Headers:     MessageHeaders{From: "f@x.com", Subject: "s"},
				Content:     MessageContent{PlainText: Ptr("b")},
				Attachments: []Attachment{{ContentType: "application/pdf", Content: "data"}},
			},
			wantErr: "fileName",
		},
		{
			name: "attachment missing contentType",
			msg: Message{
				Recipients:  []string{"r@x.com"},
				Headers:     MessageHeaders{From: "f@x.com", Subject: "s"},
				Content:     MessageContent{PlainText: Ptr("b")},
				Attachments: []Attachment{{FileName: "f.pdf", Content: "data"}},
			},
			wantErr: "contentType",
		},
		{
			name: "attachment missing content",
			msg: Message{
				Recipients:  []string{"r@x.com"},
				Headers:     MessageHeaders{From: "f@x.com", Subject: "s"},
				Content:     MessageContent{PlainText: Ptr("b")},
				Attachments: []Attachment{{FileName: "f.pdf", ContentType: "application/pdf"}},
			},
			wantErr: "content",
		},
	}

	c, _ := New("k", "u")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := c.SendMessage(context.Background(), &SendMessageRequest{Message: tc.msg})
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
// SendMessage — error responses
// ---------------------------------------------------------------------------

func TestSendMessage_ErrorResponses(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		sentinel   *PauboxError
	}{
		{
			"400", 400,
			`{"errors":[{"code":400,"title":"Bad Request","details":"invalid recipient"}]}`,
			ErrBadRequest,
		},
		{
			"401", 401,
			`{"errors":[{"code":401,"title":"Unauthorized","details":"bad key"}]}`,
			ErrUnauthorized,
		},
		{
			"403", 403,
			`{"errors":[{"code":403,"title":"Forbidden","details":""}]}`,
			ErrForbidden,
		},
		{
			"500", 500,
			`{"errors":[{"code":500,"title":"Internal Server Error","details":""}]}`,
			ErrServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				respondJSON(w, tc.statusCode, tc.body)
			}))
			defer srv.Close()

			_, err := newTestClient(t, srv).SendMessage(context.Background(), &SendMessageRequest{Message: validMessage()})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, tc.sentinel) {
				t.Errorf("errors.Is(%v) = false", tc.sentinel)
			}
			var apiErr *PauboxError
			if !errors.As(err, &apiErr) {
				t.Fatal("errors.As(*PauboxError) = false")
			}
			if apiErr.StatusCode != tc.statusCode {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tc.statusCode)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// SendBatch — happy path
// ---------------------------------------------------------------------------

func TestSendBatch_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 200, `{"messages":[
			{"sourceTrackingId":"t1","data":"Service OK"},
			{"sourceTrackingId":"t2","data":"Service OK"}
		]}`)
	}))
	defer srv.Close()

	resp, err := newTestClient(t, srv).SendBatch(context.Background(), &SendBatchRequest{
		Messages: []Message{validMessage(), validMessage()},
	})
	if err != nil {
		t.Fatalf("SendBatch() error: %v", err)
	}
	if len(resp.Messages) != 2 {
		t.Fatalf("len(Messages) = %d, want 2", len(resp.Messages))
	}
	if resp.Messages[0].SourceTrackingID != "t1" {
		t.Errorf("Messages[0].SourceTrackingID = %q, want t1", resp.Messages[0].SourceTrackingID)
	}
	if resp.Messages[1].SourceTrackingID != "t2" {
		t.Errorf("Messages[1].SourceTrackingID = %q, want t2", resp.Messages[1].SourceTrackingID)
	}
}

func TestSendBatch_SendsCorrectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		respondJSON(w, 200, `{"messages":[]}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).SendBatch(context.Background(), &SendBatchRequest{
		Messages: []Message{validMessage()},
	})
	if gotPath != "/testuser/bulk_messages" {
		t.Errorf("path = %q, want /testuser/bulk_messages", gotPath)
	}
}

func TestSendBatch_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     *SendBatchRequest
		wantErr string
	}{
		{"nil", nil, "nil"},
		{"empty messages", &SendBatchRequest{Messages: []Message{}}, "at least one"},
		{
			"second message invalid",
			&SendBatchRequest{Messages: []Message{
				validMessage(),
				// Missing recipients.
				{Headers: MessageHeaders{From: "f@x.com", Subject: "s"}, Content: MessageContent{PlainText: Ptr("b")}},
			}},
			"message[1]",
		},
	}

	c, _ := New("k", "u")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := c.SendBatch(context.Background(), tc.req)
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
// GetEmailDisposition — happy path
// ---------------------------------------------------------------------------

func TestGetEmailDisposition_HappyPath(t *testing.T) {
	const body = `{
		"sourceTrackingId":"track-xyz",
		"data":{
			"message":{
				"id":"msg-1",
				"message_deliveries":[{
					"recipient":"r@example.com",
					"status":{
						"deliveryStatus":"delivered",
						"deliveryTime":"Mon, 01 Jan 2024 12:00:00 +0000",
						"openedStatus":"opened",
						"openedTime":"Mon, 01 Jan 2024 12:05:00 +0000"
					}
				}],
				"total_opens":1,
				"distinct_opens":1,
				"total_click_count":2,
				"clicks_per_link":[{"click_count":2,"target_url":"https://example.com"}],
				"unsubscribed":false
			}
		}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("sourceTrackingId") == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		respondJSON(w, 200, body)
	}))
	defer srv.Close()

	disp, err := newTestClient(t, srv).GetEmailDisposition(context.Background(), "track-xyz")
	if err != nil {
		t.Fatalf("GetEmailDisposition() error: %v", err)
	}

	if disp.SourceTrackingID != "track-xyz" {
		t.Errorf("SourceTrackingID = %q", disp.SourceTrackingID)
	}
	if len(disp.Data.Message.MessageDeliveries) != 1 {
		t.Fatalf("deliveries = %d, want 1", len(disp.Data.Message.MessageDeliveries))
	}
	d := disp.Data.Message.MessageDeliveries[0]
	if d.Recipient != "r@example.com" {
		t.Errorf("Recipient = %q", d.Recipient)
	}
	if d.Status.DeliveryStatus != DeliveryStatusDelivered {
		t.Errorf("DeliveryStatus = %q", d.Status.DeliveryStatus)
	}
	if disp.Data.Message.TotalOpens == nil || *disp.Data.Message.TotalOpens != 1 {
		t.Error("TotalOpens not populated")
	}
	if len(disp.Data.Message.ClicksPerLink) != 1 {
		t.Error("ClicksPerLink not populated")
	}
}

func TestGetEmailDisposition_SendsTrackingIDInQuery(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("sourceTrackingId")
		respondJSON(w, 200, `{"sourceTrackingId":"my-id","data":{"message":{"id":"","message_deliveries":[]}}}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).GetEmailDisposition(context.Background(), "my-id")
	if gotQuery != "my-id" {
		t.Errorf("sourceTrackingId query param = %q, want my-id", gotQuery)
	}
}

func TestGetEmailDisposition_Validation(t *testing.T) {
	c, _ := New("k", "u")

	t.Run("empty", func(t *testing.T) {
		_, err := c.GetEmailDisposition(context.Background(), "")
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("whitespace", func(t *testing.T) {
		_, err := c.GetEmailDisposition(context.Background(), "   ")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestGetEmailDisposition_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 404, `{"errors":[{"code":404,"title":"Not Found","details":"Message with this tracking id was not found"}]}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).GetEmailDisposition(context.Background(), "bad-id")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	var apiErr *PauboxError
	if errors.As(err, &apiErr) {
		if apiErr.Details == "" {
			t.Error("Details should be populated from wire error")
		}
	}
}

func TestGetEmailDisposition_Uses_GET(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		respondJSON(w, 200, `{"sourceTrackingId":"x","data":{"message":{"id":"","message_deliveries":[]}}}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).GetEmailDisposition(context.Background(), "x")
	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
}

// ---------------------------------------------------------------------------
// MessageHeaders custom JSON marshalling
// ---------------------------------------------------------------------------

func TestMessageHeaders_MarshalJSON(t *testing.T) {
	h := MessageHeaders{
		Subject:             "Hi",
		From:                "f@x.com",
		ReplyTo:             "r@x.com",
		ListUnsubscribe:     "<mailto:u@x.com>",
		ListUnsubscribePost: "List-Unsubscribe=One-Click",
		CustomHeaders:       map[string]string{"x-foo": "bar", "x-baz": "qux"},
	}

	b, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}

	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	checks := map[string]string{
		"subject":               "Hi",
		"from":                  "f@x.com",
		"reply-to":              "r@x.com",
		"List-Unsubscribe":      "<mailto:u@x.com>",
		"List-Unsubscribe-Post": "List-Unsubscribe=One-Click",
		"x-foo":                 "bar",
		"x-baz":                 "qux",
	}
	for k, want := range checks {
		if m[k] != want {
			t.Errorf("key %q = %q, want %q", k, m[k], want)
		}
	}
}

func TestMessageHeaders_MarshalJSON_OmitsEmpty(t *testing.T) {
	h := MessageHeaders{Subject: "Hi", From: "f@x.com"}
	b, _ := json.Marshal(h)

	var m map[string]string
	_ = json.Unmarshal(b, &m)

	for _, key := range []string{"reply-to", "List-Unsubscribe", "List-Unsubscribe-Post"} {
		if _, ok := m[key]; ok {
			t.Errorf("key %q should be absent when empty", key)
		}
	}
}

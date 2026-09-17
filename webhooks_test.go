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
// ListWebhookEndpoints
// ---------------------------------------------------------------------------

func TestListWebhookEndpoints_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `[{"id":1,"target_url":"https://example.com/hook","events":["inbound_mail_received"],"active":true,"signing_key":"whsec_abc","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}]`)
	}))
	defer srv.Close()

	endpoints, err := newTestClient(t, srv).ListWebhookEndpoints(context.Background())
	if err != nil {
		t.Fatalf("ListWebhookEndpoints() error: %v", err)
	}
	if len(endpoints) != 1 {
		t.Fatalf("len(endpoints) = %d, want 1", len(endpoints))
	}
	if endpoints[0].TargetURL != "https://example.com/hook" {
		t.Errorf("TargetURL = %q, want https://example.com/hook", endpoints[0].TargetURL)
	}
	if endpoints[0].SigningKey != "whsec_abc" {
		t.Errorf("SigningKey = %q, want whsec_abc", endpoints[0].SigningKey)
	}
}

func TestListWebhookEndpoints_SendsCorrectMethodAndPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `[]`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).ListWebhookEndpoints(context.Background())
	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/webhook_endpoints" {
		t.Errorf("path = %q, want /webhook_endpoints", gotPath)
	}
}

func TestListWebhookEndpoints_401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 401, `{"errors":[{"code":401,"title":"Unauthorized","details":"bad key"}]}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).ListWebhookEndpoints(context.Background())
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// CreateWebhookEndpoint
// ---------------------------------------------------------------------------

func TestCreateWebhookEndpoint_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"message":"Webhook created!","data":{"id":42,"target_url":"https://example.com/hook","events":["inbound_mail_received"],"active":true,"signing_key":"whsec_xyz","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}}`)
	}))
	defer srv.Close()

	endpoint, err := newTestClient(t, srv).CreateWebhookEndpoint(context.Background(), &CreateWebhookEndpointRequest{
		TargetURL: "https://example.com/hook",
		Events:    []string{"inbound_mail_received"},
	})
	if err != nil {
		t.Fatalf("CreateWebhookEndpoint() error: %v", err)
	}
	if endpoint.ID != 42 {
		t.Errorf("ID = %d, want 42", endpoint.ID)
	}
	if endpoint.TargetURL != "https://example.com/hook" {
		t.Errorf("TargetURL = %q, want https://example.com/hook", endpoint.TargetURL)
	}
}

func TestCreateWebhookEndpoint_SendsCorrectMethodAndPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"message":"Webhook created!","data":{"id":1}}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).CreateWebhookEndpoint(context.Background(), &CreateWebhookEndpointRequest{
		TargetURL: "https://example.com/hook",
		Events:    []string{"inbound_mail_received"},
	})
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/webhook_endpoints" {
		t.Errorf("path = %q, want /webhook_endpoints", gotPath)
	}
}

func TestCreateWebhookEndpoint_SendsBody(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		respondJSON(w, http.StatusOK, `{"message":"Webhook created!","data":{"id":1}}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).CreateWebhookEndpoint(context.Background(), &CreateWebhookEndpointRequest{
		TargetURL: "https://example.com/hook",
		Events:    []string{"api_mail_log_delivered", "inbound_mail_received"},
	})
	if gotBody["target_url"] != "https://example.com/hook" {
		t.Errorf("target_url = %v, want https://example.com/hook", gotBody["target_url"])
	}
	events, ok := gotBody["events"].([]any)
	if !ok || len(events) != 2 {
		t.Fatalf("events = %v, want 2-element array", gotBody["events"])
	}
	if events[0] != "api_mail_log_delivered" {
		t.Errorf("events[0] = %v, want api_mail_log_delivered", events[0])
	}
}

func TestCreateWebhookEndpoint_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateWebhookEndpointRequest
		wantErr string
	}{
		{"nil request", nil, "nil"},
		{"empty target_url", &CreateWebhookEndpointRequest{TargetURL: "", Events: []string{"inbound_mail_received"}}, "target_url"},
		{"whitespace target_url", &CreateWebhookEndpointRequest{TargetURL: "   ", Events: []string{"inbound_mail_received"}}, "target_url"},
		{"empty events", &CreateWebhookEndpointRequest{TargetURL: "https://example.com", Events: []string{}}, "events"},
		{"nil events", &CreateWebhookEndpointRequest{TargetURL: "https://example.com"}, "events"},
	}

	c, _ := New("k")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := c.CreateWebhookEndpoint(context.Background(), tc.req)
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
// GetWebhookEndpoint
// ---------------------------------------------------------------------------

func TestGetWebhookEndpoint_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"data":{"id":42,"target_url":"https://example.com/hook","events":["inbound_mail_received"],"active":true,"signing_key":"whsec_abc","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}}`)
	}))
	defer srv.Close()

	endpoint, err := newTestClient(t, srv).GetWebhookEndpoint(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetWebhookEndpoint() error: %v", err)
	}
	if endpoint.ID != 42 {
		t.Errorf("ID = %d, want 42", endpoint.ID)
	}
}

func TestGetWebhookEndpoint_SendsCorrectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"data":{"id":42}}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).GetWebhookEndpoint(context.Background(), 42)
	if gotPath != "/webhook_endpoints/42" {
		t.Errorf("path = %q, want /webhook_endpoints/42", gotPath)
	}
}

func TestGetWebhookEndpoint_InvalidID(t *testing.T) {
	c, _ := New("k")
	_, err := c.GetWebhookEndpoint(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error for zero id")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("error %q should mention id", err.Error())
	}
}

func TestGetWebhookEndpoint_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 404, `{"errors":[{"code":404,"title":"Not Found","details":"webhook not found"}]}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).GetWebhookEndpoint(context.Background(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// UpdateWebhookEndpoint
// ---------------------------------------------------------------------------

func TestUpdateWebhookEndpoint_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"message":"Webhook updated!","data":{"id":42,"target_url":"https://new.example.com/hook","events":["api_mail_log_delivered"],"active":false,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z"}}`)
	}))
	defer srv.Close()

	endpoint, err := newTestClient(t, srv).UpdateWebhookEndpoint(context.Background(), 42, &UpdateWebhookEndpointRequest{
		TargetURL: Ptr("https://new.example.com/hook"),
		Active:    Ptr(false),
	})
	if err != nil {
		t.Fatalf("UpdateWebhookEndpoint() error: %v", err)
	}
	if endpoint.TargetURL != "https://new.example.com/hook" {
		t.Errorf("TargetURL = %q, want https://new.example.com/hook", endpoint.TargetURL)
	}
	if endpoint.Active {
		t.Error("Active = true, want false")
	}
}

func TestUpdateWebhookEndpoint_SendsCorrectMethodAndPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"message":"Webhook updated!","data":{"id":42}}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).UpdateWebhookEndpoint(context.Background(), 42, &UpdateWebhookEndpointRequest{
		Active: Ptr(true),
	})
	if gotMethod != http.MethodPatch {
		t.Errorf("method = %q, want PATCH", gotMethod)
	}
	if gotPath != "/webhook_endpoints/42" {
		t.Errorf("path = %q, want /webhook_endpoints/42", gotPath)
	}
}

func TestUpdateWebhookEndpoint_SendsBody(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		respondJSON(w, http.StatusOK, `{"message":"Webhook updated!","data":{"id":42}}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).UpdateWebhookEndpoint(context.Background(), 42, &UpdateWebhookEndpointRequest{
		TargetURL: Ptr("https://new.example.com"),
		Active:    Ptr(false),
	})
	if gotBody["target_url"] != "https://new.example.com" {
		t.Errorf("target_url = %v, want https://new.example.com", gotBody["target_url"])
	}
	if gotBody["active"] != false {
		t.Errorf("active = %v, want false", gotBody["active"])
	}
}

func TestUpdateWebhookEndpoint_OmitsUnsetFields(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		respondJSON(w, http.StatusOK, `{"message":"Webhook updated!","data":{"id":42}}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).UpdateWebhookEndpoint(context.Background(), 42, &UpdateWebhookEndpointRequest{
		Active: Ptr(true),
	})
	if _, ok := gotBody["target_url"]; ok {
		t.Error("target_url should be omitted when not set")
	}
	if _, ok := gotBody["events"]; ok {
		t.Error("events should be omitted when not set")
	}
}

func TestUpdateWebhookEndpoint_Validation(t *testing.T) {
	c, _ := New("k")

	_, err := c.UpdateWebhookEndpoint(context.Background(), 0, &UpdateWebhookEndpointRequest{Active: Ptr(true)})
	if err == nil {
		t.Fatal("expected error for zero id")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("error %q should mention id", err.Error())
	}

	_, err = c.UpdateWebhookEndpoint(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("error %q should mention nil", err.Error())
	}
}

// ---------------------------------------------------------------------------
// DeleteWebhookEndpoint
// ---------------------------------------------------------------------------

func TestDeleteWebhookEndpoint_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"message":"Webhook deleted!","data":{"id":42,"target_url":"https://example.com/hook","events":["inbound_mail_received"],"active":true,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}}`)
	}))
	defer srv.Close()

	endpoint, err := newTestClient(t, srv).DeleteWebhookEndpoint(context.Background(), 42)
	if err != nil {
		t.Fatalf("DeleteWebhookEndpoint() error: %v", err)
	}
	if endpoint.ID != 42 {
		t.Errorf("ID = %d, want 42", endpoint.ID)
	}
}

func TestDeleteWebhookEndpoint_SendsCorrectMethodAndPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"message":"Webhook deleted!","data":{"id":42}}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).DeleteWebhookEndpoint(context.Background(), 42)
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	if gotPath != "/webhook_endpoints/42" {
		t.Errorf("path = %q, want /webhook_endpoints/42", gotPath)
	}
}

func TestDeleteWebhookEndpoint_InvalidID(t *testing.T) {
	c, _ := New("k")
	_, err := c.DeleteWebhookEndpoint(context.Background(), -1)
	if err == nil {
		t.Fatal("expected error for negative id")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("error %q should mention id", err.Error())
	}
}

func TestDeleteWebhookEndpoint_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, 404, `{"errors":[{"code":404,"title":"Not Found","details":"webhook not found"}]}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).DeleteWebhookEndpoint(context.Background(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

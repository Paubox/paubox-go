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

const testEndpointID = "2ec66c21-bf48-48eb-8d28-f80b2d6b77c7"

// newTestWebhooksClient points a WebhooksClient at a test server with retries
// off, so a 429/5xx assertion sees exactly one request.
func newTestWebhooksClient(t *testing.T, srv *httptest.Server) *WebhooksClient {
	t.Helper()
	c, err := NewWebhooks("test-key",
		WithWebhooksBaseURL(srv.URL),
		WithWebhooksRetry(RetryConfig{MaxAttempts: 1}),
	)
	if err != nil {
		t.Fatalf("NewWebhooks() error: %v", err)
	}
	return c
}

// ---------------------------------------------------------------------------
// Construction
// ---------------------------------------------------------------------------

func TestNewWebhooks_RejectsEmptyKey(t *testing.T) {
	for _, key := range []string{"", "   "} {
		if _, err := NewWebhooks(key); err == nil {
			t.Errorf("NewWebhooks(%q) = nil error, want error", key)
		}
	}
}

func TestNewWebhooks_DefaultsToProductionBaseURL(t *testing.T) {
	c, err := NewWebhooks("k")
	if err != nil {
		t.Fatalf("NewWebhooks() error: %v", err)
	}
	if c.baseURL != "https://api.paubox.com/v1/webhooks" {
		t.Errorf("baseURL = %q, want https://api.paubox.com/v1/webhooks", c.baseURL)
	}
}

func TestWithWebhooksBaseURL_TrimsTrailingSlash(t *testing.T) {
	c, err := NewWebhooks("k", WithWebhooksBaseURL("https://example.test/v1/webhooks/"))
	if err != nil {
		t.Fatalf("NewWebhooks() error: %v", err)
	}
	if c.baseURL != "https://example.test/v1/webhooks" {
		t.Errorf("baseURL = %q, want no trailing slash", c.baseURL)
	}
}

// Bearer, not the Email API's "Token token=" — the whole reason this is a
// separate client.
func TestWebhooks_SendsBearerAuthorization(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Authorization")
		respondJSON(w, http.StatusOK, `{"data":[],"page_info":{"count":0,"items":50}}`)
	}))
	defer srv.Close()

	if _, err := newTestWebhooksClient(t, srv).ListWebhookEndpoints(context.Background(), nil); err != nil {
		t.Fatalf("ListWebhookEndpoints() error: %v", err)
	}
	if got != "Bearer test-key" {
		t.Errorf("Authorization = %q, want Bearer test-key", got)
	}
}

// ---------------------------------------------------------------------------
// ListWebhookEndpoints
// ---------------------------------------------------------------------------

func TestListWebhookEndpoints_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"data":[{"id":"`+testEndpointID+`","target_url":"https://example.com/hook","status":"active","events":["forms.submission.created"],"created_at":"2026-10-09T01:43:41.639968+00:00","updated_at":"2026-10-09T01:43:41.639968+00:00"}],"page_info":{"count":1,"items":50}}`)
	}))
	defer srv.Close()

	list, err := newTestWebhooksClient(t, srv).ListWebhookEndpoints(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListWebhookEndpoints() error: %v", err)
	}
	if len(list.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(list.Data))
	}
	if list.Data[0].ID != testEndpointID {
		t.Errorf("ID = %q, want %q", list.Data[0].ID, testEndpointID)
	}
	if list.Data[0].Status != "active" {
		t.Errorf("Status = %q, want active", list.Data[0].Status)
	}
}

func TestListWebhookEndpoints_SendsCorrectMethodAndPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		respondJSON(w, http.StatusOK, `{"data":[],"page_info":{"count":0,"items":50}}`)
	}))
	defer srv.Close()

	if _, err := newTestWebhooksClient(t, srv).ListWebhookEndpoints(context.Background(), nil); err != nil {
		t.Fatalf("ListWebhookEndpoints() error: %v", err)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("method = %s, want GET", gotMethod)
	}
	if gotPath != "/endpoints" {
		t.Errorf("path = %s, want /endpoints", gotPath)
	}
}

func TestListWebhookEndpoints_SendsPagination(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		respondJSON(w, http.StatusOK, `{"data":[],"page_info":{"count":0,"items":1}}`)
	}))
	defer srv.Close()

	_, err := newTestWebhooksClient(t, srv).ListWebhookEndpoints(
		context.Background(), &ListWebhookEndpointsParams{Page: Ptr(2), Items: Ptr(1)},
	)
	if err != nil {
		t.Fatalf("ListWebhookEndpoints() error: %v", err)
	}
	if gotQuery != "items=1&page=2" {
		t.Errorf("query = %q, want items=1&page=2", gotQuery)
	}
}

// count is the total, not the page length — pagination is easy to get wrong
// in a client that assumes otherwise.
func TestListWebhookEndpoints_CountIsTotalNotPageLength(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"data":[{"id":"`+testEndpointID+`","target_url":"https://e.test/h","status":"active","events":["forms.submission.created"],"created_at":"x","updated_at":"x"}],"page_info":{"count":2,"items":1}}`)
	}))
	defer srv.Close()

	list, err := newTestWebhooksClient(t, srv).ListWebhookEndpoints(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListWebhookEndpoints() error: %v", err)
	}
	if len(list.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(list.Data))
	}
	if list.PageInfo.Count != 2 {
		t.Errorf("PageInfo.Count = %d, want 2", list.PageInfo.Count)
	}
}

// ---------------------------------------------------------------------------
// CreateWebhookEndpoint
// ---------------------------------------------------------------------------

func TestCreateWebhookEndpoint_ReturnsSigningSecret(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusCreated, `{"data":{"id":"`+testEndpointID+`","target_url":"https://example.com/hook","status":"active","events":["forms.submission.created"],"created_at":"x","updated_at":"x","signing_secret":"whsec_abc123"},"message":"Store this signing_secret now — it is not shown again."}`)
	}))
	defer srv.Close()

	created, err := newTestWebhooksClient(t, srv).CreateWebhookEndpoint(context.Background(),
		&CreateWebhookEndpointRequest{TargetURL: "https://example.com/hook", Events: []string{"forms.submission.created"}})
	if err != nil {
		t.Fatalf("CreateWebhookEndpoint() error: %v", err)
	}
	if created.SigningSecret != "whsec_abc123" {
		t.Errorf("SigningSecret = %q, want whsec_abc123", created.SigningSecret)
	}
	if created.ID != testEndpointID {
		t.Errorf("ID = %q, want %q", created.ID, testEndpointID)
	}
}

func TestCreateWebhookEndpoint_SendsMethodPathAndBody(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		respondJSON(w, http.StatusCreated, `{"data":{"id":"`+testEndpointID+`","signing_secret":"s"}}`)
	}))
	defer srv.Close()

	_, err := newTestWebhooksClient(t, srv).CreateWebhookEndpoint(context.Background(),
		&CreateWebhookEndpointRequest{TargetURL: "https://example.com/hook", Events: []string{"forms.submission.created"}})
	if err != nil {
		t.Fatalf("CreateWebhookEndpoint() error: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/endpoints" {
		t.Errorf("got %s %s, want POST /endpoints", gotMethod, gotPath)
	}
	if gotBody["target_url"] != "https://example.com/hook" {
		t.Errorf("target_url = %v", gotBody["target_url"])
	}
}

func TestCreateWebhookEndpoint_ValidatesLocally(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("no request should be sent")
	}))
	defer srv.Close()
	c := newTestWebhooksClient(t, srv)

	if _, err := c.CreateWebhookEndpoint(context.Background(), nil); err == nil {
		t.Error("nil request: want error")
	}
	if _, err := c.CreateWebhookEndpoint(context.Background(),
		&CreateWebhookEndpointRequest{TargetURL: "  ", Events: []string{"forms.submission.created"}}); err == nil {
		t.Error("blank target_url: want error")
	}
	if _, err := c.CreateWebhookEndpoint(context.Background(),
		&CreateWebhookEndpointRequest{TargetURL: "https://e.test/h"}); err == nil {
		t.Error("empty events: want error")
	}
}

// Unknown events are the service's call, not the SDK's: the catalog grows
// without an SDK release, so the client must forward whatever it is given.
func TestCreateWebhookEndpoint_DoesNotValidateEventNames(t *testing.T) {
	var gotEvents []any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotEvents, _ = body["events"].([]any)
		respondJSON(w, http.StatusCreated, `{"data":{"id":"`+testEndpointID+`"}}`)
	}))
	defer srv.Close()

	_, err := newTestWebhooksClient(t, srv).CreateWebhookEndpoint(context.Background(),
		&CreateWebhookEndpointRequest{TargetURL: "https://e.test/h", Events: []string{"some.future.event"}})
	if err != nil {
		t.Fatalf("CreateWebhookEndpoint() error: %v", err)
	}
	if len(gotEvents) != 1 || gotEvents[0] != "some.future.event" {
		t.Errorf("events = %v, want the value forwarded unchanged", gotEvents)
	}
}

// ---------------------------------------------------------------------------
// GetWebhookEndpoint
// ---------------------------------------------------------------------------

func TestGetWebhookEndpoint_UnwrapsDataAndOmitsSecret(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/endpoints/"+testEndpointID {
			t.Errorf("path = %s", r.URL.Path)
		}
		respondJSON(w, http.StatusOK, `{"data":{"id":"`+testEndpointID+`","target_url":"https://e.test/h","status":"active","events":["forms.submission.created"],"created_at":"x","updated_at":"y"}}`)
	}))
	defer srv.Close()

	ep, err := newTestWebhooksClient(t, srv).GetWebhookEndpoint(context.Background(), testEndpointID)
	if err != nil {
		t.Fatalf("GetWebhookEndpoint() error: %v", err)
	}
	if ep.ID != testEndpointID {
		t.Errorf("ID = %q", ep.ID)
	}
	// WebhookEndpoint has no SigningSecret field at all; only the create
	// result carries one. This pins that split.
	if ep.CreatedAt == ep.UpdatedAt {
		t.Errorf("CreatedAt and UpdatedAt should be distinct in this fixture")
	}
}

// A non-UUID id is refused before a request. Without the guard a value
// carrying ".." or "/" would change which endpoint is called, and the
// Authorization header goes on the same host, so the key would ride along on
// the retargeted request.
func TestWebhookEndpoint_RejectsNonUUIDIDsBeforeAnyRequest(t *testing.T) {
	ids := []string{
		"", "  ", "abc", "1", "../endpoints", "../../v1/events",
		"2ec66c21-bf48-48eb-8d28-f80b2d6b77c7/x", "2ec66c21bf4848eb8d28f80b2d6b77c7",
	}

	for _, id := range ids {
		t.Run(id, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
				t.Errorf("no request should be sent for id %q", id)
			}))
			defer srv.Close()
			c := newTestWebhooksClient(t, srv)

			if _, err := c.GetWebhookEndpoint(context.Background(), id); err == nil {
				t.Errorf("GetWebhookEndpoint(%q): want error", id)
			}
			if _, err := c.UpdateWebhookEndpoint(context.Background(), id,
				&UpdateWebhookEndpointRequest{}); err == nil {
				t.Errorf("UpdateWebhookEndpoint(%q): want error", id)
			}
			if err := c.DeleteWebhookEndpoint(context.Background(), id); err == nil {
				t.Errorf("DeleteWebhookEndpoint(%q): want error", id)
			}
		})
	}
}

func TestGetWebhookEndpoint_AcceptsACanonicalUUID(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		respondJSON(w, http.StatusOK, `{"data":{"id":"2ec66c21-bf48-48eb-8d28-f80b2d6b77c7"}}`)
	}))
	defer srv.Close()

	if _, err := newTestWebhooksClient(t, srv).GetWebhookEndpoint(
		context.Background(), "2EC66C21-bf48-48eb-8d28-f80b2d6b77c7"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Mixed case is accepted and passed through unchanged.
	if want := "/endpoints/2EC66C21-bf48-48eb-8d28-f80b2d6b77c7"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

// ---------------------------------------------------------------------------
// UpdateWebhookEndpoint
// ---------------------------------------------------------------------------

func TestUpdateWebhookEndpoint_SendsOnlySetFields(t *testing.T) {
	var gotMethod string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		respondJSON(w, http.StatusOK, `{"data":{"id":"`+testEndpointID+`","status":"disabled"}}`)
	}))
	defer srv.Close()

	ep, err := newTestWebhooksClient(t, srv).UpdateWebhookEndpoint(context.Background(), testEndpointID,
		&UpdateWebhookEndpointRequest{Status: Ptr("disabled")})
	if err != nil {
		t.Fatalf("UpdateWebhookEndpoint() error: %v", err)
	}
	if gotMethod != http.MethodPatch {
		t.Errorf("method = %s, want PATCH", gotMethod)
	}
	if _, present := gotBody["target_url"]; present {
		t.Error("target_url must be omitted when unset")
	}
	if _, present := gotBody["events"]; present {
		t.Error("events must be omitted when unset")
	}
	if ep.Status != "disabled" {
		t.Errorf("Status = %q, want disabled", ep.Status)
	}
}

func TestUpdateWebhookEndpoint_ValidatesLocally(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("no request should be sent")
	}))
	defer srv.Close()
	c := newTestWebhooksClient(t, srv)

	if _, err := c.UpdateWebhookEndpoint(context.Background(), "", &UpdateWebhookEndpointRequest{}); err == nil {
		t.Error("blank id: want error")
	}
	if _, err := c.UpdateWebhookEndpoint(context.Background(), testEndpointID, nil); err == nil {
		t.Error("nil request: want error")
	}
}

// ---------------------------------------------------------------------------
// DeleteWebhookEndpoint
// ---------------------------------------------------------------------------

func TestDeleteWebhookEndpoint_HandlesEmpty204(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	if err := newTestWebhooksClient(t, srv).DeleteWebhookEndpoint(context.Background(), testEndpointID); err != nil {
		t.Fatalf("DeleteWebhookEndpoint() error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %s, want DELETE", gotMethod)
	}
	if gotPath != "/endpoints/"+testEndpointID {
		t.Errorf("path = %s", gotPath)
	}
}

// ---------------------------------------------------------------------------
// Errors — the service returns {"message": "..."}, not the Email API's array
// ---------------------------------------------------------------------------

func TestWebhooks_ParsesMessageShapedErrors(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		body     string
		sentinel error
		wantIn   string
	}{
		{"non-https target", http.StatusUnprocessableEntity, `{"message":"target_url: must be an https URL"}`, nil, "must be an https URL"},
		{"not entitled", http.StatusForbidden, `{"message":"events: not entitled to 'api_mail_log_delivered'"}`, ErrForbidden, "not entitled"},
		{"not found", http.StatusNotFound, `{"message":"webhook endpoint not found"}`, ErrNotFound, "not found"},
		{"unauthorized", http.StatusUnauthorized, `{"message":"unauthorized"}`, ErrUnauthorized, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				respondJSON(w, tc.status, tc.body)
			}))
			defer srv.Close()

			_, err := newTestWebhooksClient(t, srv).ListWebhookEndpoints(context.Background(), nil)
			if err == nil {
				t.Fatal("want error")
			}
			var pErr *PauboxError
			if !errors.As(err, &pErr) {
				t.Fatalf("error is %T, want *PauboxError", err)
			}
			if pErr.StatusCode != tc.status {
				t.Errorf("StatusCode = %d, want %d", pErr.StatusCode, tc.status)
			}
			if tc.sentinel != nil && !errors.Is(err, tc.sentinel) {
				t.Errorf("errors.Is(err, %v) = false", tc.sentinel)
			}
			if tc.wantIn != "" && !strings.Contains(err.Error(), tc.wantIn) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantIn)
			}
		})
	}
}

// A duplicate target_url is 422, not 409 — mapping keyed on 409 would miss it.
func TestWebhooks_DuplicateTargetURLIs422(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusUnprocessableEntity, `{"message":"target_url: an endpoint already exists for this URL"}`)
	}))
	defer srv.Close()

	_, err := newTestWebhooksClient(t, srv).CreateWebhookEndpoint(context.Background(),
		&CreateWebhookEndpointRequest{TargetURL: "https://e.test/h", Events: []string{"forms.submission.created"}})
	var pErr *PauboxError
	if !errors.As(err, &pErr) || pErr.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("got %v, want *PauboxError with 422", err)
	}
}

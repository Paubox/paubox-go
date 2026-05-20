package paubox

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestClient wires a Client to the given httptest.Server.
func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := New("test-api-key", "testuser",
		WithBaseURL(srv.URL),
		WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	return c
}

// ---------------------------------------------------------------------------
// New()
// ---------------------------------------------------------------------------

func TestNew_EmptyAPIKey(t *testing.T) {
	_, err := New("", "user")
	if err == nil {
		t.Fatal("expected error for empty apiKey")
	}
}

func TestNew_WhitespaceAPIKey(t *testing.T) {
	_, err := New("   ", "user")
	if err == nil {
		t.Fatal("expected error for whitespace apiKey")
	}
}

func TestNew_EmptyUsername(t *testing.T) {
	_, err := New("key", "")
	if err == nil {
		t.Fatal("expected error for empty username")
	}
}

func TestNew_Defaults(t *testing.T) {
	c, err := New("key", "user")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if c.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, defaultBaseURL)
	}
	if c.userAgent != defaultUserAgent {
		t.Errorf("userAgent = %q, want %q", c.userAgent, defaultUserAgent)
	}
	if c.retry.MaxAttempts != defaultRetryConfig.MaxAttempts {
		t.Errorf("MaxAttempts = %d, want %d", c.retry.MaxAttempts, defaultRetryConfig.MaxAttempts)
	}
}

func TestNew_Options(t *testing.T) {
	customUA := "myapp/2.0"
	c, err := New("key", "user",
		WithBaseURL("https://staging.example.com"),
		WithTimeout(10*time.Second),
		WithUserAgent(customUA),
		WithRetry(RetryConfig{MaxAttempts: 5, WaitMin: 50 * time.Millisecond, WaitMax: 1 * time.Second}),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if c.baseURL != "https://staging.example.com" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
	if !strings.Contains(c.userAgent, customUA) {
		t.Errorf("userAgent %q should contain %q", c.userAgent, customUA)
	}
	if !strings.Contains(c.userAgent, defaultUserAgent) {
		t.Errorf("userAgent %q should contain SDK identifier", c.userAgent)
	}
	if c.retry.MaxAttempts != 5 {
		t.Errorf("MaxAttempts = %d, want 5", c.retry.MaxAttempts)
	}
}

func TestNew_WithBaseURL_TrailingSlash(t *testing.T) {
	c, _ := New("key", "user", WithBaseURL("https://example.com/"))
	if strings.HasSuffix(c.baseURL, "/") {
		t.Errorf("baseURL should not have trailing slash, got %q", c.baseURL)
	}
}

// ---------------------------------------------------------------------------
// Authorization header
// ---------------------------------------------------------------------------

func TestClient_AuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sourceTrackingId":"x","data":"Service OK"}`))
	}))
	defer srv.Close()

	c, err := New("my-secret-key", "user",
		WithBaseURL(srv.URL),
		WithHTTPClient(srv.Client()),
	)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = c.SendMessage(context.Background(), &SendMessageRequest{
		Message: validMessage(),
	})

	const want = "Token token=my-secret-key"
	if gotAuth != want {
		t.Errorf("Authorization = %q, want %q", gotAuth, want)
	}
}

// ---------------------------------------------------------------------------
// Retry behaviour
// ---------------------------------------------------------------------------

func TestClient_RetryOn5xx_GET(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sourceTrackingId":"ok","data":{"message":{"id":"","message_deliveries":[]}}}`))
	}))
	defer srv.Close()

	c, _ := New("k", "u", WithBaseURL(srv.URL), WithRetry(RetryConfig{
		MaxAttempts: 3, WaitMin: 1 * time.Millisecond, WaitMax: 5 * time.Millisecond,
	}))

	_, err := c.GetEmailDisposition(context.Background(), "track-1")
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestClient_RetryOn429_HonoursRetryAfter(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sourceTrackingId":"ok","data":{"message":{"id":"","message_deliveries":[]}}}`))
	}))
	defer srv.Close()

	c, _ := New("k", "u", WithBaseURL(srv.URL), WithRetry(RetryConfig{
		MaxAttempts: 3, WaitMin: 1 * time.Millisecond, WaitMax: 5 * time.Millisecond,
	}))

	_, err := c.GetEmailDisposition(context.Background(), "track-1")
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2", calls)
	}
}

func TestClient_NoRetryOnPOST_ByDefault(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c, _ := New("k", "u", WithBaseURL(srv.URL), WithRetry(RetryConfig{
		MaxAttempts: 3, WaitMin: 1 * time.Millisecond, WaitMax: 5 * time.Millisecond,
		RetryNonIdempotent: false,
	}))

	_, _ = c.SendMessage(context.Background(), &SendMessageRequest{Message: validMessage()})

	if calls != 1 {
		t.Errorf("POST was retried: calls = %d, want 1", calls)
	}
}

func TestClient_RetryNonIdempotent_WhenEnabled(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sourceTrackingId":"ok","data":"Service OK"}`))
	}))
	defer srv.Close()

	c, _ := New("k", "u", WithBaseURL(srv.URL), WithRetry(RetryConfig{
		MaxAttempts: 3, WaitMin: 1 * time.Millisecond, WaitMax: 5 * time.Millisecond,
		RetryNonIdempotent: true,
	}))

	_, err := c.SendMessage(context.Background(), &SendMessageRequest{Message: validMessage()})
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, _ := New("k", "u", WithBaseURL(srv.URL))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	_, err := c.GetEmailDisposition(ctx, "track-1")
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}

// ---------------------------------------------------------------------------
// endpointURL
// ---------------------------------------------------------------------------

func TestClient_EndpointURL(t *testing.T) {
	c, _ := New("k", "myuser")
	got := c.endpointURL("/messages")
	want := "https://api.paubox.net/v1/myuser/messages"
	if got != want {
		t.Errorf("endpointURL = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Ptr helper
// ---------------------------------------------------------------------------

func TestPtr_String(t *testing.T) {
	s := "hello"
	p := Ptr(s)
	if p == nil || *p != s {
		t.Errorf("Ptr(%q) = %v", s, p)
	}
}

func TestPtr_Int(t *testing.T) {
	i := 42
	p := Ptr(i)
	if p == nil || *p != i {
		t.Errorf("Ptr(%d) = %v", i, p)
	}
}

func TestPtr_Bool(t *testing.T) {
	p := Ptr(true)
	if p == nil || !*p {
		t.Error("Ptr(true) returned unexpected value")
	}
}

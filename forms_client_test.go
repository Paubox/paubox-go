package paubox

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestFormsClient wires a FormsClient (with an API key) to the given
// httptest.Server. Retries are disabled so error tests stay fast.
func newTestFormsClient(t *testing.T, srv *httptest.Server) *FormsClient {
	t.Helper()
	c, err := NewForms("test-key",
		WithFormsBaseURL(srv.URL),
		WithFormsTimeout(5*time.Second),
		WithFormsRetry(RetryConfig{MaxAttempts: 1}),
	)
	if err != nil {
		t.Fatalf("NewForms() error: %v", err)
	}
	return c
}

// newTestPublicFormsClient wires a keyless (public-only) FormsClient to the
// given httptest.Server.
func newTestPublicFormsClient(t *testing.T, srv *httptest.Server) *FormsClient {
	t.Helper()
	c, err := NewForms("",
		WithFormsBaseURL(srv.URL),
		WithFormsTimeout(5*time.Second),
		WithFormsRetry(RetryConfig{MaxAttempts: 1}),
	)
	if err != nil {
		t.Fatalf("NewForms() error: %v", err)
	}
	return c
}

// ---------------------------------------------------------------------------
// NewForms()
// ---------------------------------------------------------------------------

func TestNewForms_Defaults(t *testing.T) {
	c, err := NewForms("key")
	if err != nil {
		t.Fatalf("NewForms() error: %v", err)
	}
	if c.baseURL != defaultFormsBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, defaultFormsBaseURL)
	}
	if c.userAgent != defaultUserAgent {
		t.Errorf("userAgent = %q, want %q", c.userAgent, defaultUserAgent)
	}
	if c.retry.MaxAttempts != defaultRetryConfig.MaxAttempts {
		t.Errorf("MaxAttempts = %d, want %d", c.retry.MaxAttempts, defaultRetryConfig.MaxAttempts)
	}
	if c.httpClient.Timeout != defaultTimeout {
		t.Errorf("timeout = %v, want %v", c.httpClient.Timeout, defaultTimeout)
	}

	tr, ok := c.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport is %T, want *http.Transport", c.httpClient.Transport)
	}
	if tr.TLSClientConfig == nil || tr.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Error("default transport must enforce TLS 1.2 minimum")
	}
	if tr.TLSClientConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify must never be set")
	}
}

func TestNewForms_EmptyKeyIsPublicOnly(t *testing.T) {
	c, err := NewForms("")
	if err != nil {
		t.Fatalf("NewForms(\"\") error: %v", err)
	}
	if c.apiKey != "" {
		t.Errorf("apiKey = %q, want empty", c.apiKey)
	}
}

func TestNewForms_WhitespaceKeyTreatedAsEmpty(t *testing.T) {
	c, err := NewForms("   \t")
	if err != nil {
		t.Fatalf("NewForms(whitespace) error: %v", err)
	}
	if c.apiKey != "" {
		t.Errorf("apiKey = %q, want empty for whitespace-only key", c.apiKey)
	}
}

func TestNewForms_Options(t *testing.T) {
	customUA := "myapp/2.0"
	c, err := NewForms("key",
		WithFormsBaseURL("https://staging.example.com"),
		WithFormsTimeout(10*time.Second),
		WithFormsUserAgent(customUA),
		WithFormsRetry(RetryConfig{MaxAttempts: 5, WaitMin: 50 * time.Millisecond, WaitMax: 1 * time.Second}),
	)
	if err != nil {
		t.Fatalf("NewForms() error: %v", err)
	}
	if c.baseURL != "https://staging.example.com" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
	if c.httpClient.Timeout != 10*time.Second {
		t.Errorf("timeout = %v, want 10s", c.httpClient.Timeout)
	}
	if !strings.HasPrefix(c.userAgent, customUA) {
		t.Errorf("userAgent %q should start with %q", c.userAgent, customUA)
	}
	if !strings.Contains(c.userAgent, defaultUserAgent) {
		t.Errorf("userAgent %q should contain SDK identifier", c.userAgent)
	}
	if c.retry.MaxAttempts != 5 {
		t.Errorf("MaxAttempts = %d, want 5", c.retry.MaxAttempts)
	}
}

func TestNewForms_WithFormsBaseURL_TrailingSlash(t *testing.T) {
	c, _ := NewForms("key", WithFormsBaseURL("https://example.com/"))
	if strings.HasSuffix(c.baseURL, "/") {
		t.Errorf("baseURL should not have trailing slash, got %q", c.baseURL)
	}
}

func TestNewForms_WithFormsHTTPClient(t *testing.T) {
	hc := &http.Client{Timeout: 3 * time.Second}
	c, _ := NewForms("key", WithFormsHTTPClient(hc))
	if c.httpClient != hc {
		t.Error("WithFormsHTTPClient should replace the default HTTP client")
	}
}

// ---------------------------------------------------------------------------
// Authorization header
// ---------------------------------------------------------------------------

func TestFormsClient_BearerAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		respondJSON(w, http.StatusOK, `{}`)
	}))
	defer srv.Close()

	c, err := NewForms("my-forms-key",
		WithFormsBaseURL(srv.URL),
		WithFormsHTTPClient(srv.Client()),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.doJSON(context.Background(), http.MethodGet, "/api/forms", nil, nil); err != nil {
		t.Fatalf("doJSON() error: %v", err)
	}

	const want = "Bearer my-forms-key"
	if gotAuth != want {
		t.Errorf("Authorization = %q, want %q", gotAuth, want)
	}
	if strings.Contains(gotAuth, "Token token=") {
		t.Errorf("Forms client must not use the Email API auth format, got %q", gotAuth)
	}
}

func TestFormsClient_NoAuthorizationHeaderWhenKeyless(t *testing.T) {
	authPresent := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, authPresent = r.Header["Authorization"]
		respondJSON(w, http.StatusOK, `{}`)
	}))
	defer srv.Close()

	c := newTestPublicFormsClient(t, srv)
	if err := c.doJSON(context.Background(), http.MethodGet, "/public/form_data/abc", nil, nil); err != nil {
		t.Fatalf("doJSON() error: %v", err)
	}
	if authPresent {
		t.Error("keyless client must not send an Authorization header")
	}
}

// ---------------------------------------------------------------------------
// requireKey
// ---------------------------------------------------------------------------

func TestFormsClient_RequireKey(t *testing.T) {
	c, _ := NewForms("")
	err := c.requireKey("ListForms")
	if err == nil {
		t.Fatal("expected error from requireKey on keyless client")
	}
	const want = "paubox: ListForms: an API key with the forms scope is required"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}

	keyed, _ := NewForms("key")
	if err := keyed.requireKey("ListForms"); err != nil {
		t.Errorf("requireKey on keyed client = %v, want nil", err)
	}
}

// ---------------------------------------------------------------------------
// Retry behaviour
// ---------------------------------------------------------------------------

func TestFormsClient_RetryOn5xx_GET(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		respondJSON(w, http.StatusOK, `{}`)
	}))
	defer srv.Close()

	c, _ := NewForms("k", WithFormsBaseURL(srv.URL), WithFormsRetry(RetryConfig{
		MaxAttempts: 3, WaitMin: 1 * time.Millisecond, WaitMax: 5 * time.Millisecond,
	}))

	if err := c.doJSON(context.Background(), http.MethodGet, "/api/forms", nil, nil); err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestFormsClient_RetryOn429_HonoursRetryAfter(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		respondJSON(w, http.StatusOK, `{}`)
	}))
	defer srv.Close()

	c, _ := NewForms("k", WithFormsBaseURL(srv.URL), WithFormsRetry(RetryConfig{
		MaxAttempts: 3, WaitMin: 1 * time.Millisecond, WaitMax: 5 * time.Millisecond,
	}))

	if err := c.doJSON(context.Background(), http.MethodGet, "/api/forms", nil, nil); err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2", calls)
	}
}

func TestFormsClient_NoRetryOnPOST_ByDefault(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c, _ := NewForms("k", WithFormsBaseURL(srv.URL), WithFormsRetry(RetryConfig{
		MaxAttempts: 3, WaitMin: 1 * time.Millisecond, WaitMax: 5 * time.Millisecond,
		RetryNonIdempotent: false,
	}))

	err := c.doJSON(context.Background(), http.MethodPost, "/api/forms", map[string]string{"title": "t"}, nil)
	if err == nil {
		t.Fatal("expected error from 500 response")
	}
	if calls != 1 {
		t.Errorf("POST was retried: calls = %d, want 1", calls)
	}
}

func TestFormsClient_NoRetryOnPUT_ByDefault(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c, _ := NewForms("k", WithFormsBaseURL(srv.URL), WithFormsRetry(RetryConfig{
		MaxAttempts: 3, WaitMin: 1 * time.Millisecond, WaitMax: 5 * time.Millisecond,
		RetryNonIdempotent: false,
	}))

	err := c.doJSON(context.Background(), http.MethodPut, "/api/forms/abc", map[string]string{"title": "t"}, nil)
	if err == nil {
		t.Fatal("expected error from 500 response")
	}
	if calls != 1 {
		t.Errorf("PUT was retried: calls = %d, want 1", calls)
	}
}

func TestFormsClient_RetryNonIdempotent_WhenEnabled(t *testing.T) {
	var bodies [][]byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, b)
		if len(bodies) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		respondJSON(w, http.StatusOK, `{}`)
	}))
	defer srv.Close()

	c, _ := NewForms("k", WithFormsBaseURL(srv.URL), WithFormsRetry(RetryConfig{
		MaxAttempts: 3, WaitMin: 1 * time.Millisecond, WaitMax: 5 * time.Millisecond,
		RetryNonIdempotent: true,
	}))

	reqBody := map[string]string{"title": "Retry Me", "description": "same bytes each attempt"}
	if err := c.doJSON(context.Background(), http.MethodPost, "/api/forms", reqBody, nil); err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if len(bodies) != 3 {
		t.Fatalf("calls = %d, want 3", len(bodies))
	}

	// The request body must be replayed intact on every retry attempt.
	if len(bodies[0]) == 0 {
		t.Fatal("first attempt received an empty body")
	}
	for i, b := range bodies[1:] {
		if !bytes.Equal(b, bodies[0]) {
			t.Errorf("attempt %d body = %q, want identical to first attempt %q", i+2, b, bodies[0])
		}
	}
}

func TestFormsClient_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, _ := NewForms("k", WithFormsBaseURL(srv.URL))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	if err := c.doJSON(ctx, http.MethodGet, "/api/forms", nil, nil); err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}

// ---------------------------------------------------------------------------
// parseFormsAPIError
// ---------------------------------------------------------------------------

func TestParseFormsAPIError_MessageEnvelope(t *testing.T) {
	raw := []byte(`{"message":"Form not found"}`)
	e := parseFormsAPIError(http.StatusNotFound, "req-1", raw)

	if e.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", e.StatusCode)
	}
	if e.Title != "Form not found" {
		t.Errorf("Title = %q, want %q", e.Title, "Form not found")
	}
	if e.RequestID != "req-1" {
		t.Errorf("RequestID = %q, want %q", e.RequestID, "req-1")
	}
	if !bytes.Equal(e.Raw, raw) {
		t.Errorf("Raw = %q, want original body", e.Raw)
	}
	if !errors.Is(e, ErrNotFound) {
		t.Error("errors.Is(e, ErrNotFound) = false, want true")
	}
}

func TestParseFormsAPIError_EmailEnvelopeFallback(t *testing.T) {
	raw := []byte(`{"errors":[{"code":401,"title":"unauthorized","details":"bad key"}]}`)
	e := parseFormsAPIError(http.StatusUnauthorized, "", raw)

	if e.Title != "unauthorized" {
		t.Errorf("Title = %q, want %q", e.Title, "unauthorized")
	}
	if e.Details != "bad key" {
		t.Errorf("Details = %q, want %q", e.Details, "bad key")
	}
	if e.Code != "401" {
		t.Errorf("Code = %q, want %q", e.Code, "401")
	}
	if !errors.Is(e, ErrUnauthorized) {
		t.Error("errors.Is(e, ErrUnauthorized) = false, want true")
	}
}

func TestParseFormsAPIError_ErrorEnvelope(t *testing.T) {
	// The CSV export returns {"error": "..."} instead of {"message": "..."}.
	raw := []byte(`{"error":"Missing form_json for form"}`)
	e := parseFormsAPIError(http.StatusInternalServerError, "", raw)

	if e.Title != "Missing form_json for form" {
		t.Errorf("Title = %q, want %q", e.Title, "Missing form_json for form")
	}
	if !errors.Is(e, ErrServerError) {
		t.Error("errors.Is(e, ErrServerError) = false, want true")
	}
	if !bytes.Equal(e.Raw, raw) {
		t.Errorf("Raw = %q, want original body", e.Raw)
	}
}

func TestParseFormsAPIError_PlainTextFallback(t *testing.T) {
	raw := []byte("No submissions found for this form")
	e := parseFormsAPIError(http.StatusNotFound, "", raw)

	if e.Title != http.StatusText(http.StatusNotFound) {
		t.Errorf("Title = %q, want status text %q", e.Title, http.StatusText(http.StatusNotFound))
	}
	if !bytes.Equal(e.Raw, raw) {
		t.Errorf("Raw = %q, want original body", e.Raw)
	}
}

func TestParseFormsAPIError_EmptyBody(t *testing.T) {
	e := parseFormsAPIError(http.StatusInternalServerError, "", nil)

	if e.Title != http.StatusText(http.StatusInternalServerError) {
		t.Errorf("Title = %q, want status text", e.Title)
	}
	if !errors.Is(e, ErrServerError) {
		t.Error("errors.Is(e, ErrServerError) = false, want true")
	}
}

// ---------------------------------------------------------------------------
// doJSON error mapping end-to-end
// ---------------------------------------------------------------------------

func TestFormsClient_DoJSON_ErrorMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Request-Id", "req-42")
		respondJSON(w, http.StatusUnauthorized, `{"message":"Invalid API key"}`)
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	err := c.doJSON(context.Background(), http.MethodGet, "/api/forms", nil, nil)
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("errors.Is(err, ErrUnauthorized) = false, want true; err = %v", err)
	}

	var apiErr *PauboxError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is %T, want *PauboxError", err)
	}
	if apiErr.Title != "Invalid API key" {
		t.Errorf("Title = %q, want %q", apiErr.Title, "Invalid API key")
	}
	if apiErr.RequestID != "req-42" {
		t.Errorf("RequestID = %q, want %q", apiErr.RequestID, "req-42")
	}
}

func TestFormsClient_DoJSON_DecodesResponse(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		respondJSON(w, http.StatusOK, `{"active_form_count":3,"total_submission_count":12,"submissions_last_7_days":4}`)
	}))
	defer srv.Close()

	c := newTestFormsClient(t, srv)
	var stats FormStats
	if err := c.doJSON(context.Background(), http.MethodGet, "/api/forms/stats", nil, &stats); err != nil {
		t.Fatalf("doJSON() error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/api/forms/stats" {
		t.Errorf("request = %s %s, want GET /api/forms/stats", gotMethod, gotPath)
	}
	if stats.ActiveFormCount != 3 || stats.TotalSubmissionCount != 12 || stats.SubmissionsLast7Days != 4 {
		t.Errorf("stats = %+v", stats)
	}
}

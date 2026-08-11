package paubox

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// defaultFormsBaseURL is the production base URL for the Paubox Forms API.
//
// The gateway is assumed to forward the remainder of the path unchanged, so
// SDK paths are the Forms service routes verbatim (e.g. /api/forms,
// /public/form_data/{id}). Point [WithFormsBaseURL] at a different mount —
// e.g. http://localhost:3000 for a local Forms service — to override.
const defaultFormsBaseURL = "https://api.paubox.com/v1/forms"

// FormsClient is the Paubox Forms API client. Create one with [NewForms] and
// reuse it across requests — it is safe for concurrent use.
//
// Forms endpoints authenticate with a scoped API key carrying the "forms"
// scope, sent as "Authorization: Bearer <key>". This differs from the Email
// API's "Token token=" format.
type FormsClient struct {
	apiKey     string // optional; empty = public-only client
	baseURL    string
	userAgent  string
	httpClient *http.Client
	retry      RetryConfig
}

// FormsOption is a functional option for configuring a [FormsClient].
type FormsOption func(*FormsClient)

// WithFormsBaseURL overrides the Forms API base URL. A trailing slash is
// trimmed. Useful for testing or pointing at a locally running Forms service.
func WithFormsBaseURL(url string) FormsOption {
	return func(c *FormsClient) {
		c.baseURL = strings.TrimRight(url, "/")
	}
}

// WithFormsHTTPClient replaces the default HTTP client. When using a custom
// client, callers are responsible for maintaining a minimum TLS version of
// 1.2 and for not setting InsecureSkipVerify.
func WithFormsHTTPClient(hc *http.Client) FormsOption {
	return func(c *FormsClient) {
		c.httpClient = hc
	}
}

// WithFormsTimeout sets the per-request timeout on the default HTTP client.
// Ignored when [WithFormsHTTPClient] is also provided.
func WithFormsTimeout(d time.Duration) FormsOption {
	return func(c *FormsClient) {
		c.httpClient.Timeout = d
	}
}

// WithFormsRetry configures retry behaviour. Pass a zero [RetryConfig] to
// disable retries entirely (MaxAttempts: 1).
func WithFormsRetry(cfg RetryConfig) FormsOption {
	return func(c *FormsClient) {
		c.retry = cfg
	}
}

// WithFormsUserAgent prepends a custom token to the User-Agent header. The
// Paubox SDK identifier is always appended after the custom value.
func WithFormsUserAgent(ua string) FormsOption {
	return func(c *FormsClient) {
		c.userAgent = ua + " " + defaultUserAgent
	}
}

// NewForms creates a new Paubox Forms API client.
//
// apiKey is a scoped API key carrying the "forms" scope. It may be empty:
// that yields a public-only client on which only the public endpoints
// (GetPublicForm, SubmitForm) work; protected methods fail fast with an
// error. A whitespace-only key is treated as empty.
func NewForms(apiKey string, opts ...FormsOption) (*FormsClient, error) {
	if strings.TrimSpace(apiKey) == "" {
		apiKey = ""
	}

	c := &FormsClient{
		apiKey:    apiKey,
		baseURL:   defaultFormsBaseURL,
		userAgent: defaultUserAgent,
		retry:     defaultRetryConfig,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			},
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// requireKey fails fast when a protected Forms method is called on a
// public-only (keyless) client. method is used in the error prefix.
func (c *FormsClient) requireKey(method string) error {
	if c.apiKey == "" {
		return fmt.Errorf("paubox: %s: an API key with the forms scope is required", method)
	}
	return nil
}

// do executes one HTTP request against the Forms service with automatic
// authentication and retry. It is the single choke-point for all outbound
// Forms API calls; retry semantics are shared with the Email client via
// doHTTP (GET/DELETE retry on 429 and 5xx; POST/PUT do not unless
// RetryNonIdempotent is set).
func (c *FormsClient) do(ctx context.Context, method, path string, body io.Reader, contentType string) (*http.Response, error) {
	// Authorization — Bearer with a scoped key; set only when a key exists
	// so public-only clients send no Authorization header at all.
	var authorization string
	if c.apiKey != "" {
		authorization = "Bearer " + c.apiKey
	}
	return doHTTP(ctx, c.httpClient, c.retry, method, c.baseURL+path, body, contentType, authorization, c.userAgent)
}

// doJSON marshals reqBody to JSON, sends the request to the given path, and
// unmarshals the response into respBody. A non-2xx status is returned as
// *[PauboxError] via parseFormsAPIError.
func (c *FormsClient) doJSON(ctx context.Context, method, path string, reqBody, respBody any) error {
	var bodyReader io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("paubox: marshalling request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	resp, err := c.do(ctx, method, path, bodyReader, "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck // close-on-defer; read errors already reported by ReadAll above

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("paubox: reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseFormsAPIError(resp.StatusCode, resp.Header.Get("X-Request-Id"), raw)
	}

	if respBody != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, respBody); err != nil {
			return fmt.Errorf("paubox: decoding response: %w", err)
		}
	}
	return nil
}

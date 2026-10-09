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

// defaultWebhooksBaseURL is the production base URL for the Paubox webhooks
// service.
//
// The public gateway exposes only /v1/webhooks/endpoints and rewrites it onto
// the service's own /v1/endpoints, so the producers' event-ingest route is not
// reachable from here. Point [WithWebhooksBaseURL] at a different mount to
// override.
const defaultWebhooksBaseURL = "https://api.paubox.com/v1/webhooks"

// WebhooksClient is the Paubox webhooks API client. Create one with
// [NewWebhooks] and reuse it across requests — it is safe for concurrent use.
//
// Webhook endpoints authenticate with a scoped API key sent as
// "Authorization: Bearer <key>", the same scheme as the Forms API and unlike
// the Email API's "Token token=" format. The key's scopes decide which events
// it may subscribe to; subscribing to an event the key is not scoped for is
// refused with 403.
type WebhooksClient struct {
	apiKey     string
	baseURL    string
	userAgent  string
	httpClient *http.Client
	retry      RetryConfig
}

// WebhooksOption is a functional option for configuring a [WebhooksClient].
type WebhooksOption func(*WebhooksClient)

// WithWebhooksBaseURL overrides the webhooks API base URL. A trailing slash is
// trimmed. Useful for testing or pointing at a non-production environment.
func WithWebhooksBaseURL(url string) WebhooksOption {
	return func(c *WebhooksClient) {
		c.baseURL = strings.TrimRight(url, "/")
	}
}

// WithWebhooksHTTPClient replaces the default HTTP client. When using a custom
// client, callers are responsible for maintaining a minimum TLS version of
// 1.2 and for not setting InsecureSkipVerify.
func WithWebhooksHTTPClient(hc *http.Client) WebhooksOption {
	return func(c *WebhooksClient) {
		c.httpClient = hc
	}
}

// WithWebhooksTimeout sets the per-request timeout on the default HTTP client.
// Ignored when [WithWebhooksHTTPClient] is also provided.
func WithWebhooksTimeout(d time.Duration) WebhooksOption {
	return func(c *WebhooksClient) {
		c.httpClient.Timeout = d
	}
}

// WithWebhooksRetry configures retry behaviour. Pass a zero [RetryConfig] to
// disable retries entirely (MaxAttempts: 1).
func WithWebhooksRetry(cfg RetryConfig) WebhooksOption {
	return func(c *WebhooksClient) {
		c.retry = cfg
	}
}

// WithWebhooksUserAgent prepends a custom token to the User-Agent header. The
// Paubox SDK identifier is always appended after the custom value.
func WithWebhooksUserAgent(ua string) WebhooksOption {
	return func(c *WebhooksClient) {
		c.userAgent = ua + " " + defaultUserAgent
	}
}

// NewWebhooks creates a new Paubox webhooks API client.
//
// apiKey is a scoped API key. Every endpoint here is authenticated, so unlike
// [NewForms] an empty key is an error rather than a public-only client.
func NewWebhooks(apiKey string, opts ...WebhooksOption) (*WebhooksClient, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("paubox: NewWebhooks: apiKey must not be empty")
	}

	c := &WebhooksClient{
		apiKey:    apiKey,
		baseURL:   defaultWebhooksBaseURL,
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

// do executes one HTTP request against the webhooks service with automatic
// authentication and retry, sharing doHTTP's retry semantics with the Email
// and Forms clients (GET/DELETE retry on 429 and 5xx; POST/PATCH do not
// unless RetryNonIdempotent is set).
func (c *WebhooksClient) do(ctx context.Context, method, path string, body io.Reader, contentType string) (*http.Response, error) {
	return doHTTP(ctx, c.httpClient, c.retry, method, c.baseURL+path, body, contentType, "Bearer "+c.apiKey, c.userAgent)
}

// doJSON marshals reqBody to JSON, sends the request to the given path, and
// unmarshals the response into respBody.
//
// Errors are parsed with parseFormsAPIError: the webhooks service returns the
// same {"message": "..."} shape as Forms, not the Email API's errors array.
func (c *WebhooksClient) doJSON(ctx context.Context, method, path string, reqBody, respBody any) error {
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
	defer resp.Body.Close() //nolint:errcheck // close-on-defer; read errors already reported by ReadAll below

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

package paubox

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL   = "https://api.paubox.net/v1"
	defaultTimeout   = 30 * time.Second
	defaultUserAgent = "paubox-go/0.1.1"
)

// Client is the Paubox Email API client. Create one with [New] and reuse it
// across requests — it is safe for concurrent use.
type Client struct {
	apiKey     string
	username   string
	baseURL    string
	userAgent  string
	httpClient *http.Client
	retry      RetryConfig
}

// RetryConfig controls automatic retry behaviour for API requests.
type RetryConfig struct {
	// MaxAttempts is the total number of attempts (first attempt + retries).
	// A value of 0 or 1 means no retries.
	MaxAttempts int

	// WaitMin is the minimum backoff duration between attempts.
	WaitMin time.Duration

	// WaitMax is the maximum backoff duration between attempts.
	WaitMax time.Duration

	// RetryNonIdempotent enables retries for POST and PATCH requests.
	// Disabled by default because those operations are not guaranteed safe
	// to repeat. Enable only when you know the server is idempotent.
	RetryNonIdempotent bool
}

// defaultRetryConfig retries GET requests up to 3 times on 429 and 5xx.
// POST/PATCH are not retried by default.
var defaultRetryConfig = RetryConfig{
	MaxAttempts:        3,
	WaitMin:            100 * time.Millisecond,
	WaitMax:            2 * time.Second,
	RetryNonIdempotent: false,
}

// Option is a functional option for configuring a [Client].
type Option func(*Client)

// WithBaseURL overrides the API base URL. The URL must not have a trailing
// slash. Useful for testing or pointing at a staging environment.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(url, "/")
	}
}

// WithHTTPClient replaces the default HTTP client. When using a custom client,
// callers are responsible for maintaining a minimum TLS version of 1.2 and
// for not setting InsecureSkipVerify.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithTimeout sets the per-request timeout on the default HTTP client.
// Ignored when [WithHTTPClient] is also provided.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithRetry configures retry behaviour. Pass a zero [RetryConfig] to disable
// retries entirely (MaxAttempts: 1).
func WithRetry(cfg RetryConfig) Option {
	return func(c *Client) {
		c.retry = cfg
	}
}

// WithUserAgent prepends a custom token to the User-Agent header. The Paubox
// SDK identifier is always appended after the custom value.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.userAgent = ua + " " + defaultUserAgent
	}
}

// New creates a new Paubox API client.
//
// apiKey is your API key from the Paubox dashboard.
// username is your API username (the endpoint username) from the dashboard.
//
// Both values are required and must be non-empty.
func New(apiKey, username string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("paubox: apiKey must not be empty")
	}
	if strings.TrimSpace(username) == "" {
		return nil, fmt.Errorf("paubox: username must not be empty")
	}

	c := &Client{
		apiKey:    apiKey,
		username:  username,
		baseURL:   defaultBaseURL,
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

// endpointURL builds the full URL for a given path by inserting the username.
//
//	endpointURL("/messages") → "https://api.paubox.net/v1/{username}/messages"
func (c *Client) endpointURL(path string) string {
	return fmt.Sprintf("%s/%s%s", c.baseURL, c.username, path)
}

// do executes one HTTP request with automatic authentication and retry.
// It is the single choke-point for all outbound calls.
//
// body is read into memory once so it can be replayed across retry attempts.
func (c *Client) do(ctx context.Context, method, path string, body io.Reader, contentType string) (*http.Response, error) {
	// Buffer the body once so we can replay it on retries.
	var bodyBytes []byte
	if body != nil {
		b, err := io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("paubox: reading request body: %w", err)
		}
		bodyBytes = b
	}

	maxAttempts := c.retry.MaxAttempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	isIdempotent := method == http.MethodGet || method == http.MethodDelete
	retryEnabled := isIdempotent || c.retry.RetryNonIdempotent

	var lastResp *http.Response

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, c.endpointURL(path), reqBody)
		if err != nil {
			return nil, fmt.Errorf("paubox: building request: %w", err)
		}

		// Authorization — the Paubox format is non-standard; must be set here.
		req.Header.Set("Authorization", "Token token="+c.apiKey)
		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Accept", "application/json")
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if attempt < maxAttempts && retryEnabled {
				c.backoff(ctx, attempt, nil)
				continue
			}
			return nil, fmt.Errorf("paubox: executing request: %w", err)
		}

		// Success or a non-retryable client error — return immediately.
		if resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		// Drain and close so the TCP connection can be reused.
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		lastResp = resp

		if attempt < maxAttempts && retryEnabled {
			c.backoff(ctx, attempt, resp)
			continue
		}
		break // not retrying — either retryEnabled is false or we've exhausted attempts
	}

	// Re-open the last server-error response so the caller can parse it.
	if lastResp != nil {
		lastResp.Body = io.NopCloser(bytes.NewReader(nil))
		return lastResp, nil
	}
	return nil, fmt.Errorf("paubox: request failed after %d attempts", maxAttempts)
}

// backoff sleeps for an exponentially increasing duration before the next
// retry. It honours the Retry-After response header when present.
func (c *Client) backoff(ctx context.Context, attempt int, resp *http.Response) {
	wait := c.retry.WaitMin

	if resp != nil {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if secs, err := strconv.Atoi(ra); err == nil {
				wait = time.Duration(secs) * time.Second
			}
		}
	}

	if wait == c.retry.WaitMin {
		// Exponential: WaitMin × 2^(attempt-1), capped at WaitMax, plus jitter.
		exp := c.retry.WaitMin * (1 << uint(attempt-1))
		if exp > c.retry.WaitMax {
			exp = c.retry.WaitMax
		}
		jitter := time.Duration(rand.Int64N(int64(exp) / 5)) //nolint:gosec // jitter does not require cryptographic randomness
		wait = exp + jitter
	}

	select {
	case <-ctx.Done():
	case <-time.After(wait):
	}
}

// doJSON marshals reqBody to JSON, POSTs/GETs the given path, and unmarshals
// the response into respBody. A non-2xx status is returned as *[PauboxError].
func (c *Client) doJSON(ctx context.Context, method, path string, reqBody, respBody any) error {
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
		return parseAPIError(resp.StatusCode, resp.Header.Get("X-Request-Id"), raw)
	}

	if respBody != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, respBody); err != nil {
			return fmt.Errorf("paubox: decoding response: %w", err)
		}
	}
	return nil
}

// Ptr returns a pointer to v. Use it to set optional pointer-typed fields in
// request structs without declaring a named variable.
//
//	Content: paubox.MessageContent{
//	    PlainText: paubox.Ptr("Hello!"),
//	}
func Ptr[T any](v T) *T { return &v }

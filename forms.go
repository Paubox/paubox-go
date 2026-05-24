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

const defaultFormsBaseURL = "https://apx.paubox.com/forms"

// FormsClient is the Paubox Forms API client. It requires no authentication —
// credentials from [Client] are never sent. Create one with [NewFormsClient]
// and reuse it across requests; it is safe for concurrent use.
type FormsClient struct {
	baseURL    string
	userAgent  string
	httpClient *http.Client
	retry      RetryConfig
}

// FormsOption is a functional option for configuring a [FormsClient].
type FormsOption func(*FormsClient)

// WithFormsBaseURL overrides the Forms API base URL. The URL must not have a
// trailing slash. Useful for testing or pointing at a staging environment.
func WithFormsBaseURL(url string) FormsOption {
	return func(c *FormsClient) {
		c.baseURL = strings.TrimRight(url, "/")
	}
}

// WithFormsHTTPClient replaces the default HTTP client. When using a custom
// client, callers are responsible for maintaining a minimum TLS version of 1.2
// and for not setting InsecureSkipVerify.
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

// NewFormsClient creates a new Paubox Forms API client.
//
// No credentials are required — the Forms API endpoints used here are public.
// Use [New] for authenticated Email API operations; never share API keys with
// this client.
func NewFormsClient(opts ...FormsOption) (*FormsClient, error) {
	c := &FormsClient{
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

// endpointURL builds the full URL for a given path.
func (c *FormsClient) endpointURL(path string) string {
	return fmt.Sprintf("%s%s", c.baseURL, path)
}

// GetForm retrieves the metadata and field schema for a Paubox Form by its UUID.
func (c *FormsClient) GetForm(ctx context.Context, formID string) (*Form, error) {
	if strings.TrimSpace(formID) == "" {
		return nil, fmt.Errorf("paubox: GetForm: formID must not be empty")
	}

	var form Form
	if err := c.doFormsJSON(ctx, http.MethodGet, "/public/form_data/"+formID, nil, &form); err != nil {
		return nil, err
	}
	return &form, nil
}

// SubmitForm submits a response to a Paubox Form.
//
// formID is the UUID of the target form. sub.FormData must not be nil; its
// keys should match the field names defined in the form's schema.
func (c *FormsClient) SubmitForm(ctx context.Context, formID string, sub FormSubmission) (*FormSubmitResponse, error) {
	if strings.TrimSpace(formID) == "" {
		return nil, fmt.Errorf("paubox: SubmitForm: formID must not be empty")
	}
	if sub.FormData == nil {
		return nil, fmt.Errorf("paubox: SubmitForm: FormData must not be nil")
	}

	var resp FormSubmitResponse
	if err := c.doFormsJSON(ctx, http.MethodPost, "/api/forms/"+formID+"/submissions", sub, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// do executes one HTTP request without authentication.
// It is the single choke-point for all FormsClient outbound calls.
//
// body is read into memory once so it can be replayed across retry attempts.
func (c *FormsClient) do(ctx context.Context, method, path string, body io.Reader, contentType string) (*http.Response, error) {
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

		// Success or non-retryable client error — return immediately.
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
		break
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
func (c *FormsClient) backoff(ctx context.Context, attempt int, resp *http.Response) {
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

// doFormsJSON marshals reqBody to JSON, calls the given path, and unmarshals
// the response into respBody. A non-2xx status is returned as *[PauboxError].
func (c *FormsClient) doFormsJSON(ctx context.Context, method, path string, reqBody, respBody any) error {
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
		return parseAPIError(resp.StatusCode, resp.Header.Get("X-Request-Id"), raw)
	}

	if respBody != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, respBody); err != nil {
			return fmt.Errorf("paubox: decoding response: %w", err)
		}
	}
	return nil
}

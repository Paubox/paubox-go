package paubox

// Submission methods for the Paubox Forms API: ListFormSubmissions plus the
// CSV/PDF export endpoints. All of these require a scoped API key. The
// transport lives in forms_client.go and the request/response types in
// forms_types.go.

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// encode renders the params as a URL query string. Zero values are omitted;
// a nil receiver yields an empty string.
func (p *ListFormSubmissionsParams) encode() string {
	if p == nil {
		return ""
	}
	q := url.Values{}
	if p.SubmissionID != "" {
		q.Set("submission_id", p.SubmissionID)
	}
	if p.OrderBy != "" {
		q.Set("order_by", p.OrderBy)
	}
	if p.Order != "" {
		q.Set("order", p.Order)
	}
	if p.Page > 0 {
		q.Set("page", strconv.Itoa(p.Page))
	}
	if p.Items > 0 {
		q.Set("items", strconv.Itoa(p.Items))
	}
	return q.Encode()
}

// ListFormSubmissions lists the submissions received by a form.
//
// params may be nil, in which case the server defaults apply (page 1, 50
// items per page — capped at 100 — ordered by created_at descending).
//
// API: GET /api/forms/{formID}/submissions
func (c *FormsClient) ListFormSubmissions(ctx context.Context, formID string, params *ListFormSubmissionsParams) (*ListFormSubmissionsResponse, error) {
	if err := c.requireKey("ListFormSubmissions"); err != nil {
		return nil, err
	}
	formID = strings.TrimSpace(formID)
	if formID == "" {
		return nil, fmt.Errorf("paubox: ListFormSubmissions: form id must not be empty")
	}

	path := "/api/forms/" + url.PathEscape(formID) + "/submissions"
	if q := params.encode(); q != "" {
		path += "?" + q
	}

	var resp ListFormSubmissionsResponse
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ExportFormSubmissionsCSV exports all submissions for a form as CSV and
// returns the raw file bytes.
//
// The returned bytes may contain PHI; handle and store them accordingly.
//
// API: GET /api/forms/{formID}/submissions/submission-csv
func (c *FormsClient) ExportFormSubmissionsCSV(ctx context.Context, formID string) ([]byte, error) {
	if err := c.requireKey("ExportFormSubmissionsCSV"); err != nil {
		return nil, err
	}
	formID = strings.TrimSpace(formID)
	if formID == "" {
		return nil, fmt.Errorf("paubox: ExportFormSubmissionsCSV: form id must not be empty")
	}

	return c.exportRaw(ctx, "/api/forms/"+url.PathEscape(formID)+"/submissions/submission-csv")
}

// ExportFormSubmissionCSV exports a single submission as CSV and returns the
// raw file bytes.
//
// The returned bytes may contain PHI; handle and store them accordingly.
//
// API: GET /api/forms/{formID}/submissions/submission-csv/{submissionID}
func (c *FormsClient) ExportFormSubmissionCSV(ctx context.Context, formID, submissionID string) ([]byte, error) {
	if err := c.requireKey("ExportFormSubmissionCSV"); err != nil {
		return nil, err
	}
	formID = strings.TrimSpace(formID)
	if formID == "" {
		return nil, fmt.Errorf("paubox: ExportFormSubmissionCSV: form id must not be empty")
	}
	submissionID = strings.TrimSpace(submissionID)
	if submissionID == "" {
		return nil, fmt.Errorf("paubox: ExportFormSubmissionCSV: submission id must not be empty")
	}

	return c.exportRaw(ctx, "/api/forms/"+url.PathEscape(formID)+"/submissions/submission-csv/"+url.PathEscape(submissionID))
}

// ExportFormSubmissionPDF exports a single submission as PDF and returns the
// raw file bytes.
//
// The returned bytes may contain PHI; handle and store them accordingly.
//
// API: GET /api/forms/{formID}/submissions/{submissionID}/submission-pdf
func (c *FormsClient) ExportFormSubmissionPDF(ctx context.Context, formID, submissionID string) ([]byte, error) {
	if err := c.requireKey("ExportFormSubmissionPDF"); err != nil {
		return nil, err
	}
	formID = strings.TrimSpace(formID)
	if formID == "" {
		return nil, fmt.Errorf("paubox: ExportFormSubmissionPDF: form id must not be empty")
	}
	submissionID = strings.TrimSpace(submissionID)
	if submissionID == "" {
		return nil, fmt.Errorf("paubox: ExportFormSubmissionPDF: submission id must not be empty")
	}

	return c.exportRaw(ctx, "/api/forms/"+url.PathEscape(formID)+"/submissions/"+url.PathEscape(submissionID)+"/submission-pdf")
}

// exportRaw GETs the given path and returns the raw response body bytes.
// A non-2xx status is returned as *[PauboxError] via parseFormsAPIError.
// The service's CSV error paths respond with text/plain bodies (and a
// Content-Disposition header) — parseFormsAPIError falls back to the HTTP
// status text for those.
func (c *FormsClient) exportRaw(ctx context.Context, path string) ([]byte, error) {
	resp, err := c.do(ctx, http.MethodGet, path, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // close-on-defer; read errors already reported by ReadAll above

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("paubox: reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseFormsAPIError(resp.StatusCode, resp.Header.Get("X-Request-Id"), raw)
	}
	return raw, nil
}

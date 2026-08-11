package paubox

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ArchiveForm archives a form. Archiving also deactivates the form, so it
// stops accepting submissions until unarchived and reactivated.
//
// API: POST /api/forms/{id}/archive
func (c *FormsClient) ArchiveForm(ctx context.Context, id string) (*FormActionResponse, error) {
	if err := c.requireKey("ArchiveForm"); err != nil {
		return nil, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("paubox: ArchiveForm: id must not be empty")
	}

	var resp FormActionResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/forms/"+url.PathEscape(id)+"/archive", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnarchiveForm unarchives a previously archived form. It does not
// reactivate the form; toggle Active via FormsClient.UpdateForm if needed.
//
// API: POST /api/forms/{id}/unarchive
func (c *FormsClient) UnarchiveForm(ctx context.Context, id string) (*FormActionResponse, error) {
	if err := c.requireKey("UnarchiveForm"); err != nil {
		return nil, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("paubox: UnarchiveForm: id must not be empty")
	}

	var resp FormActionResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/forms/"+url.PathEscape(id)+"/unarchive", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CopyForm duplicates an existing form under a new title and returns the
// newly created form. The copy starts with a fresh submission counter, no
// vanity URL, and is neither archived nor deleted.
//
// API: POST /api/forms/copy
func (c *FormsClient) CopyForm(ctx context.Context, req *CopyFormRequest) (*Form, error) {
	if err := c.requireKey("CopyForm"); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, fmt.Errorf("paubox: CopyForm: request must not be nil")
	}
	if strings.TrimSpace(req.FormID) == "" {
		return nil, fmt.Errorf("paubox: CopyForm: form_id must not be empty")
	}
	if strings.TrimSpace(req.Title) == "" {
		return nil, fmt.Errorf("paubox: CopyForm: title must not be empty")
	}

	var form Form
	if err := c.doJSON(ctx, http.MethodPost, "/api/forms/copy", req, &form); err != nil {
		return nil, err
	}
	return &form, nil
}

// GetFormStats retrieves aggregate form statistics: the active form count,
// the all-time submission count, and submissions over the last 7 days.
//
// params may be nil; the server then scopes the stats to the API key's
// customer.
//
// API: GET /api/forms/stats
func (c *FormsClient) GetFormStats(ctx context.Context, params *FormStatsParams) (*FormStats, error) {
	if err := c.requireKey("GetFormStats"); err != nil {
		return nil, err
	}

	path := "/api/forms/stats"
	if params != nil && params.CustomerID > 0 {
		q := url.Values{}
		q.Set("customer_id", strconv.Itoa(params.CustomerID))
		path += "?" + q.Encode()
	}

	var stats FormStats
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

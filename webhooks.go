package paubox

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// ListWebhookEndpoints returns the webhook endpoints this key can see.
//
// Endpoints carrying an event the key is not scoped for are filtered out by
// the service, so the result is what this key may act on rather than
// everything on the account.
//
// API: GET /endpoints
func (c *WebhooksClient) ListWebhookEndpoints(ctx context.Context, params *ListWebhookEndpointsParams) (*WebhookEndpointList, error) {
	path := "/endpoints"
	if params != nil {
		q := url.Values{}
		if params.Page != nil {
			q.Set("page", strconv.Itoa(*params.Page))
		}
		if params.Items != nil {
			q.Set("items", strconv.Itoa(*params.Items))
		}
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
	}

	var list WebhookEndpointList
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// CreateWebhookEndpoint subscribes a URL to one or more events.
//
// The returned CreatedWebhookEndpoint carries SigningSecret, which the service
// returns only here. Store it before discarding the result.
//
// API: POST /endpoints
func (c *WebhooksClient) CreateWebhookEndpoint(ctx context.Context, req *CreateWebhookEndpointRequest) (*CreatedWebhookEndpoint, error) {
	if req == nil {
		return nil, fmt.Errorf("paubox: CreateWebhookEndpoint: request must not be nil")
	}
	if strings.TrimSpace(req.TargetURL) == "" {
		return nil, fmt.Errorf("paubox: CreateWebhookEndpoint: target_url must not be empty")
	}
	if len(req.Events) == 0 {
		return nil, fmt.Errorf("paubox: CreateWebhookEndpoint: events must not be empty")
	}

	var env createdWebhookEndpointEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/endpoints", req, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// GetWebhookEndpoint retrieves a single webhook endpoint by its UUID.
//
// The response does not include the signing secret; it is available only from
// CreateWebhookEndpoint.
//
// API: GET /endpoints/{id}
func (c *WebhooksClient) GetWebhookEndpoint(ctx context.Context, id string) (*WebhookEndpoint, error) {
	id = strings.TrimSpace(id)
	if err := validateEndpointID("GetWebhookEndpoint", id); err != nil {
		return nil, err
	}

	var env webhookEndpointEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/endpoints/"+url.PathEscape(id), nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// UpdateWebhookEndpoint changes an existing endpoint. Only the fields set on
// req are sent.
//
// API: PATCH /endpoints/{id}
func (c *WebhooksClient) UpdateWebhookEndpoint(ctx context.Context, id string, req *UpdateWebhookEndpointRequest) (*WebhookEndpoint, error) {
	id = strings.TrimSpace(id)
	if err := validateEndpointID("UpdateWebhookEndpoint", id); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, fmt.Errorf("paubox: UpdateWebhookEndpoint: request must not be nil")
	}

	var env webhookEndpointEnvelope
	if err := c.doJSON(ctx, http.MethodPatch, "/endpoints/"+url.PathEscape(id), req, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// DeleteWebhookEndpoint removes an endpoint, stopping every event on it.
//
// The service answers 204 with no body, so there is nothing to return.
//
// API: DELETE /endpoints/{id}
func (c *WebhooksClient) DeleteWebhookEndpoint(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if err := validateEndpointID("DeleteWebhookEndpoint", id); err != nil {
		return err
	}

	return c.doJSON(ctx, http.MethodDelete, "/endpoints/"+url.PathEscape(id), nil, nil)
}

// uuidPattern is the canonical 8-4-4-4-12 hex shape.
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// validateEndpointID rejects an id that is not a UUID before it reaches a URL.
//
// url.PathEscape already turns a "/" into "%2F", so traversal is not reachable
// through it, but that relies on nothing between here and the service decoding
// the escape before routing. Checking the shape removes the dependency, fails
// without a round trip, and matches every other Paubox SDK.
//
// This is stricter than forms.go, which escapes and sends. Every id on this
// service is a UUID and this client is new, so there is no caller relying on
// passing something else.
func validateEndpointID(op, id string) error {
	if id == "" {
		return fmt.Errorf("paubox: %s: id must not be empty", op)
	}
	if !uuidPattern.MatchString(id) {
		return fmt.Errorf("paubox: %s: id must be a UUID, got %q", op, id)
	}
	return nil
}

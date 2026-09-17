package paubox

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// ListWebhookEndpoints returns all webhook endpoints.
//
// API: GET /webhook_endpoints
func (c *Client) ListWebhookEndpoints(ctx context.Context) ([]WebhookEndpoint, error) {
	var endpoints []WebhookEndpoint
	if err := c.doJSON(ctx, http.MethodGet, "/webhook_endpoints", nil, &endpoints); err != nil {
		return nil, err
	}
	return endpoints, nil
}

// CreateWebhookEndpoint creates a new webhook endpoint.
//
// API: POST /webhook_endpoints
func (c *Client) CreateWebhookEndpoint(ctx context.Context, req *CreateWebhookEndpointRequest) (*WebhookEndpoint, error) {
	if req == nil {
		return nil, fmt.Errorf("paubox: CreateWebhookEndpoint: request must not be nil")
	}
	if strings.TrimSpace(req.TargetURL) == "" {
		return nil, fmt.Errorf("paubox: CreateWebhookEndpoint: target_url must not be empty")
	}
	if len(req.Events) == 0 {
		return nil, fmt.Errorf("paubox: CreateWebhookEndpoint: events must not be empty")
	}

	var env webhookEndpointDataEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/webhook_endpoints", req, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// GetWebhookEndpoint retrieves a single webhook endpoint by ID.
//
// API: GET /webhook_endpoints/{id}
func (c *Client) GetWebhookEndpoint(ctx context.Context, id int) (*WebhookEndpoint, error) {
	if id <= 0 {
		return nil, fmt.Errorf("paubox: GetWebhookEndpoint: id must be positive")
	}

	var env webhookEndpointDataEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/webhook_endpoints/"+strconv.Itoa(id), nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// UpdateWebhookEndpoint updates an existing webhook endpoint.
//
// API: PATCH /webhook_endpoints/{id}
func (c *Client) UpdateWebhookEndpoint(ctx context.Context, id int, req *UpdateWebhookEndpointRequest) (*WebhookEndpoint, error) {
	if id <= 0 {
		return nil, fmt.Errorf("paubox: UpdateWebhookEndpoint: id must be positive")
	}
	if req == nil {
		return nil, fmt.Errorf("paubox: UpdateWebhookEndpoint: request must not be nil")
	}

	var env webhookEndpointDataEnvelope
	if err := c.doJSON(ctx, http.MethodPatch, "/webhook_endpoints/"+strconv.Itoa(id), req, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// DeleteWebhookEndpoint deletes a webhook endpoint by ID.
//
// API: DELETE /webhook_endpoints/{id}
func (c *Client) DeleteWebhookEndpoint(ctx context.Context, id int) (*WebhookEndpoint, error) {
	if id <= 0 {
		return nil, fmt.Errorf("paubox: DeleteWebhookEndpoint: id must be positive")
	}

	var env webhookEndpointDataEnvelope
	if err := c.doJSON(ctx, http.MethodDelete, "/webhook_endpoints/"+strconv.Itoa(id), nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

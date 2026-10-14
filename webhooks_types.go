package paubox

// WebhookEndpoint is a configured webhook endpoint.
//
// IDs are UUIDs. Timestamps are RFC 3339 with microsecond precision and an
// explicit +00:00 offset, kept as strings so no precision is lost on the way
// through.
type WebhookEndpoint struct {
	ID        string   `json:"id"`
	TargetURL string   `json:"target_url"`
	Status    string   `json:"status"`
	Events    []string `json:"events"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

// CreatedWebhookEndpoint is the result of [WebhooksClient.CreateWebhookEndpoint].
//
// SigningSecret is returned once, at creation, and never again — not by
// GetWebhookEndpoint and not by ListWebhookEndpoints. Store it when you
// receive it; recovering from a lost secret means replacing the endpoint.
type CreatedWebhookEndpoint struct {
	WebhookEndpoint
	SigningSecret string `json:"signing_secret"`
}

// CreateWebhookEndpointRequest is the request for
// [WebhooksClient.CreateWebhookEndpoint].
//
// Events are not validated client-side on purpose: the catalog is owned by the
// service and grows without an SDK release. An event this key is not scoped
// for is refused with 403, an unrecognised one with 422.
type CreateWebhookEndpointRequest struct {
	TargetURL string   `json:"target_url"`
	Events    []string `json:"events"`
}

// UpdateWebhookEndpointRequest is the request for
// [WebhooksClient.UpdateWebhookEndpoint]. Every field is optional and only the
// ones set are sent, so an update that moves TargetURL leaves Events and
// Status untouched.
//
// Status accepts "active" or "disabled".
type UpdateWebhookEndpointRequest struct {
	TargetURL *string   `json:"target_url,omitempty"`
	Status    *string   `json:"status,omitempty"`
	Events    *[]string `json:"events,omitempty"`
}

// WebhookEndpointPageInfo describes a page of results. Count is the total
// number of endpoints matching the request, not the length of this page.
type WebhookEndpointPageInfo struct {
	Count int `json:"count"`
	Items int `json:"items"`
}

// WebhookEndpointList is the result of [WebhooksClient.ListWebhookEndpoints].
type WebhookEndpointList struct {
	Data     []WebhookEndpoint       `json:"data"`
	PageInfo WebhookEndpointPageInfo `json:"page_info"`
}

// ListWebhookEndpointsParams are the optional query parameters for
// [WebhooksClient.ListWebhookEndpoints].
type ListWebhookEndpointsParams struct {
	Page  *int
	Items *int
}

type webhookEndpointEnvelope struct {
	Data WebhookEndpoint `json:"data"`
}

type createdWebhookEndpointEnvelope struct {
	Message string                 `json:"message,omitempty"`
	Data    CreatedWebhookEndpoint `json:"data"`
}

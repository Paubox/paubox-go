package paubox

// WebhookEndpoint is a configured webhook endpoint.
type WebhookEndpoint struct {
	ID         int      `json:"id"`
	TargetURL  string   `json:"target_url"`
	Events     []string `json:"events"`
	Active     bool     `json:"active"`
	SigningKey string   `json:"signing_key,omitempty"`
	APIKey     *string  `json:"api_key,omitempty"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
}

// CreateWebhookEndpointRequest is the request for [Client.CreateWebhookEndpoint].
type CreateWebhookEndpointRequest struct {
	TargetURL string   `json:"target_url"`
	Events    []string `json:"events"`
	Active    *bool    `json:"active,omitempty"`
}

// UpdateWebhookEndpointRequest is the request for [Client.UpdateWebhookEndpoint].
// Pointer fields allow callers to distinguish between "not set" and zero values.
type UpdateWebhookEndpointRequest struct {
	TargetURL *string   `json:"target_url,omitempty"`
	Events    *[]string `json:"events,omitempty"`
	Active    *bool     `json:"active,omitempty"`
}

type webhookEndpointDataEnvelope struct {
	Message string          `json:"message,omitempty"`
	Data    WebhookEndpoint `json:"data"`
}

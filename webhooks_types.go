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
// Valid Events are api_mail_log_delivered, api_mail_log_opened,
// api_mail_log_temporary_failure and api_mail_log_permanent_failure. Inbound
// mail (email.inbound.received) subscriptions are created in the Paubox
// Dashboard, not through this API.
type CreateWebhookEndpointRequest struct {
	TargetURL string   `json:"target_url"`
	Events    []string `json:"events"`
	Active    *bool    `json:"active,omitempty"`
}

// UpdateWebhookEndpointRequest is the request for [Client.UpdateWebhookEndpoint].
// Pointer fields allow callers to distinguish between "not set" and zero values.
// Events accepts the same values as [CreateWebhookEndpointRequest].
type UpdateWebhookEndpointRequest struct {
	TargetURL *string   `json:"target_url,omitempty"`
	Events    *[]string `json:"events,omitempty"`
	Active    *bool     `json:"active,omitempty"`
}

type webhookEndpointDataEnvelope struct {
	Message string          `json:"message,omitempty"`
	Data    WebhookEndpoint `json:"data"`
}

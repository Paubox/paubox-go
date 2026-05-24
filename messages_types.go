package paubox

import "encoding/json"

// ---------------------------------------------------------------------------
// Shared message components
// ---------------------------------------------------------------------------

// Message is an email message used in [Client.SendMessage] and
// [Client.SendBatch].
type Message struct {
	// Recipients is the list of To: addresses. Required; minimum one entry.
	// Each entry may be a plain address or "Display Name <addr@example.com>".
	Recipients []string `json:"recipients"`

	// BCC is the list of blind carbon-copy recipients.
	BCC []string `json:"bcc,omitempty"`

	// CC is the list of carbon-copy recipients.
	CC []string `json:"cc,omitempty"`

	// Headers contains the required and optional email header fields.
	Headers MessageHeaders `json:"headers"`

	// AllowNonTLS permits delivery over a non-TLS connection when the
	// recipient's mail server does not support TLS. Defaults to false.
	// Enabling this may affect HIPAA compliance — consult your compliance
	// team before setting it to true.
	AllowNonTLS bool `json:"allowNonTLS,omitempty"`

	// ForceSecureNotification forces delivery via the Paubox Secure Message
	// portal instead of directly to the recipient's inbox.
	ForceSecureNotification bool `json:"forceSecureNotification,omitempty"`

	// Content holds the message body. At least one of PlainText or HTML
	// must be non-nil.
	Content MessageContent `json:"content"`

	// Attachments is an optional list of file attachments.
	Attachments []Attachment `json:"attachments,omitempty"`
}

// MessageHeaders holds the email header fields for a message.
// Subject and From are required.
type MessageHeaders struct {
	// Subject is the email subject line. Required.
	Subject string `json:"subject"`

	// From is the sender address. Required. Must match a domain that is
	// verified in your Paubox account.
	From string `json:"from"`

	// ReplyTo is the optional Reply-To address.
	ReplyTo string `json:"reply-to,omitempty"`

	// ListUnsubscribe sets the List-Unsubscribe header (RFC 2369).
	ListUnsubscribe string `json:"List-Unsubscribe,omitempty"`

	// ListUnsubscribePost sets the List-Unsubscribe-Post header (RFC 8058).
	ListUnsubscribePost string `json:"List-Unsubscribe-Post,omitempty"`

	// CustomHeaders holds additional headers. Keys must begin with "x-" or "X-".
	// These are serialised as top-level fields alongside the standard headers.
	CustomHeaders map[string]string `json:"-"`
}

// MarshalJSON serialises MessageHeaders so that custom X- headers are emitted
// as top-level fields in the JSON object, which is what the Paubox API expects.
// Value receiver is required: MessageHeaders is embedded as a value field on
// Message, and encoding/json only invokes MarshalJSON via a pointer receiver
// when the struct field is addressable (which is not the case for a Message
// passed to json.Marshal by value through the wire types).
func (h MessageHeaders) MarshalJSON() ([]byte, error) { //nolint:gocritic // hugeParam: see comment above
	m := make(map[string]string, 5+len(h.CustomHeaders))

	if h.Subject != "" {
		m["subject"] = h.Subject
	}
	if h.From != "" {
		m["from"] = h.From
	}
	if h.ReplyTo != "" {
		m["reply-to"] = h.ReplyTo
	}
	if h.ListUnsubscribe != "" {
		m["List-Unsubscribe"] = h.ListUnsubscribe
	}
	if h.ListUnsubscribePost != "" {
		m["List-Unsubscribe-Post"] = h.ListUnsubscribePost
	}
	for k, v := range h.CustomHeaders {
		m[k] = v
	}

	return json.Marshal(m)
}

// MessageContent holds the message body. At least one field must be non-nil.
type MessageContent struct {
	// PlainText is the text/plain body.
	PlainText *string `json:"text/plain,omitempty"`

	// HTML is the text/html body. May be plain HTML or base64-encoded HTML.
	HTML *string `json:"text/html,omitempty"`
}

// Attachment is a file attached to an email.
type Attachment struct {
	// FileName is the filename displayed to the recipient (e.g. "report.pdf").
	FileName string `json:"fileName"`

	// ContentType is the MIME type (e.g. "application/pdf", "image/png").
	ContentType string `json:"contentType"`

	// Content is the base64-encoded file data.
	Content string `json:"content"`
}

// ---------------------------------------------------------------------------
// Send single message
// ---------------------------------------------------------------------------

// SendMessageRequest is the request for [Client.SendMessage].
type SendMessageRequest struct {
	// Message is the email to send.
	Message Message

	// OverrideOpenTracking enables open tracking for this message.
	OverrideOpenTracking bool

	// OverrideLinkTracking enables click tracking for this message.
	OverrideLinkTracking bool

	// UnsubscribeURL is the URL to redirect unsubscribe requests to.
	UnsubscribeURL string
}

// Wire types — not exported; only used for JSON encoding.

type sendMessageWire struct {
	Data sendMessageData `json:"data"`
}

type sendMessageData struct {
	Message              Message `json:"message"`
	OverrideOpenTracking bool    `json:"override_open_tracking,omitempty"`
	OverrideLinkTracking bool    `json:"override_link_tracking,omitempty"`
	UnsubscribeURL       string  `json:"unsubscribe_url,omitempty"`
}

// SendMessageResponse is the response from [Client.SendMessage].
type SendMessageResponse struct {
	// SourceTrackingID identifies this message for delivery tracking.
	// Pass it to [Client.GetEmailDisposition] to poll for status.
	SourceTrackingID string `json:"sourceTrackingId"`

	// CustomHeaders is the map of custom headers that were accepted.
	CustomHeaders map[string]string `json:"customHeaders"`

	// Data contains the service status message (typically "Service OK").
	Data string `json:"data"`
}

// ---------------------------------------------------------------------------
// Send batch messages
// ---------------------------------------------------------------------------

// SendBatchRequest is the request for [Client.SendBatch].
// Paubox recommends batches of 50 or fewer messages.
type SendBatchRequest struct {
	// Messages is the list of emails to send. Maximum recommended: 50.
	Messages []Message
}

type sendBatchWire struct {
	Data sendBatchData `json:"data"`
}

type sendBatchData struct {
	Messages []Message `json:"messages"`
}

// SendBatchResponse is the response from [Client.SendBatch].
// Entries are returned in the same order as the request messages.
type SendBatchResponse struct {
	// Messages contains one entry per message in the request.
	Messages []SendMessageResponse `json:"messages"`
}

// ---------------------------------------------------------------------------
// Get email disposition
// ---------------------------------------------------------------------------

// EmailDisposition is the response from [Client.GetEmailDisposition].
type EmailDisposition struct {
	// SourceTrackingID echoes the tracking ID that was queried.
	SourceTrackingID string `json:"sourceTrackingId"`

	// Data contains the message delivery record.
	Data EmailDispositionData `json:"data"`
}

// EmailDispositionData wraps the message-level delivery record.
type EmailDispositionData struct {
	Message MessageDisposition `json:"message"`
}

// MessageDisposition contains per-recipient delivery details and aggregate
// engagement metrics for a sent message.
type MessageDisposition struct {
	// ID is the internal Paubox message identifier.
	ID string `json:"id"`

	// MessageDeliveries contains one entry per recipient.
	MessageDeliveries []MessageDelivery `json:"message_deliveries"`

	// TotalOpens is the total number of open events across all recipients.
	TotalOpens *int `json:"total_opens,omitempty"`

	// DistinctOpens is the count of unique recipients who opened the message.
	DistinctOpens *int `json:"distinct_opens,omitempty"`

	// TotalClickCount is the aggregate number of link clicks.
	TotalClickCount *int `json:"total_click_count,omitempty"`

	// ClicksPerLink lists click counts broken down by tracked link.
	ClicksPerLink []LinkClick `json:"clicks_per_link,omitempty"`

	// Unsubscribed indicates whether any recipient unsubscribed.
	Unsubscribed *bool `json:"unsubscribed,omitempty"`
}

// MessageDelivery is the delivery record for a single recipient.
type MessageDelivery struct {
	// Recipient is the email address of the recipient.
	Recipient string `json:"recipient"`

	// Status contains delivery and open tracking details.
	Status DeliveryStatus `json:"status"`
}

// DeliveryStatus contains the delivery and open state for one recipient.
type DeliveryStatus struct {
	// DeliveryStatus is the current delivery state.
	// See the [DeliveryStatus…] constants for possible values.
	DeliveryStatus string `json:"deliveryStatus"`

	// DeliveryTime is the RFC 2822 timestamp of delivery, if delivered.
	DeliveryTime *string `json:"deliveryTime,omitempty"`

	// OpenedStatus is either [OpenedStatusOpened] or [OpenedStatusNotOpened].
	OpenedStatus *string `json:"openedStatus,omitempty"`

	// OpenedTime is the timestamp of the first open event, if opened.
	OpenedTime *string `json:"openedTime,omitempty"`
}

// LinkClick contains the interaction count for a specific tracked link.
type LinkClick struct {
	// ClickCount is the total number of clicks on this link.
	ClickCount int `json:"click_count"`

	// TargetURL is the destination URL of the link.
	TargetURL string `json:"target_url"`
}

// Delivery status values returned by the Paubox API.
const (
	DeliveryStatusProcessing             = "processing"
	DeliveryStatusTLSNotOffered          = "TLS not offered, sending via Secure Portal"
	DeliveryStatusSoftBounced            = "soft bounced"
	DeliveryStatusSoftBouncedMailboxFull = "soft bounced - mailbox full"
	DeliveryStatusHardBounced            = "hard bounced"
	DeliveryStatusInternalError          = "Internal error. Please check back later."
	DeliveryStatusDelivered              = "delivered"
	DeliveryStatusDeliveredViaPortal     = "delivered via secure portal"
)

// Open status values returned by the Paubox API.
const (
	OpenedStatusOpened    = "opened"
	OpenedStatusNotOpened = "not opened"
)

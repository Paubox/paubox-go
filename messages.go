package paubox

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// SendMessage sends a single HIPAA-compliant email message.
//
// At least one of Message.Content.PlainText or Message.Content.HTML must be
// set. Message.Headers.From and Message.Headers.Subject are required.
//
// API: POST /messages
func (c *Client) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	if err := validateSendMessageRequest(req); err != nil {
		return nil, err
	}

	wire := sendMessageWire{
		Data: sendMessageData{
			Message:              req.Message,
			OverrideOpenTracking: req.OverrideOpenTracking,
			OverrideLinkTracking: req.OverrideLinkTracking,
			UnsubscribeURL:       req.UnsubscribeURL,
		},
	}

	var resp SendMessageResponse
	if err := c.doJSON(ctx, http.MethodPost, "/messages", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendBatch sends multiple email messages in a single API call.
//
// Paubox recommends batches of 50 or fewer messages. Each message is
// validated individually before the request is sent. Responses are returned
// in the same order as the request messages.
//
// API: POST /bulk_messages
func (c *Client) SendBatch(ctx context.Context, req *SendBatchRequest) (*SendBatchResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("paubox: SendBatch: request must not be nil")
	}
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("paubox: SendBatch: at least one message is required")
	}
	for i := range req.Messages {
		if err := validateMessage(&req.Messages[i], i); err != nil {
			return nil, err
		}
	}

	wire := sendBatchWire{Data: sendBatchData{Messages: req.Messages}}

	var resp SendBatchResponse
	if err := c.doJSON(ctx, http.MethodPost, "/bulk_messages", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetEmailDisposition retrieves the delivery status and engagement metrics
// for a previously sent message.
//
// sourceTrackingID is the SourceTrackingID value returned by [Client.SendMessage]
// or one of the entries in a [Client.SendBatch] response.
//
// API: GET /message_receipt
func (c *Client) GetEmailDisposition(ctx context.Context, sourceTrackingID string) (*EmailDisposition, error) {
	if strings.TrimSpace(sourceTrackingID) == "" {
		return nil, fmt.Errorf("paubox: GetEmailDisposition: sourceTrackingID must not be empty")
	}

	q := url.Values{}
	q.Set("sourceTrackingId", sourceTrackingID)

	var disp EmailDisposition
	if err := c.doJSON(ctx, http.MethodGet, "/message_receipt?"+q.Encode(), nil, &disp); err != nil {
		return nil, err
	}
	return &disp, nil
}

// ---------------------------------------------------------------------------
// Validation helpers
// ---------------------------------------------------------------------------

func validateSendMessageRequest(req *SendMessageRequest) error {
	if req == nil {
		return fmt.Errorf("paubox: SendMessage: request must not be nil")
	}
	return validateMessage(&req.Message, -1)
}

func validateMessage(msg *Message, idx int) error {
	prefix := "paubox: message"
	if idx >= 0 {
		prefix = fmt.Sprintf("paubox: message[%d]", idx)
	}

	if len(msg.Recipients) == 0 {
		return fmt.Errorf("%s: recipients must not be empty", prefix)
	}
	if msg.Headers.From == "" {
		return fmt.Errorf("%s: headers.from must not be empty", prefix)
	}
	if msg.Headers.Subject == "" {
		return fmt.Errorf("%s: headers.subject must not be empty", prefix)
	}
	if msg.Content.PlainText == nil && msg.Content.HTML == nil {
		return fmt.Errorf("%s: content: at least one of text/plain or text/html is required", prefix)
	}
	for _, a := range msg.Attachments {
		if a.FileName == "" {
			return fmt.Errorf("%s: attachment: fileName must not be empty", prefix)
		}
		if a.ContentType == "" {
			return fmt.Errorf("%s: attachment %q: contentType must not be empty", prefix, a.FileName)
		}
		if a.Content == "" {
			return fmt.Errorf("%s: attachment %q: content must not be empty", prefix, a.FileName)
		}
	}
	return nil
}

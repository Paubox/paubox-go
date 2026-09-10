package paubox

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ScheduleMessage schedules an email for future delivery.
//
// scheduledAt must be in the future and within 30 days. The message is
// validated the same way as [Client.SendMessage].
//
// API: POST /schedule
func (c *Client) ScheduleMessage(ctx context.Context, req *ScheduleMessageRequest) (*ScheduleMessageResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("paubox: ScheduleMessage: request must not be nil")
	}
	if err := validateMessage(&req.Message, -1); err != nil {
		return nil, err
	}
	if req.ScheduledAt.IsZero() {
		return nil, fmt.Errorf("paubox: ScheduleMessage: scheduledAt must not be zero")
	}

	wire := scheduleMessageWire{
		Data: scheduleMessageData{
			Message:     req.Message,
			ScheduledAt: req.ScheduledAt.UTC().Format(time.RFC3339),
		},
	}

	var resp ScheduleMessageResponse
	if err := c.doJSON(ctx, http.MethodPost, "/schedule", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetScheduledMessage retrieves the status of a scheduled message.
//
// API: GET /schedule/{sourceTrackingId}
func (c *Client) GetScheduledMessage(ctx context.Context, sourceTrackingID string) (*ScheduledMessageStatus, error) {
	if strings.TrimSpace(sourceTrackingID) == "" {
		return nil, fmt.Errorf("paubox: GetScheduledMessage: sourceTrackingID must not be empty")
	}

	var status ScheduledMessageStatus
	if err := c.doJSON(ctx, http.MethodGet, "/schedule/"+sourceTrackingID, nil, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// RescheduleMessage changes the scheduled send time for a pending message.
//
// API: PATCH /schedule/{sourceTrackingId}
func (c *Client) RescheduleMessage(ctx context.Context, sourceTrackingID string, scheduledAt time.Time) (*RescheduleResponse, error) {
	if strings.TrimSpace(sourceTrackingID) == "" {
		return nil, fmt.Errorf("paubox: RescheduleMessage: sourceTrackingID must not be empty")
	}
	if scheduledAt.IsZero() {
		return nil, fmt.Errorf("paubox: RescheduleMessage: scheduledAt must not be zero")
	}

	wire := rescheduleWire{
		ScheduledAt: scheduledAt.UTC().Format(time.RFC3339),
	}

	var resp RescheduleResponse
	if err := c.doJSON(ctx, http.MethodPatch, "/schedule/"+sourceTrackingID, wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CancelScheduledMessage cancels a pending scheduled message.
//
// API: POST /schedule/{sourceTrackingId}/cancel
func (c *Client) CancelScheduledMessage(ctx context.Context, sourceTrackingID string) (*CancelScheduledResponse, error) {
	if strings.TrimSpace(sourceTrackingID) == "" {
		return nil, fmt.Errorf("paubox: CancelScheduledMessage: sourceTrackingID must not be empty")
	}

	var resp CancelScheduledResponse
	if err := c.doJSON(ctx, http.MethodPost, "/schedule/"+sourceTrackingID+"/cancel", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

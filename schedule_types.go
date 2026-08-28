package paubox

import "time"

// ---------------------------------------------------------------------------
// Schedule message
// ---------------------------------------------------------------------------

// ScheduleMessageRequest is the request for [Client.ScheduleMessage].
type ScheduleMessageRequest struct {
	// Message is the email to schedule.
	Message Message

	// ScheduledAt is the UTC time to send the message.
	// Must be in the future and within 30 days.
	ScheduledAt time.Time
}

type scheduleMessageWire struct {
	Data scheduleMessageData `json:"data"`
}

type scheduleMessageData struct {
	Message     Message `json:"message"`
	ScheduledAt string  `json:"scheduled_at"`
}

// ScheduleMessageResponse is the response from [Client.ScheduleMessage].
type ScheduleMessageResponse struct {
	SourceTrackingID string `json:"sourceTrackingId"`
	ScheduledAt      string `json:"scheduledAt"`
	State            string `json:"state"`
	Data             string `json:"data"`
}

// ---------------------------------------------------------------------------
// Get scheduled message status
// ---------------------------------------------------------------------------

// ScheduledMessageStatus is the response from [Client.GetScheduledMessage].
type ScheduledMessageStatus struct {
	SourceTrackingID string  `json:"sourceTrackingId"`
	ScheduledAt      string  `json:"scheduledAt"`
	State            string  `json:"state"`
	MessageID        int     `json:"messageId"`
	SentAt           *string `json:"sentAt"`
	CancelledAt      *string `json:"cancelledAt"`
	ErrorMessage     *string `json:"errorMessage"`
}

// ---------------------------------------------------------------------------
// Reschedule
// ---------------------------------------------------------------------------

type rescheduleWire struct {
	ScheduledAt string `json:"scheduled_at"`
}

// RescheduleResponse is the response from [Client.RescheduleMessage].
type RescheduleResponse struct {
	SourceTrackingID string `json:"sourceTrackingId"`
	ScheduledAt      string `json:"scheduledAt"`
	Data             string `json:"data"`
}

// ---------------------------------------------------------------------------
// Cancel
// ---------------------------------------------------------------------------

// CancelScheduledResponse is the response from [Client.CancelScheduledMessage].
type CancelScheduledResponse struct {
	SourceTrackingID string `json:"sourceTrackingId"`
	State            string `json:"state"`
	Data             string `json:"data"`
}

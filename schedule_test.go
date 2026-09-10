package paubox

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestScheduleMessage_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"sourceTrackingId":"tid-sched","scheduledAt":"2025-12-25T15:00:00Z","state":"pending","data":"Service OK"}`)
	}))
	defer srv.Close()

	resp, err := newTestClient(t, srv).ScheduleMessage(context.Background(), &ScheduleMessageRequest{
		Message:     validMessage(),
		ScheduledAt: time.Date(2025, 12, 25, 15, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("ScheduleMessage() error: %v", err)
	}
	if resp.SourceTrackingID != "tid-sched" {
		t.Errorf("SourceTrackingID = %q, want tid-sched", resp.SourceTrackingID)
	}
	if resp.State != "pending" {
		t.Errorf("State = %q, want pending", resp.State)
	}
}

func TestScheduleMessage_SendsCorrectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		respondJSON(w, http.StatusOK, `{"sourceTrackingId":"x","scheduledAt":"2025-12-25T15:00:00Z","state":"pending","data":"OK"}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).ScheduleMessage(context.Background(), &ScheduleMessageRequest{
		Message:     validMessage(),
		ScheduledAt: time.Date(2025, 12, 25, 15, 0, 0, 0, time.UTC),
	})
	if gotPath != "/schedule" {
		t.Errorf("path = %q, want /schedule", gotPath)
	}
}

func TestScheduleMessage_SendsScheduledAt(t *testing.T) {
	var body map[string]json.RawMessage
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		respondJSON(w, http.StatusOK, `{"sourceTrackingId":"x","scheduledAt":"2025-12-25T15:00:00Z","state":"pending","data":"OK"}`)
	}))
	defer srv.Close()

	_, _ = newTestClient(t, srv).ScheduleMessage(context.Background(), &ScheduleMessageRequest{
		Message:     validMessage(),
		ScheduledAt: time.Date(2025, 12, 25, 15, 0, 0, 0, time.UTC),
	})

	var data struct {
		ScheduledAt string `json:"scheduled_at"`
	}
	_ = json.Unmarshal(body["data"], &data)
	if data.ScheduledAt != "2025-12-25T15:00:00Z" {
		t.Errorf("scheduled_at = %q, want 2025-12-25T15:00:00Z", data.ScheduledAt)
	}
}

func TestScheduleMessage_NilRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).ScheduleMessage(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
}

func TestScheduleMessage_ZeroTime(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).ScheduleMessage(context.Background(), &ScheduleMessageRequest{
		Message: validMessage(),
	})
	if err == nil {
		t.Fatal("expected error for zero scheduledAt")
	}
}

func TestGetScheduledMessage_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{"sourceTrackingId":"tid-sched","scheduledAt":"2025-12-25T15:00:00Z","state":"pending","messageId":123}`)
	}))
	defer srv.Close()

	status, err := newTestClient(t, srv).GetScheduledMessage(context.Background(), "tid-sched")
	if err != nil {
		t.Fatalf("GetScheduledMessage() error: %v", err)
	}
	if status.State != "pending" {
		t.Errorf("State = %q, want pending", status.State)
	}
}

func TestGetScheduledMessage_EmptyID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).GetScheduledMessage(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty sourceTrackingID")
	}
}

func TestRescheduleMessage_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		respondJSON(w, http.StatusOK, `{"sourceTrackingId":"tid","scheduledAt":"2025-12-26T10:00:00Z","data":"Rescheduled"}`)
	}))
	defer srv.Close()

	resp, err := newTestClient(t, srv).RescheduleMessage(context.Background(), "tid", time.Date(2025, 12, 26, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("RescheduleMessage() error: %v", err)
	}
	if resp.Data != "Rescheduled" {
		t.Errorf("Data = %q, want Rescheduled", resp.Data)
	}
}

func TestRescheduleMessage_EmptyID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).RescheduleMessage(context.Background(), "", time.Now().Add(time.Hour))
	if err == nil {
		t.Fatal("expected error for empty sourceTrackingID")
	}
}

func TestCancelScheduledMessage_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		respondJSON(w, http.StatusOK, `{"sourceTrackingId":"tid","state":"cancelled","data":"Cancelled"}`)
	}))
	defer srv.Close()

	resp, err := newTestClient(t, srv).CancelScheduledMessage(context.Background(), "tid")
	if err != nil {
		t.Fatalf("CancelScheduledMessage() error: %v", err)
	}
	if resp.State != "cancelled" {
		t.Errorf("State = %q, want cancelled", resp.State)
	}
}

func TestCancelScheduledMessage_EmptyID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, `{}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).CancelScheduledMessage(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty sourceTrackingID")
	}
}

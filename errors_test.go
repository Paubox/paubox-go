package paubox

import (
	"bytes"
	"errors"
	"net/http"
	"testing"
)

func TestPauboxError_Error_WithTitle(t *testing.T) {
	e := &PauboxError{StatusCode: 400, Title: "Bad Request"}
	want := "paubox: HTTP 400: Bad Request"
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestPauboxError_Error_NoTitle(t *testing.T) {
	e := &PauboxError{StatusCode: 503}
	want := "paubox: HTTP 503"
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestPauboxError_Is_MatchesBySatusCode(t *testing.T) {
	e := &PauboxError{StatusCode: 401, Title: "Unauthorized", Details: "bad key"}

	if !errors.Is(e, ErrUnauthorized) {
		t.Error("errors.Is(e, ErrUnauthorized) = false, want true")
	}
	if errors.Is(e, ErrNotFound) {
		t.Error("errors.Is(e, ErrNotFound) = true, want false")
	}
	if errors.Is(e, ErrRateLimit) {
		t.Error("errors.Is(e, ErrRateLimit) = true, want false")
	}
}

func TestPauboxError_As(t *testing.T) {
	e := &PauboxError{StatusCode: 403, Title: "Forbidden", Details: "no access", RequestID: "req-xyz"}

	var apiErr *PauboxError
	if !errors.As(e, &apiErr) {
		t.Fatal("errors.As returned false")
	}
	if apiErr.StatusCode != 403 {
		t.Errorf("StatusCode = %d, want 403", apiErr.StatusCode)
	}
	if apiErr.Details != "no access" {
		t.Errorf("Details = %q, want 'no access'", apiErr.Details)
	}
	if apiErr.RequestID != "req-xyz" {
		t.Errorf("RequestID = %q, want req-xyz", apiErr.RequestID)
	}
}

func TestParseAPIError_EmailWireFormat(t *testing.T) {
	raw := []byte(`{"errors":[{"code":1001,"title":"Unauthorized","details":"Invalid API key"}]}`)
	e := parseAPIError(http.StatusUnauthorized, "rid-123", raw)

	if e.StatusCode != 401 {
		t.Errorf("StatusCode = %d, want 401", e.StatusCode)
	}
	if e.Code != "1001" {
		t.Errorf("Code = %q, want 1001", e.Code)
	}
	if e.Title != "Unauthorized" {
		t.Errorf("Title = %q, want Unauthorized", e.Title)
	}
	if e.Details != "Invalid API key" {
		t.Errorf("Details = %q, want 'Invalid API key'", e.Details)
	}
	if e.RequestID != "rid-123" {
		t.Errorf("RequestID = %q, want rid-123", e.RequestID)
	}
}

func TestParseAPIError_EmptyErrorsArray(t *testing.T) {
	raw := []byte(`{"errors":[]}`)
	e := parseAPIError(http.StatusBadRequest, "", raw)
	// Falls back to HTTP status text.
	if e.Title != "Bad Request" {
		t.Errorf("Title = %q, want 'Bad Request'", e.Title)
	}
}

func TestParseAPIError_UnparsableBody(t *testing.T) {
	raw := []byte(`not json`)
	e := parseAPIError(http.StatusInternalServerError, "", raw)
	if e.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", e.StatusCode)
	}
	if e.Title != "Internal Server Error" {
		t.Errorf("Title = %q, want 'Internal Server Error'", e.Title)
	}
}

func TestParseAPIError_RawPreserved(t *testing.T) {
	raw := []byte(`{"errors":[{"code":400,"title":"Bad","details":"d"}]}`)
	e := parseAPIError(400, "", raw)
	if !bytes.Equal(e.Raw, raw) {
		t.Error("Raw body was not preserved")
	}
}

func TestSentinelErrors_StatusCodes(t *testing.T) {
	tests := []struct {
		sentinel *PauboxError
		code     int
	}{
		{ErrBadRequest, 400},
		{ErrUnauthorized, 401},
		{ErrForbidden, 403},
		{ErrNotFound, 404},
		{ErrRateLimit, 429},
		{ErrServerError, 500},
	}
	for _, tc := range tests {
		if tc.sentinel.StatusCode != tc.code {
			t.Errorf("sentinel StatusCode = %d, want %d", tc.sentinel.StatusCode, tc.code)
		}
	}
}

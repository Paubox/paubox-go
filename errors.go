package paubox

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// PauboxError represents an error returned by the Paubox API.
//
// Use [errors.As] to extract the full error detail and [errors.Is] with the
// package-level sentinel values to match specific HTTP status codes:
//
//	var apiErr *paubox.PauboxError
//	if errors.As(err, &apiErr) {
//	    fmt.Printf("HTTP %d: %s — %s\n", apiErr.StatusCode, apiErr.Title, apiErr.Details)
//	}
//
//	if errors.Is(err, paubox.ErrUnauthorized) {
//	    // handle authentication failure
//	}
type PauboxError struct { //nolint:revive // intentional: "PauboxError" is unambiguous in user code; avoids collision with the standard "error" interface
	// StatusCode is the HTTP status code returned by the API.
	StatusCode int

	// Code is the application-level error code. Preserved as a string because
	// the Paubox API uses integer codes in some responses.
	Code string

	// Title is the short error summary from the API response.
	Title string

	// Details contains extended error information from the API response.
	Details string

	// RequestID is the value of the X-Request-Id response header. Include
	// this value when contacting Paubox support.
	RequestID string

	// Raw is the unmodified response body for debugging. The SDK never logs
	// this value to avoid inadvertently capturing PHI.
	Raw []byte
}

// Error implements the error interface.
func (e *PauboxError) Error() string {
	if e.Title != "" {
		return fmt.Sprintf("paubox: HTTP %d: %s", e.StatusCode, e.Title)
	}
	return fmt.Sprintf("paubox: HTTP %d", e.StatusCode)
}

// Is reports whether this error matches target by HTTP status code.
// This makes [errors.Is] work with the sentinel values below.
func (e *PauboxError) Is(target error) bool {
	var t *PauboxError
	if !errors.As(target, &t) {
		return false
	}
	// A sentinel has only StatusCode set; match on that alone.
	return t.StatusCode != 0 && t.StatusCode == e.StatusCode
}

// Sentinel errors. Use with [errors.Is]:
//
//	if errors.Is(err, paubox.ErrNotFound) { ... }
var (
	ErrBadRequest   = &PauboxError{StatusCode: http.StatusBadRequest}
	ErrUnauthorized = &PauboxError{StatusCode: http.StatusUnauthorized}
	ErrForbidden    = &PauboxError{StatusCode: http.StatusForbidden}
	ErrNotFound     = &PauboxError{StatusCode: http.StatusNotFound}
	ErrRateLimit    = &PauboxError{StatusCode: http.StatusTooManyRequests}
	ErrServerError  = &PauboxError{StatusCode: http.StatusInternalServerError}
)

// wireErrors is the error envelope used by the Paubox Email API:
//
//	{"errors": [{"code": 1001, "title": "...", "details": "..."}]}
type wireErrors struct {
	Errors []wireError `json:"errors"`
}

type wireError struct {
	Code    int    `json:"code"`
	Title   string `json:"title"`
	Details string `json:"details"`
}

// parseAPIError converts a non-2xx HTTP response into a *[PauboxError].
// The raw body is attached for debugging but is never logged by the SDK.
func parseAPIError(statusCode int, requestID string, raw []byte) *PauboxError {
	e := &PauboxError{
		StatusCode: statusCode,
		RequestID:  requestID,
		Raw:        raw,
	}

	var wire wireErrors
	if jsonErr := json.Unmarshal(raw, &wire); jsonErr == nil && len(wire.Errors) > 0 {
		first := wire.Errors[0]
		e.Code = fmt.Sprintf("%d", first.Code)
		e.Title = first.Title
		e.Details = first.Details
		return e
	}

	// Fallback to HTTP status text when the body is unparseable.
	e.Title = http.StatusText(statusCode)
	return e
}

// formsWireError is the error envelope used by the Paubox Forms service:
//
//	{"message": "..."}
type formsWireError struct {
	Message string `json:"message"`
}

// formsWireErrorAlt is the alternate error envelope used by a few Forms
// service paths (e.g. the CSV export when a form has no JSON definition):
//
//	{"error": "..."}
type formsWireErrorAlt struct {
	Error string `json:"error"`
}

// parseFormsAPIError converts a non-2xx Forms service response into a
// *[PauboxError]. The Forms error envelope is {"message": "..."}, but some
// error paths return the Email-style envelope, {"error": "..."} (CSV export),
// plain text (CSV/PDF exports), or an empty body — each falls back gracefully. The raw body is attached
// for debugging but is never logged by the SDK (it may contain PHI).
func parseFormsAPIError(statusCode int, requestID string, raw []byte) *PauboxError {
	e := &PauboxError{
		StatusCode: statusCode,
		RequestID:  requestID,
		Raw:        raw,
	}

	var fw formsWireError
	if jsonErr := json.Unmarshal(raw, &fw); jsonErr == nil && fw.Message != "" {
		e.Title = fw.Message
		return e
	}

	// Fall back to the Email API envelope in case the gateway proxies one.
	var wire wireErrors
	if jsonErr := json.Unmarshal(raw, &wire); jsonErr == nil && len(wire.Errors) > 0 {
		first := wire.Errors[0]
		e.Code = fmt.Sprintf("%d", first.Code)
		e.Title = first.Title
		e.Details = first.Details
		return e
	}

	// A few Forms paths use {"error": "..."} instead of {"message": "..."}.
	var alt formsWireErrorAlt
	if jsonErr := json.Unmarshal(raw, &alt); jsonErr == nil && alt.Error != "" {
		e.Title = alt.Error
		return e
	}

	// Final fallback: HTTP status text (covers plain-text and empty bodies).
	e.Title = http.StatusText(statusCode)
	return e
}

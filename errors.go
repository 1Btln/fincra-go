package fincra

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/1Btln/fincra-go/internal/api"
)

var (
	// ErrMissingRequiredParameter is wrapped by client-side validation errors
	// caused by a missing required argument.
	ErrMissingRequiredParameter = api.ErrMissingRequiredParameter

	// ErrInvalidParameter is wrapped by client-side validation errors caused by
	// a malformed argument.
	ErrInvalidParameter = api.ErrInvalidParameter
)

// APIError is returned when Fincra responds with a non-2xx status code. It is
// compatible with errors.As:
//
//	var apiErr *fincra.APIError
//	if errors.As(err, &apiErr) && apiErr.IsNotFound() {
//		// handle a missing resource
//	}
type APIError struct {
	// StatusCode is the HTTP status code.
	StatusCode int

	// Method and Path identify the failed request. Path includes the query
	// string but never the request body.
	Method string
	Path   string

	// Message is the human-readable message reported by Fincra. Fincra uses
	// both "error" and "message" fields depending on the endpoint.
	Message string

	// ErrorType is Fincra's machine-readable error category, such as
	// RESOURCE_NOT_FOUND, when supplied.
	ErrorType string

	// Details is the raw "errors" field from the response. Fincra encodes it
	// as a string on some endpoints and as an array or object on others.
	Details json.RawMessage

	// RequestID is the value of the x-request-id response header, when present.
	// Include it when reporting issues to Fincra support.
	RequestID string

	// Body is the raw response body, capped by the client response limit.
	Body []byte

	// Retryable reports whether the client considered this status retryable.
	Retryable bool

	// Err is the underlying transport, read, or decode error, when applicable.
	Err error
}

// Error implements error.
func (e *APIError) Error() string {
	var b strings.Builder
	b.WriteString("fincra: ")
	b.WriteString(e.Method)
	b.WriteByte(' ')
	b.WriteString(e.Path)
	b.WriteString(": ")
	b.WriteString(strconv.Itoa(e.StatusCode))
	if e.Message != "" {
		b.WriteString(": ")
		b.WriteString(e.Message)
	}
	if e.ErrorType != "" {
		b.WriteString(" (")
		b.WriteString(e.ErrorType)
		b.WriteByte(')')
	}
	if e.Err != nil {
		b.WriteString(": ")
		b.WriteString(e.Err.Error())
	}
	return b.String()
}

// Unwrap returns the underlying error, if any.
func (e *APIError) Unwrap() error { return e.Err }

// IsNotFound reports whether the resource was not found (HTTP 404).
func (e *APIError) IsNotFound() bool { return e.StatusCode == 404 }

// IsUnauthorized reports whether authentication failed (HTTP 401).
func (e *APIError) IsUnauthorized() bool { return e.StatusCode == 401 }

// IsForbidden reports whether the request was denied (HTTP 403). Fincra also
// returns 403 when the caller's IP address is not whitelisted.
func (e *APIError) IsForbidden() bool { return e.StatusCode == 403 }

// IsRateLimited reports whether the request was throttled (HTTP 429).
func (e *APIError) IsRateLimited() bool { return e.StatusCode == 429 }

// IsServerError reports whether Fincra returned a 5xx status.
func (e *APIError) IsServerError() bool { return e.StatusCode >= 500 && e.StatusCode <= 599 }

// IsAPIError extracts an *APIError from err.
func IsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

// Package api holds helpers shared by the Fincra resource packages.
//
// It is internal so the SDK's public surface stays focused on resource services
// and the root client. Feature packages use these helpers to validate arguments
// and to avoid importing the root package, which would create an import cycle.
package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Header names used by Fincra.
const (
	HeaderAPIKey     = "api-key"
	HeaderPublicKey  = "x-pub-key"
	HeaderBusinessID = "x-business-id"
)

var (
	// ErrMissingRequiredParameter is wrapped by client-side validation errors
	// caused by a missing required argument.
	ErrMissingRequiredParameter = errors.New("fincra: missing required parameter")

	// ErrInvalidParameter is wrapped by client-side validation errors caused by
	// a malformed argument.
	ErrInvalidParameter = errors.New("fincra: invalid parameter")
)

// Missing returns an error wrapping ErrMissingRequiredParameter for field.
func Missing(field string) error {
	return fmt.Errorf("%w: %s", ErrMissingRequiredParameter, field)
}

// Invalid returns an error wrapping ErrInvalidParameter for field.
func Invalid(field, reason string) error {
	return fmt.Errorf("%w: %s %s", ErrInvalidParameter, field, reason)
}

// Require returns a missing-parameter error when value is blank.
func Require(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return Missing(field)
	}
	return nil
}

// RequireAny returns a missing-parameter error when every value is blank. The
// field argument describes the group of alternatives.
func RequireAny(field string, values ...string) error {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return nil
		}
	}
	return Missing(field)
}

// BusinessID resolves an optional override against the client's configured
// business identifier.
func BusinessID(override, fallback string) (string, error) {
	if id := strings.TrimSpace(override); id != "" {
		return id, nil
	}
	if id := strings.TrimSpace(fallback); id != "" {
		return id, nil
	}
	return "", Missing("businessID")
}

// ValidateDate checks that value, when set, is YYYY-MM-DD or RFC3339.
func ValidateDate(field, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02", time.RFC3339} {
		if _, err := time.Parse(layout, value); err == nil {
			return nil
		}
	}
	return Invalid(field, "must use YYYY-MM-DD or RFC3339 format")
}

// ValidateCurrency checks that value is one of the allowed ISO codes.
func ValidateCurrency(field, value string, allowed ...string) error {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return Missing(field)
	}
	for _, candidate := range allowed {
		if value == candidate {
			return nil
		}
	}
	return Invalid(field, "is not a supported currency")
}

// BusinessHeader returns a request option that overrides x-business-id.
func BusinessHeader(businessID string) func(*http.Request) {
	return func(r *http.Request) {
		if id := strings.TrimSpace(businessID); id != "" {
			r.Header.Set(HeaderBusinessID, id)
		}
	}
}

// SetString sets key in query when value is not blank.
func SetString(query url.Values, key, value string) {
	if value = strings.TrimSpace(value); value != "" {
		query.Set(key, value)
	}
}

// SetInt sets key in query when value is greater than zero.
func SetInt(query url.Values, key string, value int) {
	if value > 0 {
		query.Set(key, strconv.Itoa(value))
	}
}

// SetBool sets key in query when value is true.
func SetBool(query url.Values, key string, value bool) {
	if value {
		query.Set(key, "true")
	}
}

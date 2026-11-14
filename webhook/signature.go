// Package webhook provides Fincra webhook event types and signature helpers.
package webhook

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"strings"
)

// SignatureHeader is the header Fincra uses for webhook signatures.
const SignatureHeader = "signature"

// SignPayload returns the lowercase hex HMAC-SHA512 of payload using secret.
func SignPayload(secret string, payload []byte) string {
	mac := hmac.New(sha512.New, []byte(secret))
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature reports whether signature is the HMAC-SHA512 of the raw
// payload under secret. Pass the request body exactly as received; re-marshaling
// a decoded value changes key order and whitespace and invalidates the signature.
func VerifySignature(secret string, payload []byte, signature string) bool {
	if secret == "" {
		return false
	}
	expected := SignPayload(secret, payload)
	actual := strings.ToLower(strings.TrimSpace(signature))
	return hmac.Equal([]byte(expected), []byte(actual))
}

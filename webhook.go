package fincra

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"strings"
)

// WebhookSignatureHeader is the header Fincra uses for the HMAC-SHA512
// signature of a webhook payload. Header lookups are case-insensitive in net/http.
const WebhookSignatureHeader = "signature"

// SignWebhookPayload returns the lowercase hex HMAC-SHA512 of payload using
// secret. It is exposed so tests and local mocks can produce signatures that
// match Fincra's.
func SignWebhookPayload(secret string, payload []byte) string {
	mac := hmac.New(sha512.New, []byte(secret))
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyWebhookSignature reports whether signature is the HMAC-SHA512 of the
// raw payload under secret.
//
// Pass the raw request body exactly as received. Re-marshalling a decoded struct
// changes key order and whitespace, which invalidates the signature.
func VerifyWebhookSignature(secret string, payload []byte, signature string) bool {
	if secret == "" {
		return false
	}
	expected := SignWebhookPayload(secret, payload)
	actual := strings.ToLower(strings.TrimSpace(signature))
	return hmac.Equal([]byte(expected), []byte(actual))
}

// VerifyWebhookSignature verifies a signature using the client's configured
// WebhookSecret. It returns false when no webhook secret is configured.
func (c *Client) VerifyWebhookSignature(payload []byte, signature string) bool {
	return VerifyWebhookSignature(c.cfg.WebhookSecret, payload, signature)
}

package fincra_test

import (
	"testing"

	fincra "github.com/1Btln/fincra-go"
)

const (
	testWebhookSecret    = "whsec_test_51H8xQ2"
	testWebhookPayload   = `{"event":"payout.successful","data":{"reference":"33aef44be6f7441a"}}`
	testWebhookSignature = "f81161c8992a7f51634658b6bef22ea85c747b408175202df398d3faa3cf932d26a23d8c38f768e0dee25eaef1165461b2692e023c79e205a176b0a83e801b04"
)

func TestSignWebhookPayload(t *testing.T) {
	got := fincra.SignWebhookPayload(testWebhookSecret, []byte(testWebhookPayload))
	if got != testWebhookSignature {
		t.Errorf("SignWebhookPayload() = %q, want %q", got, testWebhookSignature)
	}
}

func TestVerifyWebhookSignature(t *testing.T) {
	tests := []struct {
		name      string
		secret    string
		signature string
		want      bool
	}{
		{name: "valid", secret: testWebhookSecret, signature: testWebhookSignature, want: true},
		{name: "uppercase signature", secret: testWebhookSecret, signature: "F81161C8992A7F51634658B6BEF22EA85C747B408175202DF398D3FAA3CF932D26A23D8C38F768E0DEE25EAEF1165461B2692E023C79E205A176B0A83E801B04", want: true},
		{name: "surrounding whitespace", secret: testWebhookSecret, signature: "  " + testWebhookSignature + "\n", want: true},
		{name: "wrong signature", secret: testWebhookSecret, signature: "deadbeef", want: false},
		{name: "wrong secret", secret: "whsec_other", signature: testWebhookSignature, want: false},
		{name: "empty secret", secret: "", signature: testWebhookSignature, want: false},
		{name: "empty signature", secret: testWebhookSecret, signature: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fincra.VerifyWebhookSignature(tt.secret, []byte(testWebhookPayload), tt.signature)
			if got != tt.want {
				t.Errorf("VerifyWebhookSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyWebhookSignatureRejectsModifiedPayload(t *testing.T) {
	modified := []byte(`{"event":"payout.successful","data":{"reference":"tampered"}}`)
	if fincra.VerifyWebhookSignature(testWebhookSecret, modified, testWebhookSignature) {
		t.Error("VerifyWebhookSignature() = true for modified payload, want false")
	}
}

func TestClientVerifyWebhookSignature(t *testing.T) {
	client := newTestClient(t, "https://example.test", fincra.WithWebhookSecret(testWebhookSecret))

	if !client.VerifyWebhookSignature([]byte(testWebhookPayload), testWebhookSignature) {
		t.Error("client verification = false, want true")
	}

	noSecret := newTestClient(t, "https://example.test")
	if noSecret.VerifyWebhookSignature([]byte(testWebhookPayload), testWebhookSignature) {
		t.Error("client without webhook secret verified signature, want false")
	}
}

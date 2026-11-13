package fincra_test

import (
	"fmt"
	"log"

	"github.com/Ledgenta/fincra-go"
)

func ExampleNew() {
	client, err := fincra.New(
		fincra.WithSecretKey("sk_test_..."),
		fincra.WithBusinessID("64f1c9ef838af8cda0211d9d"),
		fincra.WithEnvironment(fincra.EnvironmentSandbox),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(client.BaseURL())
	// Output: https://sandboxapi.fincra.com
}

func ExampleNewWithConfig() {
	cfg := fincra.DefaultConfig()
	cfg.SecretKey = "sk_test_..."
	cfg.BusinessID = "64f1c9ef838af8cda0211d9d"

	client, err := fincra.NewWithConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(client.BaseURL())
	// Output: https://sandboxapi.fincra.com
}

func ExampleVerifyWebhookSignature() {
	payload := []byte(`{"event":"payout.successful"}`)
	signature := fincra.SignWebhookPayload("whsec_test_51H8xQ2", payload)

	fmt.Println(fincra.VerifyWebhookSignature("whsec_test_51H8xQ2", payload, signature))
	// Output: true
}

package fincra_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fincra "github.com/1Btln/fincra-go"
	"github.com/1Btln/fincra-go/beneficiaries"
	"github.com/1Btln/fincra-go/checkout"
	"github.com/1Btln/fincra-go/conversions"
	"github.com/1Btln/fincra-go/directcharge"
	"github.com/1Btln/fincra-go/mandates"
	"github.com/1Btln/fincra-go/virtualaccounts"
)

func decodeRequest[T any](t *testing.T, r *http.Request) T {
	t.Helper()
	var body T
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	return body
}

func respond(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, body)
}

type handlerRoundTripper struct{ handler http.Handler }

func (rt handlerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()
	rt.handler.ServeHTTP(recorder, req)
	return recorder.Result(), nil
}

func newClientWithHandler(t *testing.T, handler http.Handler, opts ...fincra.Option) *fincra.Client {
	t.Helper()
	cfg := fincra.DefaultConfig()
	cfg.SecretKey = "sk_test_123"
	cfg.PublicKey = "pk_test_456"
	cfg.BusinessID = "biz_789"
	cfg.BaseURL = "http://fincra.test"
	cfg.HTTPClient = &http.Client{Transport: handlerRoundTripper{handler: handler}}
	client, err := fincra.NewWithConfig(cfg, opts...)
	if err != nil {
		t.Fatalf("NewWithConfig: %v", err)
	}
	return client
}

func assertCredentials(t *testing.T, r *http.Request, businessID string) {
	t.Helper()
	if got := r.Header.Get("api-key"); got != "sk_test_123" {
		t.Errorf("api-key = %q, want configured secret", got)
	}
	if got := r.Header.Get("x-business-id"); got != businessID {
		t.Errorf("x-business-id = %q, want %q", got, businessID)
	}
}

func TestBeneficiariesServiceContract(t *testing.T) {
	requests := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if got := r.URL.Path; got != "/profile/beneficiaries/business/biz_789/ben_1" && got != "/profile/beneficiaries/business/biz_789" {
			t.Errorf("path = %q", got)
		}
		assertCredentials(t, r, "biz_789")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/profile/beneficiaries/business/biz_789":
			if r.URL.Query().Get("page") != "2" || r.URL.Query().Get("perPage") != "5" {
				t.Errorf("pagination query = %v", r.URL.Query())
			}
			respond(w, `{"success":true,"data":{"results":[{"_id":"ben_1"}],"total":1}}`)
		case r.Method == http.MethodPost:
			body := decodeRequest[map[string]any](t, r)
			if body["accountHolderName"] != "Ada Lovelace" {
				t.Errorf("request body = %v", body)
			}
			respond(w, `{"success":true,"data":{"_id":"ben_1"}}`)
		case r.Method == http.MethodPatch:
			respond(w, `{"success":true,"data":{"_id":"ben_1"}}`)
		case r.Method == http.MethodDelete:
			respond(w, `{"success":true,"message":"deleted"}`)
		case r.Method == http.MethodGet:
			respond(w, `{"success":true,"data":{"_id":"ben_1"}}`)
		default:
			t.Errorf("unexpected method/path: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	})

	client := newClientWithHandler(t, handler)
	ctx := context.Background()
	list, err := client.Beneficiaries.List(ctx, &beneficiaries.ListParams{Page: 2, PerPage: 5})
	if err != nil || len(list.Data.Results) != 1 || list.Data.Results[0].ID != "ben_1" {
		t.Fatalf("List() = %+v, %v", list, err)
	}
	create := &beneficiaries.CreateBeneficiaryRequest{
		FirstName: "Ada", AccountHolderName: "Ada Lovelace", Type: "individual", Currency: "NGN",
		PaymentDestination: "bank_account", DestinationAddress: "0123456789",
	}
	if _, err := client.Beneficiaries.Create(ctx, create); err != nil {
		t.Fatalf("Create(): %v", err)
	}
	if _, err := client.Beneficiaries.Get(ctx, "ben_1"); err != nil {
		t.Fatalf("Get(): %v", err)
	}
	if _, err := client.Beneficiaries.Update(ctx, "ben_1", create); err != nil {
		t.Fatalf("Update(): %v", err)
	}
	if _, err := client.Beneficiaries.Delete(ctx, "ben_1"); err != nil {
		t.Fatalf("Delete(): %v", err)
	}
	if requests != 5 {
		t.Errorf("requests = %d, want 5", requests)
	}
}

func TestCheckoutServiceContract(t *testing.T) {
	requests := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		assertCredentials(t, r, "biz_789")
		switch r.URL.Path {
		case "/checkout/payments":
			if r.Method != http.MethodPost || r.Header.Get("x-pub-key") != "pk_test_456" {
				t.Errorf("checkout request = %s, public key %q", r.Method, r.Header.Get("x-pub-key"))
			}
			body := decodeRequest[map[string]any](t, r)
			if body["currency"] != "NGN" || body["amount"] != float64(1200) {
				t.Errorf("checkout body = %v", body)
			}
			respond(w, `{"status":true,"message":"Hosted link generated","data":{"link":"https://checkout.example/pay/abc","payCode":"abc","reference":"order-1"}}`)
		case "/checkout/data/fees":
			body := decodeRequest[map[string]any](t, r)
			if body["type"] != "card" || body["amount"] != float64(1200) {
				t.Errorf("fee body = %v", body)
			}
			respond(w, `{"success":true,"data":{"fee":25,"vat":1.75,"amountToSettle":1173.25}}`)
		case "/checkout/payments/merchant-reference/order-1":
			if r.Method != http.MethodGet {
				t.Errorf("verify method = %s, want GET", r.Method)
			}
			respond(w, `{"status":true,"data":{"reference":"order-1","status":"success"}}`)
		default:
			t.Errorf("unexpected checkout path %q", r.URL.Path)
			http.NotFound(w, r)
		}
	})

	client := newClientWithHandler(t, handler)
	ctx := context.Background()
	created, err := client.Checkout.Initiate(ctx, &checkout.InitiateRequest{
		Currency: "NGN", Amount: 1200, Customer: &checkout.Customer{Name: "Ada Lovelace", Email: "ada@example.test"},
		PaymentMethods: []string{"card"}, Reference: "order-1",
	})
	if err != nil || created.Data.Link == "" {
		t.Fatalf("Initiate() = %+v, %v", created, err)
	}
	fee, err := client.Checkout.CalculateFee(ctx, &checkout.FeeRequest{Currency: "NGN", Type: "card", Amount: 1200})
	if err != nil || fee.Data.Fee != 25 {
		t.Fatalf("CalculateFee() = %+v, %v", fee, err)
	}
	verified, err := client.Checkout.Verify(ctx, "order-1")
	if err != nil || verified.Data["reference"] != "order-1" {
		t.Fatalf("Verify() = %+v, %v", verified, err)
	}
	if requests != 3 {
		t.Errorf("requests = %d, want 3", requests)
	}
}

func TestCheckoutPropagatesAPIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		respond(w, `{"message":"payment not found","errorType":"NOT_FOUND"}`)
	})

	client := newClientWithHandler(t, handler)
	_, err := client.Checkout.Verify(context.Background(), "missing")
	if _, ok := fincra.IsAPIError(err); !ok {
		t.Fatalf("Verify() error = %v, want *fincra.APIError", err)
	}
}

func TestConversionsServiceContract(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/quotes/generate":
			if r.Method != http.MethodPost {
				t.Errorf("quote method = %s, want POST", r.Method)
			}
			body := decodeRequest[map[string]any](t, r)
			if body["business"] != "biz_789" || body["transactionType"] != "conversion" {
				t.Errorf("quote body = %v", body)
			}
			respond(w, `{"success":true,"data":{"sourceCurrency":"USD","destinationCurrency":"NGN","reference":"quote-1","rate":1500,"expireAt":"soon"}}`)
		case "/conversions/initiate":
			body := decodeRequest[map[string]any](t, r)
			if body["quoteReference"] != "quote-1" || body["customerReference"] != "conversion-1" || body["business"] != "biz_789" {
				t.Errorf("conversion body = %v", body)
			}
			respond(w, `{"success":true,"data":{"reference":"conversion-1","status":"successful"}}`)
		case "/quotes/treasury-orders/rates":
			if r.URL.Query().Get("currencyPair") != "USD-NGN" || r.URL.Query().Get("page") != "3" {
				t.Errorf("rates query = %v", r.URL.Query())
			}
			respond(w, `{"success":true,"data":{"rates":[{"baseCurrency":"USD","quoteCurrency":"NGN","side":"buy","price":1500}],"pagination":{"page":3}}}`)
		default:
			t.Errorf("unexpected conversions path %q", r.URL.Path)
			http.NotFound(w, r)
		}
	})

	client := newClientWithHandler(t, handler)
	ctx := context.Background()
	quote, err := client.Conversions.GenerateQuote(ctx, &conversions.QuoteRequest{
		SourceCurrency: "USD", DestinationCurrency: "NGN", Amount: "10", Action: "send", TransactionType: "conversion",
		PaymentDestination: "fliqpay_wallet",
	})
	if err != nil || quote.Data.Reference != "quote-1" {
		t.Fatalf("GenerateQuote() = %+v, %v", quote, err)
	}
	converted, err := client.Conversions.Initiate(ctx, &conversions.InitiateRequest{QuoteReference: "quote-1", CustomerReference: "conversion-1"})
	if err != nil || converted.Data.Status != "successful" {
		t.Fatalf("Initiate() = %+v, %v", converted, err)
	}
	rates, err := client.Conversions.ListRates(ctx, &conversions.RateParams{CurrencyPair: "usd-ngn", Page: 3})
	if err != nil || len(rates.Data.Rates) != 1 || rates.Data.Rates[0].Price != 1500 {
		t.Fatalf("ListRates() = %+v, %v", rates, err)
	}
}

func TestDirectChargeServiceContract(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertCredentials(t, r, "biz_789")
		switch r.URL.Path {
		case "/checkout/charges":
			body := decodeRequest[map[string]any](t, r)
			card := body["card"].(map[string]any)
			if body["reference"] != "charge-1" || card["number"] != "4111111111111111" {
				t.Errorf("charge body = %v", body)
			}
			respond(w, `{"status":true,"data":{"id":123,"reference":"charge-1","status":"pending","auth_model":"OTP"}}`)
		case "/checkout/charges/123/authorize":
			if r.Method != http.MethodPost {
				t.Errorf("authorize method = %s", r.Method)
			}
			body := decodeRequest[map[string]string](t, r)
			if body["otp"] != "123456" {
				t.Errorf("authorize body = %v", body)
			}
			respond(w, `{"status":true,"data":{"id":123,"status":"success"}}`)
		case "/checkout/charges/merchant-reference/charge-1":
			if r.Method != http.MethodGet {
				t.Errorf("verify method = %s", r.Method)
			}
			respond(w, `{"status":true,"data":{"id":123,"reference":"charge-1","status":"success"}}`)
		default:
			t.Errorf("unexpected charge path %q", r.URL.Path)
			http.NotFound(w, r)
		}
	})

	client := newClientWithHandler(t, handler)
	ctx := context.Background()
	charge, err := client.DirectCharge.Initiate(ctx, &directcharge.InitiateRequest{
		Amount: 2500, Reference: "charge-1", Currency: "NGN", Type: "card",
		Customer: &directcharge.Customer{Name: "Ada Lovelace", Email: "ada@example.test"},
		Card:     &directcharge.Card{Number: "4111111111111111", CVV: "123", ExpiryMonth: "01", ExpiryYear: "2030"},
	})
	if err != nil || charge.Data.ID != 123 || charge.Data.AuthModel != "OTP" {
		t.Fatalf("Initiate() = %+v, %v", charge, err)
	}
	if _, err := client.DirectCharge.Authorize(ctx, "123", &directcharge.AuthorizeRequest{OTP: "123456"}, ""); err != nil {
		t.Fatalf("Authorize(): %v", err)
	}
	verified, err := client.DirectCharge.Verify(ctx, "charge-1")
	if err != nil || verified.Data.Status != "success" {
		t.Fatalf("Verify() = %+v, %v", verified, err)
	}
}

func TestMandatesServiceContract(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/mandate-mgt/mandates/":
			body := decodeRequest[map[string]any](t, r)
			if body["currency"] != "NGN" || body["customer"] == nil {
				t.Errorf("create mandate body = %v", body)
			}
			respond(w, `{"amount":1000,"reference":"mandate-1","status":"initiated"}`)
		case "/mandate-mgt/mandates/reference/mandate-1":
			respond(w, `{"amount":1000,"reference":"mandate-1","status":"approved"}`)
		case "/v2/mandate-mgt/mandates/payment":
			if r.Method != http.MethodPost {
				t.Errorf("debit method = %s", r.Method)
			}
			body := decodeRequest[map[string]any](t, r)
			if body["mandateReference"] != "mandate-1" || body["reference"] != "debit-1" {
				t.Errorf("debit body = %v", body)
			}
			respond(w, `{"status":true,"data":{"reference":"debit-1","status":"successful"}}`)
		case "/v2/mandate-mgt/payments/debit-1":
			respond(w, `{"status":true,"data":{"reference":"debit-1","status":"successful"}}`)
		default:
			t.Errorf("unexpected mandate path %q", r.URL.Path)
			http.NotFound(w, r)
		}
	})

	client := newClientWithHandler(t, handler)
	ctx := context.Background()
	mandate, err := client.Mandates.Create(ctx, &mandates.CreateRequest{
		Amount: "1000", Description: "monthly service", StartDate: "2030-01-01", EndDate: "2030-12-31",
		BusinessAddress: "Lagos", Currency: "NGN", Customer: &mandates.MandateCustomer{
			AccountNumber: "0123456789", AccountName: "Ada Lovelace", BankCode: "001",
			Address: "Lagos", Email: "ada@example.test", Phone: "08000000000",
		},
	})
	if err != nil || mandate.Reference != "mandate-1" {
		t.Fatalf("Create() = %+v, %v", mandate, err)
	}
	got, err := client.Mandates.Get(ctx, "mandate-1")
	if err != nil || got.Status != "approved" {
		t.Fatalf("Get() = %+v, %v", got, err)
	}
	debitReq := &mandates.DebitRequest{
		Amount: "100", Description: "installment", MandateReference: "mandate-1", Reference: "debit-1",
		Currency: "NGN", BankCode: "001", AccountNumber: "0123456789", InitiatorAccountName: "Ada Lovelace", BeneficiaryNarration: "monthly",
	}
	if _, err := client.Mandates.Debit(ctx, debitReq); err != nil {
		t.Fatalf("Debit(): %v", err)
	}
	if _, err := client.Mandates.VerifyDebit(ctx, "debit-1"); err != nil {
		t.Fatalf("VerifyDebit(): %v", err)
	}
}

func TestVirtualAccountsServiceContract(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/profile/virtual-accounts/requests":
			if r.Method == http.MethodPost {
				body := decodeRequest[map[string]any](t, r)
				if body["currency"] != "USD" || body["KYCInformation"] == nil {
					t.Errorf("virtual account request body = %v", body)
				}
				respond(w, `{"success":true,"data":{"_id":"va-1","currency":"USD","status":"pending"}}`)
			} else {
				respond(w, `{"success":true,"data":[{"currency":"USD","available":true}]}`)
			}
		case "/profile/virtual-accounts/transfer":
			body := decodeRequest[map[string]string](t, r)
			if body["amount"] != "2500" || body["expiresAt"] != "30" {
				t.Errorf("temporary account body = %v", body)
			}
			respond(w, `{"status":true,"data":{"accountNumber":"1234567890","reference":"temp-1"}}`)
		case "/profile/virtual-accounts/va-1":
			respond(w, `{"success":true,"data":{"_id":"va-1","accountNumber":"1234567890","currency":"USD"}}`)
		case "/profile/virtual-accounts/":
			if r.URL.Query().Get("currency") != "usd" {
				t.Errorf("currency query = %v", r.URL.Query())
			}
			respond(w, `{"success":true,"data":{"results":[{"_id":"va-1","currency":"USD"}],"total":1}}`)
		default:
			t.Errorf("unexpected virtual account path %q", r.URL.Path)
			http.NotFound(w, r)
		}
	})

	client := newClientWithHandler(t, handler)
	ctx := context.Background()
	requested, err := client.VirtualAccounts.Request(ctx, &virtualaccounts.Request{
		Currency: "USD", AccountType: "individual", KYCInformation: map[string]any{"firstName": "Ada"},
	})
	if err != nil || requested.Data.ID != "va-1" {
		t.Fatalf("Request() = %+v, %v", requested, err)
	}
	if _, err := client.VirtualAccounts.RequestOptions(ctx); err != nil {
		t.Fatalf("RequestOptions(): %v", err)
	}
	temporary, err := client.VirtualAccounts.RequestTemporary(ctx, &virtualaccounts.TemporaryRequest{
		Amount: "2500", ExpiresAt: "30", MerchantReference: "order-1",
	})
	if err != nil || temporary.Data.Reference != "temp-1" {
		t.Fatalf("RequestTemporary() = %+v, %v", temporary, err)
	}
	account, err := client.VirtualAccounts.Get(ctx, "va-1")
	if err != nil || account.Data.AccountNumber != "1234567890" {
		t.Fatalf("Get() = %+v, %v", account, err)
	}
	list, err := client.VirtualAccounts.List(ctx, "USD")
	if err != nil || list.Data.Total != 1 {
		t.Fatalf("List() = %+v, %v", list, err)
	}
}

func TestDirectChargeLogsMaskCardNumber(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respond(w, `{"status":true,"data":{"reference":"charge-1"}}`)
	})

	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	client := newClientWithHandler(t, handler,
		fincra.WithLogger(logger), fincra.WithLogDepth(fincra.LogDepthBody), fincra.WithLogLevel(slog.LevelDebug),
	)
	_, err := client.DirectCharge.Initiate(context.Background(), &directcharge.InitiateRequest{
		Amount: 100, Reference: "charge-1", Currency: "NGN", Type: "card",
		Customer: &directcharge.Customer{Name: "Ada", Email: "ada@example.test"},
		Card:     &directcharge.Card{Number: "4111111111111111", CVV: "123", ExpiryMonth: "01", ExpiryYear: "2030"},
	})
	if err != nil {
		t.Fatalf("Initiate(): %v", err)
	}
	if strings.Contains(logs.String(), "4111111111111111") {
		t.Fatalf("logs contain unmasked card number: %s", logs.String())
	}
	if !strings.Contains(logs.String(), "[REDACTED]") {
		t.Fatalf("logs do not show redaction marker: %s", logs.String())
	}
}

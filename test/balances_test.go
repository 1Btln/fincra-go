package fincra_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	fincra "github.com/Ledgenta/fincra-go"
)

const walletsJSON = `{
  "success": true,
  "message": "Wallets fetched successfully",
  "data": [
    {
      "id": 105146,
      "_id": "105146",
      "business": "64f1c9ef838af8cda0211d9d",
      "businessId": 9056,
      "ledgerBalance": 1999965363.990817,
      "availableBalance": 1999792413.990817,
      "lockedBalance": 172950,
      "rollingReserveBalance": 0,
      "walletNumber": 100000190560001,
      "currency": "NGN",
      "status": "enabled",
      "createdAt": "2023-09-01T11:24:32.000Z",
      "updatedAt": "2024-07-09T12:39:05.000Z"
    },
    {
      "id": 105147,
      "_id": "105147",
      "business": "64f1c9ef838af8cda0211d9d",
      "businessId": 9056,
      "ledgerBalance": 4145.558824,
      "availableBalance": 4145.558824,
      "lockedBalance": 0,
      "rollingReserveBalance": 0,
      "walletNumber": 100000190560002,
      "currency": "GBP",
      "status": "enabled",
      "createdAt": "2023-09-01T11:24:32.000Z",
      "updatedAt": "2024-03-28T10:57:02.000Z"
    }
  ]
}`

func TestListWallets(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wallets" {
			t.Errorf("path = %s, want /wallets", r.URL.Path)
		}
		if got := r.URL.Query().Get("businessID"); got != "biz_789" {
			t.Errorf("businessID = %q, want configured business ID", got)
		}
		_, _ = io.WriteString(w, walletsJSON)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	resp, err := client.Balances.ListWallets(context.Background(), "")
	if err != nil {
		t.Fatalf("ListWallets: %v", err)
	}
	if !resp.Success {
		t.Error("Success = false, want true")
	}
	if len(resp.Data) != 2 {
		t.Fatalf("len(Data) = %d, want 2", len(resp.Data))
	}

	wallet := resp.Data[0]
	if wallet.ID != 105146 || wallet.WalletNumber != 100000190560001 {
		t.Errorf("wallet identifiers = %+v", wallet)
	}
	if wallet.Currency != "NGN" || wallet.Status != "enabled" {
		t.Errorf("wallet = %+v", wallet)
	}
	if wallet.LedgerBalance != 1999965363.990817 {
		t.Errorf("LedgerBalance = %v", wallet.LedgerBalance)
	}
	if wallet.CreatedAt.IsZero() || wallet.UpdatedAt.IsZero() {
		t.Errorf("timestamps not decoded: %+v", wallet)
	}
}

func TestListWalletsExplicitBusinessID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("businessID"); got != "biz_other" {
			t.Errorf("businessID = %q, want biz_other", got)
		}
		_, _ = io.WriteString(w, `{"success":true,"data":[]}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if _, err := client.Balances.ListWallets(context.Background(), "biz_other"); err != nil {
		t.Fatalf("ListWallets: %v", err)
	}
}

func TestListWalletsMissingBusinessID(t *testing.T) {
	cfg := fincra.DefaultConfig()
	cfg.SecretKey = "sk_test"
	cfg.BaseURL = "https://example.test"
	client, err := fincra.NewWithConfig(cfg)
	if err != nil {
		t.Fatalf("NewWithConfig: %v", err)
	}

	_, err = client.Balances.ListWallets(context.Background(), "")
	if !errors.Is(err, fincra.ErrMissingRequiredParameter) {
		t.Fatalf("error = %v, want fincra.ErrMissingRequiredParameter", err)
	}
}

func TestGetWallet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wallets/wal_123" {
			t.Errorf("path = %s, want /wallets/wal_123", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{
			"success": true,
			"message": "Wallet fetched successfully",
			"data": {
				"id": 621,
				"_id": "60ed8883994a6d00195e1017",
				"business": "60ed8883c89ea214999fa9bf",
				"businessId": 186,
				"ledgerBalance": 900,
				"availableBalance": 300,
				"lockedBalance": 600,
				"rollingReserveBalance": 0,
				"walletNumber": 10000011860004,
				"currency": "NGN",
				"status": "enabled",
				"createdAt": "2023-09-01T11:24:32.000Z",
				"updatedAt": "2024-07-09T12:39:05.000Z"
			}
		}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	resp, err := client.Balances.GetWallet(context.Background(), "wal_123")
	if err != nil {
		t.Fatalf("GetWallet: %v", err)
	}
	if resp.Data.ID != 621 || resp.Data.AvailableBalance != 300 {
		t.Errorf("wallet = %+v", resp.Data)
	}

	if _, err := client.Balances.GetWallet(context.Background(), " "); !errors.Is(err, fincra.ErrMissingRequiredParameter) {
		t.Fatalf("empty wallet ID error = %v, want fincra.ErrMissingRequiredParameter", err)
	}
}

func TestGetBalanceHistory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wallets/logs" {
			t.Errorf("path = %s, want /wallets/logs", r.URL.Path)
		}
		query := r.URL.Query()
		want := map[string]string{
			"business": "biz_789",
			"action":   "debit",
			"dateFrom": "2024-01-01",
			"dateTo":   "2024-10-10",
			"page":     "2",
			"perPage":  "50",
			"amount":   "5200",
		}
		for key, value := range want {
			if got := query.Get(key); got != value {
				t.Errorf("%s = %q, want %q", key, got, value)
			}
		}
		_, _ = io.WriteString(w, `{
			"success": true,
			"message": "Wallet logs fetched successfully",
			"data": {
				"results": [
					{
						"id": 116077,
						"reference": "88f3ec7e57784d82",
						"createdAt": "2024-06-19T14:28:51.000Z",
						"description": "disbursement initiated",
						"action": "debit",
						"amount": 5220,
						"currency": "NGN",
						"method": "wallet",
						"reason": "payout initiated",
						"merchantDescription": "Test",
						"availableBalanceBefore": 1999802853.99,
						"availableBalanceAfter": 1999797633.99,
						"ledgerBalanceBefore": 1999975803.99,
						"ledgerBalanceAfter": 1999975803.99,
						"lockedBalanceBefore": 172950,
						"lockedBalanceAfter": 178170,
						"rollingReserveBalanceBefore": 0,
						"rollingReserveBalanceAfter": 0
					}
				],
				"total": 192
			}
		}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	resp, err := client.Balances.GetBalanceHistory(context.Background(), &fincra.BalanceHistoryParams{
		Action:   "DEBIT",
		DateFrom: "2024-01-01",
		DateTo:   "2024-10-10",
		Page:     2,
		PerPage:  50,
		Amount:   "5200",
	})
	if err != nil {
		t.Fatalf("GetBalanceHistory: %v", err)
	}
	if resp.Data.Total != 192 || len(resp.Data.Results) != 1 {
		t.Fatalf("data = %+v", resp.Data)
	}
	entry := resp.Data.Results[0]
	if entry.Action != "debit" || entry.Amount != 5220 {
		t.Errorf("entry = %+v", entry)
	}
	if entry.MerchantDescription == nil || *entry.MerchantDescription != "Test" {
		t.Errorf("MerchantDescription = %v", entry.MerchantDescription)
	}
}

func TestGetBalanceHistoryValidation(t *testing.T) {
	client := newTestClient(t, "https://example.test")

	if _, err := client.Balances.GetBalanceHistory(context.Background(), &fincra.BalanceHistoryParams{Action: "sideways"}); !errors.Is(err, fincra.ErrInvalidParameter) {
		t.Errorf("invalid action error = %v, want fincra.ErrInvalidParameter", err)
	}
	if _, err := client.Balances.GetBalanceHistory(context.Background(), &fincra.BalanceHistoryParams{DateFrom: "01/01/2024"}); !errors.Is(err, fincra.ErrInvalidParameter) {
		t.Errorf("invalid date error = %v, want fincra.ErrInvalidParameter", err)
	}
}

func TestGetAccountStatement(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wallets/logs/statement" {
			t.Errorf("path = %s, want /wallets/logs/statement", r.URL.Path)
		}
		if got := r.Header.Get("x-business-id"); got != "biz_override" {
			t.Errorf("x-business-id = %q, want biz_override", got)
		}
		query := r.URL.Query()
		want := map[string]string{
			"dateFrom": "2024-01-01",
			"dateTo":   "2024-10-10",
			"currency": "KES",
			"perPage":  "5",
			"page":     "2",
		}
		for key, value := range want {
			if got := query.Get(key); got != value {
				t.Errorf("%s = %q, want %q", key, got, value)
			}
		}
		_, _ = io.WriteString(w, `{
			"success": true,
			"message": "Wallet Log Statement fetched successfully",
			"data": {
				"pagination": {
					"current_page": 1,
					"page_size": 5,
					"count": 1,
					"total_count": 6,
					"total_pages": 2,
					"next_page": "?page=2",
					"previous_page": null
				},
				"openingAvailableBalance": 0,
				"openingLedgerBalance": 0,
				"closingAvailableBalance": 51990,
				"closingLedgerBalance": 1217295,
				"dateFrom": "2024-01-01",
				"dateTo": "2024-10-10",
				"business": "Isabelle Navarro",
				"businessAddress": null,
				"businessEmail": "joy@fincra.com",
				"results": [
					{
						"createdAt": "2024-10-09T21:32:51.000Z",
						"description": "collection successful",
						"reference": "b7b73b5e-181a-4be6-b1ec-6488b1a99c9d",
						"merchantDescription": null,
						"merchantReference": null,
						"reason": "topup successful",
						"action": "credit",
						"transactionType": "collection",
						"availableBalanceBefore": 0,
						"availableBalanceAfter": 51990,
						"ledgerBalanceBefore": 0,
						"ledgerBalanceAfter": 51990,
						"lockedBalanceBefore": 0,
						"lockedBalanceAfter": 0,
						"rollingReserveBalanceBefore": 0,
						"rollingReserveBalanceAfter": 0,
						"sourceCurrency": "KES",
						"destinationCurrency": "KES",
						"sourceAmount": 52000,
						"destinationAmount": 52000,
						"fee": 10,
						"recipientDetails": {
							"bankAccountName": "FINCRA TECHNOLOGIES LT",
							"bankCode": "76001",
							"bankName": "United Bank for Africa Kenya",
							"bankAccountNumber": "55010160011115"
						},
						"senderDetails": {
							"bankAccountName": "STEPHEN DAMILOLA IBIDUN",
							"bankCode": "000004",
							"bankName": "UNITED BANK FOR AFRICA",
							"bankAccountNumber": "228046775927"
						},
						"metadata": {
							"customerId": null
						}
					}
				]
			}
		}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	resp, err := client.Balances.GetAccountStatement(context.Background(), &fincra.AccountStatementParams{
		BusinessID: "biz_override",
		DateFrom:   "2024-01-01",
		DateTo:     "2024-10-10",
		Currency:   "kes",
		PerPage:    5,
		Page:       2,
	})
	if err != nil {
		t.Fatalf("GetAccountStatement: %v", err)
	}
	if resp.Data.Pagination.TotalCount != 6 || resp.Data.Pagination.NextPage != "?page=2" {
		t.Errorf("pagination = %+v", resp.Data.Pagination)
	}
	if len(resp.Data.Results) != 1 {
		t.Fatalf("len(Results) = %d, want 1", len(resp.Data.Results))
	}
	entry := resp.Data.Results[0]
	if entry.SourceAmount != 52000 || entry.Fee != 10 {
		t.Errorf("entry amounts = %+v", entry)
	}
	if entry.RecipientDetails == nil || entry.RecipientDetails.BankCode != "76001" {
		t.Errorf("RecipientDetails = %+v", entry.RecipientDetails)
	}
	if entry.Metadata == nil {
		t.Error("Metadata is nil")
	}
}

func TestGetAccountStatementDefaultsPerPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("perPage"); got != "20" {
			t.Errorf("perPage = %q, want 20", got)
		}
		_, _ = io.WriteString(w, `{"success":true,"data":{}}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if _, err := client.Balances.GetAccountStatement(context.Background(), &fincra.AccountStatementParams{
		DateFrom: "2024-01-01",
		DateTo:   "2024-01-31",
		Currency: "NGN",
	}); err != nil {
		t.Fatalf("GetAccountStatement: %v", err)
	}
}

func TestGetAccountStatementValidation(t *testing.T) {
	client := newTestClient(t, "https://example.test")

	tests := []struct {
		name   string
		params *fincra.AccountStatementParams
	}{
		{name: "nil params", params: nil},
		{name: "missing dates", params: &fincra.AccountStatementParams{Currency: "NGN"}},
		{name: "missing currency", params: &fincra.AccountStatementParams{DateFrom: "2024-01-01", DateTo: "2024-01-31"}},
		{name: "bad date", params: &fincra.AccountStatementParams{DateFrom: "01/01/2024", DateTo: "2024-01-31", Currency: "NGN"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.Balances.GetAccountStatement(context.Background(), tt.params)
			if !errors.Is(err, fincra.ErrMissingRequiredParameter) && !errors.Is(err, fincra.ErrInvalidParameter) {
				t.Fatalf("error = %v, want a parameter validation error", err)
			}
		})
	}
}

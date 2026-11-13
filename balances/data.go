// Package balances exposes Fincra wallet, balance-history, statement, and top-up
// endpoints.
package balances

import "time"

// Wallet is a Fincra account balance. Fincra maintains one wallet per supported
// currency for a business.
type Wallet struct {
	ID       int64  `json:"id"`
	ObjectID string `json:"_id"`

	// Business is the Mongo-style business identifier; BusinessID is the
	// numeric identifier used by the API and dashboard.
	Business   string `json:"business"`
	BusinessID int64  `json:"businessId"`

	// Amounts are returned exactly as Fincra encodes them. They are float64
	// because the API mixes integers and decimals; convert to a decimal type
	// before using them for accounting.
	LedgerBalance         float64 `json:"ledgerBalance"`
	AvailableBalance      float64 `json:"availableBalance"`
	LockedBalance         float64 `json:"lockedBalance"`
	RollingReserveBalance float64 `json:"rollingReserveBalance"`

	WalletNumber int64     `json:"walletNumber"`
	Currency     string    `json:"currency"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// ListWalletsResponse is returned by Service.ListWallets.
type ListWalletsResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Data    []Wallet `json:"data"`
}

// WalletResponse is returned by Service.GetWallet.
type WalletResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    Wallet `json:"data"`
}

// BalanceHistoryParams filters Service.GetBalanceHistory. All fields are
// optional except the business identifier, which falls back to the configured
// BusinessID.
type BalanceHistoryParams struct {
	// BusinessID is the parent business or sub-account ID. Empty uses
	// Config.BusinessID.
	BusinessID string

	// Action filters by "credit" or "debit". Empty uses Fincra's default.
	Action string

	// DateFrom and DateTo are inclusive bounds formatted as YYYY-MM-DD (an
	// RFC3339 timestamp is also accepted).
	DateFrom string
	DateTo   string

	// Page and PerPage paginate the results.
	Page    int
	PerPage int

	// Amount filters by amount when set.
	Amount string
}

// BalanceEntry is a single credit or debit in the balance history.
type BalanceEntry struct {
	ID          int64     `json:"id"`
	Reference   string    `json:"reference"`
	CreatedAt   time.Time `json:"createdAt"`
	Description string    `json:"description"`
	Action      string    `json:"action"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Method      string    `json:"method"`
	Reason      string    `json:"reason"`

	MerchantDescription *string `json:"merchantDescription"`

	AvailableBalanceBefore      float64 `json:"availableBalanceBefore"`
	AvailableBalanceAfter       float64 `json:"availableBalanceAfter"`
	LedgerBalanceBefore         float64 `json:"ledgerBalanceBefore"`
	LedgerBalanceAfter          float64 `json:"ledgerBalanceAfter"`
	LockedBalanceBefore         float64 `json:"lockedBalanceBefore"`
	LockedBalanceAfter          float64 `json:"lockedBalanceAfter"`
	RollingReserveBalanceBefore float64 `json:"rollingReserveBalanceBefore"`
	RollingReserveBalanceAfter  float64 `json:"rollingReserveBalanceAfter"`
}

// BalanceHistoryData is the data payload of BalanceHistoryResponse.
type BalanceHistoryData struct {
	Results []BalanceEntry `json:"results"`
	Total   int64          `json:"total"`
}

// BalanceHistoryResponse is returned by Service.GetBalanceHistory.
type BalanceHistoryResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	Data    BalanceHistoryData `json:"data"`
}

// AccountStatementParams filters Service.GetAccountStatement.
type AccountStatementParams struct {
	// BusinessID overrides the x-business-id header for this request. Empty
	// uses Config.BusinessID.
	BusinessID string

	// DateFrom and DateTo are required and use YYYY-MM-DD (RFC3339 is also
	// accepted).
	DateFrom string
	DateTo   string

	// Currency is the wallet currency, for example "NGN", "USD", or "KES".
	Currency string

	// PerPage defaults to 20 when zero.
	PerPage int

	// Page selects a page of results. Zero uses the first page.
	Page int
}

// Pagination describes a paginated Fincra response.
type Pagination struct {
	CurrentPage  int    `json:"current_page"`
	PageSize     int    `json:"page_size"`
	Count        int    `json:"count"`
	TotalCount   int    `json:"total_count"`
	TotalPages   int    `json:"total_pages"`
	NextPage     string `json:"next_page"`
	PreviousPage string `json:"previous_page"`
}

// StatementAccountDetails describes a bank account taking part in a statement
// entry.
type StatementAccountDetails struct {
	BankAccountName   string `json:"bankAccountName"`
	BankCode          string `json:"bankCode"`
	BankName          string `json:"bankName"`
	BankAccountNumber string `json:"bankAccountNumber"`
}

// StatementEntry is a single line in an account statement.
type StatementEntry struct {
	CreatedAt           time.Time `json:"createdAt"`
	Description         string    `json:"description"`
	Reference           string    `json:"reference"`
	MerchantDescription *string   `json:"merchantDescription"`
	MerchantReference   *string   `json:"merchantReference"`
	Reason              string    `json:"reason"`
	Action              string    `json:"action"`
	TransactionType     string    `json:"transactionType"`

	AvailableBalanceBefore      float64 `json:"availableBalanceBefore"`
	AvailableBalanceAfter       float64 `json:"availableBalanceAfter"`
	LedgerBalanceBefore         float64 `json:"ledgerBalanceBefore"`
	LedgerBalanceAfter          float64 `json:"ledgerBalanceAfter"`
	LockedBalanceBefore         float64 `json:"lockedBalanceBefore"`
	LockedBalanceAfter          float64 `json:"lockedBalanceAfter"`
	RollingReserveBalanceBefore float64 `json:"rollingReserveBalanceBefore"`
	RollingReserveBalanceAfter  float64 `json:"rollingReserveBalanceAfter"`

	SourceCurrency      string                   `json:"sourceCurrency"`
	DestinationCurrency string                   `json:"destinationCurrency"`
	SourceAmount        float64                  `json:"sourceAmount"`
	DestinationAmount   float64                  `json:"destinationAmount"`
	Fee                 float64                  `json:"fee"`
	RecipientDetails    *StatementAccountDetails `json:"recipientDetails"`
	SenderDetails       *StatementAccountDetails `json:"senderDetails"`
	Metadata            map[string]any           `json:"metadata"`
}

// AccountStatementData is the data payload of AccountStatementResponse.
type AccountStatementData struct {
	Pagination              Pagination       `json:"pagination"`
	OpeningAvailableBalance float64          `json:"openingAvailableBalance"`
	OpeningLedgerBalance    float64          `json:"openingLedgerBalance"`
	ClosingAvailableBalance float64          `json:"closingAvailableBalance"`
	ClosingLedgerBalance    float64          `json:"closingLedgerBalance"`
	DateFrom                string           `json:"dateFrom"`
	DateTo                  string           `json:"dateTo"`
	Business                string           `json:"business"`
	BusinessAddress         *string          `json:"businessAddress"`
	BusinessEmail           string           `json:"businessEmail"`
	Results                 []StatementEntry `json:"results"`
}

// AccountStatementResponse is returned by Service.GetAccountStatement.
type AccountStatementResponse struct {
	Success bool                 `json:"success"`
	Message string               `json:"message"`
	Data    AccountStatementData `json:"data"`
}

// TopupHistoryParams filters Service.GetTopupHistory.
type TopupHistoryParams struct {
	BusinessID string
	Reference  string
	Page       int
	PerPage    int
}

// Topup is a wallet funding transaction.
type Topup struct {
	ID        int64     `json:"id"`
	Reference string    `json:"reference"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	Method    string    `json:"method"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TopupHistoryData is the data payload of TopupHistoryResponse.
type TopupHistoryData struct {
	Results []Topup `json:"results"`
	Total   int64   `json:"total"`
}

// TopupHistoryResponse is returned by Service.GetTopupHistory.
type TopupHistoryResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    TopupHistoryData `json:"data"`
}

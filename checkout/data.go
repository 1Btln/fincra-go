// Package checkout exposes Fincra hosted-checkout endpoints.
package checkout

// Customer identifies the person paying through checkout.
type Customer struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
}

// InitiateRequest creates a hosted checkout session.
type InitiateRequest struct {
	BusinessID            string         `json:"-"`
	Currency              string         `json:"currency"`
	Amount                int64          `json:"amount"`
	Customer              *Customer      `json:"customer"`
	PaymentMethods        []string       `json:"paymentMethods,omitempty"`
	Reference             string         `json:"reference,omitempty"`
	FeeBearer             string         `json:"feeBearer,omitempty"`
	SettlementDestination string         `json:"settlementDestination,omitempty"`
	RedirectURL           string         `json:"redirectUrl,omitempty"`
	DefaultPaymentMethod  string         `json:"defaultPaymentMethod,omitempty"`
	QuoteReference        string         `json:"quoteReference,omitempty"`
	Metadata              map[string]any `json:"metadata,omitempty"`
}

// PaymentLink describes the hosted payment page returned by Fincra.
type PaymentLink struct {
	Link      string `json:"link"`
	PayCode   string `json:"payCode"`
	Reference string `json:"reference,omitempty"`
}

// PaymentResponse is returned by Service.Initiate.
type PaymentResponse struct {
	Status  bool        `json:"status"`
	Success bool        `json:"success,omitempty"`
	Message string      `json:"message"`
	Data    PaymentLink `json:"data"`
}

// FeeRequest calculates a hosted-checkout fee.
type FeeRequest struct {
	BusinessID string `json:"-"`
	Currency   string `json:"currency"`
	Type       string `json:"type"`
	Amount     int64  `json:"amount"`
}

// FeeBreakdown contains the amounts returned for a checkout fee calculation.
type FeeBreakdown struct {
	Amount         float64 `json:"amount"`
	Fee            float64 `json:"fee"`
	VAT            float64 `json:"vat"`
	TotalFee       float64 `json:"totalFee"`
	AmountToSettle float64 `json:"amountToSettle"`
	AmountToCharge float64 `json:"amountToCharge"`
}

// FeeResponse is returned by Service.CalculateFee.
type FeeResponse struct {
	Success bool         `json:"success"`
	Status  bool         `json:"status,omitempty"`
	Message string       `json:"message"`
	Data    FeeBreakdown `json:"data"`
}

// TransactionResponse is returned by Service.Verify.
type TransactionResponse struct {
	Success bool           `json:"success"`
	Status  bool           `json:"status,omitempty"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}

// Package conversions exposes Fincra FX rates, quote, and conversion endpoints.
package conversions

// QuoteRequest requests a conversion or payout quote.
type QuoteRequest struct {
	SourceCurrency      string `json:"sourceCurrency"`
	DestinationCurrency string `json:"destinationCurrency"`
	Amount              string `json:"amount"`
	Action              string `json:"action"`
	TransactionType     string `json:"transactionType,omitempty"`
	Business            string `json:"business"`
	FeeBearer           string `json:"feeBearer,omitempty"`
	PaymentDestination  string `json:"paymentDestination,omitempty"`
	PaymentScheme       string `json:"paymentScheme,omitempty"`
	BeneficiaryType     string `json:"beneficiaryType,omitempty"`
	Delay               bool   `json:"delay,omitempty"`
}

// Quote contains the amounts and reference returned for a quote.
type Quote struct {
	SourceCurrency      string  `json:"sourceCurrency"`
	DestinationCurrency string  `json:"destinationCurrency"`
	SourceAmount        float64 `json:"sourceAmount"`
	DestinationAmount   float64 `json:"destinationAmount"`
	Action              string  `json:"action"`
	TransactionType     string  `json:"transactionType"`
	Fee                 float64 `json:"fee"`
	InitialAmount       float64 `json:"initialAmount"`
	QuotedAmount        float64 `json:"quotedAmount"`
	Rate                float64 `json:"rate"`
	AmountToCharge      float64 `json:"amountToCharge"`
	AmountToReceive     float64 `json:"amountToReceive"`
	Reference           string  `json:"reference"`
	ExpireAt            string  `json:"expireAt"`
}

// QuoteResponse is returned by Service.GenerateQuote.
type QuoteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    Quote  `json:"data"`
}

// InitiateRequest converts wallet funds using an existing quote.
type InitiateRequest struct {
	Business          string `json:"business,omitempty"`
	QuoteReference    string `json:"quoteReference"`
	CustomerReference string `json:"customerReference"`
}

// Conversion is the result of a wallet currency conversion.
type Conversion struct {
	ID                      int64   `json:"id"`
	ObjectID                string  `json:"_id"`
	BusinessName            string  `json:"businessName,omitempty"`
	Business                string  `json:"business"`
	AmountCharged           float64 `json:"amountCharged"`
	AmountReceived          float64 `json:"amountReceived"`
	Fee                     float64 `json:"fee"`
	SourceFee               float64 `json:"sourceFee,omitempty"`
	DestinationFee          float64 `json:"destinationFee,omitempty"`
	SourceCurrency          string  `json:"sourceCurrency"`
	DestinationCurrency     string  `json:"destinationCurrency"`
	Rate                    float64 `json:"rate"`
	SourceExchangeRate      float64 `json:"sourceExchangeRate,omitempty"`
	DestinationExchangeRate float64 `json:"destinationExchangeRate,omitempty"`
	SettlementDate          string  `json:"settlementDate,omitempty"`
	SettledAt               string  `json:"settledAt,omitempty"`
	SettlementDestination   string  `json:"settlementDestination,omitempty"`
	Description             string  `json:"description,omitempty"`
	Message                 string  `json:"message,omitempty"`
	IsSubAccount            int     `json:"isSubAccount,omitempty"`
	Status                  string  `json:"status"`
	Reference               string  `json:"reference"`
	CustomerReference       string  `json:"customerReference"`
	CreatedAt               string  `json:"createdAt"`
	UpdatedAt               string  `json:"updatedAt"`
}

// ConversionResponse is returned by Service.Initiate.
type ConversionResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	Data    Conversion `json:"data"`
}

// RateParams filters Service.ListRates.
type RateParams struct {
	CurrencyPair  string
	Currency      string
	Side          string
	BaseCurrency  string
	QuoteCurrency string
	Page          int
	PerPage       int
}

// ListParams filters Service.List.
type ListParams struct {
	BusinessID string
}

// ConversionListData is the data payload of ConversionListResponse.
type ConversionListData struct {
	Results []Conversion `json:"results"`
	Total   int          `json:"total"`
}

// ConversionListResponse is returned by Service.List.
type ConversionListResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	Data    ConversionListData `json:"data"`
}

// Rate is a buy or sell rate for a currency pair.
type Rate struct {
	BaseCurrency  string  `json:"baseCurrency"`
	QuoteCurrency string  `json:"quoteCurrency"`
	Side          string  `json:"side"`
	Price         float64 `json:"price"`
}

// RatePagination describes pagination metadata returned by Fincra.
type RatePagination struct {
	Page       int    `json:"page"`
	PerPage    int    `json:"perPage"`
	TotalItems int    `json:"totalItems"`
	TotalPages int    `json:"totalPages"`
	Timestamp  string `json:"timestamp"`
}

// RatesData is the data payload of RatesResponse.
type RatesData struct {
	Rates      []Rate         `json:"rates"`
	Pagination RatePagination `json:"pagination"`
}

// RatesResponse is returned by Service.ListRates.
type RatesResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Data    RatesData `json:"data"`
}

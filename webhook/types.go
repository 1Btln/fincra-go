package webhook

import "time"

// WebhookEventType identifies a documented Fincra webhook event.
type WebhookEventType string

const (
	WebhookPayoutSuccessful       WebhookEventType = "payout.successful"
	WebhookPayoutFailed           WebhookEventType = "payout.failed"
	WebhookCollectionSuccessful   WebhookEventType = "collection.successful"
	WebhookCollectionFailed       WebhookEventType = "collection.failed"
	WebhookConversionSuccessful   WebhookEventType = "conversion.successful"
	WebhookConversionFailed       WebhookEventType = "conversion.failed"
	WebhookVirtualAccountApproved WebhookEventType = "virtualaccount.approved"
	WebhookVirtualAccountIssued   WebhookEventType = "virtualaccount.issued"
	WebhookVirtualAccountChanged  WebhookEventType = "virtualaccount.changed"
	WebhookVirtualAccountDeclined WebhookEventType = "virtualaccount.declined"
	WebhookVirtualAccountClosed   WebhookEventType = "virtualaccount.closed"
	WebhookMandateApproved        WebhookEventType = "mandate.approved"
	WebhookChargeSuccessful       WebhookEventType = "charge.successful"
)

// WebhookEnvelope is the common Fincra webhook shape. Decode the raw request
// body into WebhookEnvelope[SpecificPayload] after verifying its signature.
type WebhookEnvelope[T any] struct {
	Event WebhookEventType `json:"event"`
	Data  T                `json:"data"`
}

// PayoutWebhookData is the payload for payout.successful and payout.failed.
type PayoutWebhookData struct {
	ID                  int64            `json:"id"`
	AmountCharged       float64          `json:"amountCharged"`
	AmountReceived      float64          `json:"amountReceived"`
	Recipient           WebhookRecipient `json:"recipient"`
	Fee                 float64          `json:"fee"`
	Rate                float64          `json:"rate"`
	PaymentScheme       string           `json:"paymentScheme"`
	PaymentDestination  string           `json:"paymentDestination"`
	SourceCurrency      string           `json:"sourceCurrency"`
	DestinationCurrency string           `json:"destinationCurrency"`
	Status              string           `json:"status"`
	CreatedAt           time.Time        `json:"createdAt"`
	UpdatedAt           time.Time        `json:"updatedAt"`
	Reference           string           `json:"reference"`
	CustomerReference   string           `json:"customerReference"`
	Reason              string           `json:"reason"`
	TraceID             string           `json:"traceId"`
	ValuedAt            *time.Time       `json:"valuedAt"`
}

// WebhookRecipient is a recipient embedded in payout webhook data.
type WebhookRecipient struct {
	Name          string `json:"name"`
	AccountNumber string `json:"accountNumber"`
	Type          string `json:"type"`
	Email         string `json:"email"`
}

// CollectionWebhookData is the payload for collection.successful/failed.
type CollectionWebhookData struct {
	Business              string    `json:"business"`
	VirtualAccount        string    `json:"virtualAccount"`
	SessionID             string    `json:"sessionId"`
	SenderBankName        string    `json:"senderBankName"`
	SenderAccountName     string    `json:"senderAccountName"`
	SenderAccountNumber   string    `json:"senderAccountNumber"`
	SourceCurrency        string    `json:"sourceCurrency"`
	DestinationCurrency   string    `json:"destinationCurrency"`
	SourceAmount          float64   `json:"sourceAmount"`
	DestinationAmount     float64   `json:"destinationAmount"`
	AmountReceived        float64   `json:"amountReceived"`
	Fee                   float64   `json:"fee"`
	CustomerName          string    `json:"customerName"`
	SettlementDestination string    `json:"settlementDestination"`
	Status                string    `json:"status"`
	Reason                string    `json:"reason"`
	PaymentScheme         string    `json:"paymentScheme"`
	Description           string    `json:"description"`
	InitiatedAt           time.Time `json:"initiatedAt"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
	Reference             string    `json:"reference"`
}

// ConversionWebhookData is the payload for conversion.successful/failed.
type ConversionWebhookData struct {
	ID                  int64     `json:"id"`
	ObjectID            string    `json:"_id"`
	Business            string    `json:"business"`
	AmountCharged       float64   `json:"amountCharged"`
	AmountReceived      float64   `json:"amountReceived"`
	Fee                 float64   `json:"fee"`
	SourceCurrency      string    `json:"sourceCurrency"`
	DestinationCurrency string    `json:"destinationCurrency"`
	Rate                float64   `json:"rate"`
	Price               float64   `json:"price"`
	Settlement          any       `json:"settlement"`
	Status              string    `json:"status"`
	Reference           string    `json:"reference"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// VirtualAccountWebhookData is the payload for virtual-account lifecycle events.
// Account information differs by country and account type, so it remains a map.
type VirtualAccountWebhookData struct {
	ID                 string         `json:"id"`
	Business           string         `json:"business"`
	IsSubAccount       bool           `json:"isSubAccount"`
	Currency           string         `json:"currency"`
	CurrencyType       string         `json:"currencyType"`
	Status             string         `json:"status"`
	AccountType        string         `json:"accountType"`
	Email              string         `json:"email"`
	AccountInformation map[string]any `json:"accountInformation"`
	Reason             string         `json:"reason"`
	AccountOpeningFee  float64        `json:"accountOpeningFee"`
	IsPermanent        bool           `json:"isPermanent"`
	VirtualAccountType string         `json:"virtualAccountType"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
}

// MandateWebhookData is the payload for mandate.approved.
type MandateWebhookData struct {
	Amount              float64   `json:"amount"`
	Description         string    `json:"description"`
	ResponseDescription string    `json:"responseDescription"`
	StartDate           time.Time `json:"startDate"`
	EndDate             time.Time `json:"endDate"`
	Status              string    `json:"status"`
	Reference           string    `json:"reference"`
	CreatedAt           time.Time `json:"createdAt"`
}

// ChargeWebhookData is the payload for charge.successful from checkout or direct charge.
type ChargeWebhookData struct {
	ChargeReference       string          `json:"chargeReference"`
	AmountToSettle        float64         `json:"amountToSettle"`
	ID                    int64           `json:"id"`
	Amount                float64         `json:"amount"`
	AmountExpected        float64         `json:"amountExpected"`
	AmountReceived        float64         `json:"amountReceived"`
	VarianceType          string          `json:"varianceType"`
	Currency              string          `json:"currency"`
	Fee                   float64         `json:"fee"`
	VAT                   float64         `json:"vat"`
	Message               string          `json:"message"`
	Status                string          `json:"status"`
	Reference             string          `json:"reference"`
	Description           string          `json:"description"`
	Type                  string          `json:"type"`
	Customer              WebhookCustomer `json:"customer"`
	Metadata              map[string]any  `json:"metadata"`
	SettlementDestination string          `json:"settlementDestination"`
	SettlementTime        string          `json:"settlementTime"`
	VirtualAccount        map[string]any  `json:"virtualAccount"`
}

// WebhookCustomer is the customer embedded in charge webhook data.
type WebhookCustomer struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
}

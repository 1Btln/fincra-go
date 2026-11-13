// Package directcharge exposes Fincra's payment-method charge endpoints.
package directcharge

// Customer identifies a customer being charged.
type Customer struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
}

// Card contains card details for a card charge. Direct card charging requires
// the merchant to meet Fincra's PCI-DSS requirements.
type Card struct {
	Number      string `json:"number"`
	CVV         string `json:"cvv"`
	ExpiryMonth string `json:"expiryMonth"`
	ExpiryYear  string `json:"expiryYear"`
	PIN         string `json:"pin,omitempty"`
}

// InitiateRequest starts a direct charge.
type InitiateRequest struct {
	BusinessID            string         `json:"-"`
	Amount                int64          `json:"amount"`
	Reference             string         `json:"reference"`
	Currency              string         `json:"currency"`
	Type                  string         `json:"type"`
	Customer              *Customer      `json:"customer"`
	Card                  *Card          `json:"card,omitempty"`
	Phone                 string         `json:"phone,omitempty"`
	Description           string         `json:"description,omitempty"`
	SettlementDestination string         `json:"settlementDestination,omitempty"`
	CallbackURL           string         `json:"callbackUrl,omitempty"`
	WebhookURL            string         `json:"webhookUrl,omitempty"`
	Operator              string         `json:"operator,omitempty"`
	QuoteReference        string         `json:"quoteReference,omitempty"`
	Metadata              map[string]any `json:"metadata,omitempty"`
}

// Authorization contains the available authorization method and redirect.
type Authorization struct {
	Mode     string         `json:"mode"`
	Redirect string         `json:"redirect"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Charge is a direct-charge transaction.
type Charge struct {
	ID                    int64          `json:"id"`
	Amount                float64        `json:"amount"`
	AmountReceived        float64        `json:"amountReceived"`
	Currency              string         `json:"currency"`
	Fee                   float64        `json:"fee"`
	Message               string         `json:"message"`
	Status                string         `json:"status"`
	Reference             string         `json:"reference"`
	Description           string         `json:"description"`
	Type                  string         `json:"type"`
	Customer              Customer       `json:"customer"`
	SettlementDestination string         `json:"settlementDestination"`
	AuthModel             string         `json:"auth_model"`
	Authorization         Authorization  `json:"authorization"`
	Metadata              map[string]any `json:"metadata,omitempty"`
}

// ChargeResponse is returned by Service.Initiate, Authorize, and Verify.
type ChargeResponse struct {
	Status  bool   `json:"status"`
	Success bool   `json:"success,omitempty"`
	Message string `json:"message"`
	Data    Charge `json:"data"`
}

// AVSDetails holds card address-verification details.
type AVSDetails struct {
	State   string `json:"state,omitempty"`
	City    string `json:"city,omitempty"`
	Country string `json:"country,omitempty"`
	Address string `json:"address,omitempty"`
	ZipCode string `json:"zip_code,omitempty"`
}

// AuthorizeRequest contains the challenge response required to authorize a charge.
type AuthorizeRequest struct {
	PIN     string      `json:"pin,omitempty"`
	OTP     string      `json:"otp,omitempty"`
	AVS     *AVSDetails `json:"avs,omitempty"`
	State   string      `json:"state,omitempty"`
	City    string      `json:"city,omitempty"`
	Country string      `json:"country,omitempty"`
	Address string      `json:"address,omitempty"`
	ZipCode string      `json:"zip_code,omitempty"`
}

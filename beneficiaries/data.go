// Package beneficiaries exposes Fincra's saved-recipient management endpoints.
package beneficiaries

import "time"

// Address is a beneficiary or bank address.
type Address struct {
	Country string `json:"country,omitempty"`
	State   string `json:"state,omitempty"`
	Zip     string `json:"zip,omitempty"`
	City    string `json:"city,omitempty"`
	Street  string `json:"street,omitempty"`
}

// Bank holds the beneficiary's bank details.
type Bank struct {
	Name      string   `json:"name,omitempty"`
	Code      string   `json:"code,omitempty"`
	SortCode  string   `json:"sortCode,omitempty"`
	SwiftCode string   `json:"swiftCode,omitempty"`
	Branch    string   `json:"branch,omitempty"`
	Address   *Address `json:"address,omitempty"`
}

// Beneficiary is a saved recipient.
type Beneficiary struct {
	ID                 string    `json:"_id"`
	Business           string    `json:"business"`
	FirstName          string    `json:"firstName"`
	LastName           string    `json:"lastName"`
	Email              string    `json:"email"`
	PhoneNumber        string    `json:"phoneNumber"`
	AccountHolderName  string    `json:"accountHolderName"`
	Bank               Bank      `json:"bank"`
	Address            *Address  `json:"address"`
	Type               string    `json:"type"`
	Currency           string    `json:"currency"`
	PaymentDestination string    `json:"paymentDestination"`
	DestinationAddress string    `json:"destinationAddress"`
	UniqueIdentifier   string    `json:"uniqueIdentifier"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// CreateBeneficiaryRequest creates a saved recipient.
type CreateBeneficiaryRequest struct {
	FirstName          string   `json:"firstName"`
	LastName           string   `json:"lastName,omitempty"`
	Email              string   `json:"email,omitempty"`
	PhoneNumber        string   `json:"phoneNumber,omitempty"`
	AccountHolderName  string   `json:"accountHolderName"`
	Bank               Bank     `json:"bank,omitempty"`
	Address            *Address `json:"address,omitempty"`
	Type               string   `json:"type"`
	Currency           string   `json:"currency"`
	PaymentDestination string   `json:"paymentDestination"`
	DestinationAddress string   `json:"destinationAddress"`
	UniqueIdentifier   string   `json:"uniqueIdentifier,omitempty"`
}

// UpdateBeneficiaryRequest updates a saved recipient. It has the same fields as
// creation.
type UpdateBeneficiaryRequest = CreateBeneficiaryRequest

// ListParams paginate Service.List.
type ListParams struct {
	BusinessID string
	Page       int
	PerPage    int
}

// BeneficiaryData is the data payload of BeneficiaryListResponse.
type BeneficiaryData struct {
	Results []Beneficiary `json:"results"`
	Total   int64         `json:"total"`
}

// BeneficiaryListResponse is returned by Service.List.
type BeneficiaryListResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    BeneficiaryData `json:"data"`
}

// BeneficiaryResponse is returned by Service.Get, Create, and Update.
type BeneficiaryResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    Beneficiary `json:"data"`
}

// ActionResponse is returned by Service.Delete.
type ActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// Package mandates exposes Fincra direct-debit mandate endpoints.
package mandates

// MandateCustomer is the bank-account owner on a mandate.
type MandateCustomer struct {
	AccountNumber string `json:"accountNumber"`
	AccountName   string `json:"accountName"`
	BankCode      string `json:"bankCode"`
	Address       string `json:"address"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
}

// CreateRequest creates a standing debit mandate. Amount is represented as a
// decimal string to avoid floating-point rounding in payment values.
type CreateRequest struct {
	Amount          string           `json:"amount"`
	Description     string           `json:"description"`
	StartDate       string           `json:"startDate"`
	EndDate         string           `json:"endDate"`
	BusinessAddress string           `json:"businessAddress"`
	Customer        *MandateCustomer `json:"customer"`
	Currency        string           `json:"currency"`
}

// Mandate is a direct-debit agreement.
type Mandate struct {
	Amount              float64 `json:"amount"`
	Description         string  `json:"description"`
	ResponseDescription string  `json:"responseDescription"`
	StartDate           string  `json:"startDate"`
	EndDate             string  `json:"endDate"`
	Status              string  `json:"status"`
	Reference           string  `json:"reference"`
	CreatedAt           string  `json:"createdAt"`
}

// MandateResponse is returned by Service.Create and Service.Get.
type MandateResponse struct {
	Mandate
	Success bool    `json:"success"`
	Message string  `json:"message,omitempty"`
	Data    Mandate `json:"data"`
}

// MandateListResponse is returned by Service.List.
type MandateListResponse struct {
	Results []Mandate `json:"results"`
	Total   int       `json:"total"`
	PerPage string    `json:"perPage"`
	Page    int       `json:"page"`
}

// DebitRequest initiates a debit under an approved mandate.
type DebitRequest struct {
	Amount               string `json:"amount"`
	Description          string `json:"description"`
	MandateReference     string `json:"mandateReference"`
	Reference            string `json:"reference"`
	Currency             string `json:"currency"`
	BankCode             string `json:"bankCode"`
	AccountNumber        string `json:"accountNumber"`
	InitiatorAccountName string `json:"initiatorAccountName"`
	BeneficiaryNarration string `json:"beneficiaryNarration"`
}

// Debit describes a direct-debit transaction under a mandate.
type Debit struct {
	Reference        string  `json:"reference"`
	MandateReference string  `json:"mandateReference"`
	Amount           float64 `json:"amount"`
	Currency         string  `json:"currency"`
	Status           string  `json:"status"`
	Message          string  `json:"message"`
}

// DebitResponse is returned by Service.Debit and Service.VerifyDebit.
type DebitResponse struct {
	Success bool   `json:"success"`
	Status  bool   `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Data    Debit  `json:"data"`
}

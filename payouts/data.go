// Package payouts exposes Fincra disbursement, payout-status, document-upload,
// and Fincra-to-Fincra transfer endpoints.
package payouts

import (
	"io"
	"time"
)

// Address is a postal address used by beneficiaries and bank records.
type Address struct {
	Country    string `json:"country,omitempty"`
	State      string `json:"state,omitempty"`
	Zip        string `json:"zip,omitempty"`
	City       string `json:"city,omitempty"`
	Street     string `json:"street,omitempty"`
	PostalCode string `json:"postalCode,omitempty"`
}

// Beneficiary is the recipient of a payout.
type Beneficiary struct {
	FirstName            string   `json:"firstName,omitempty"`
	LastName             string   `json:"lastName,omitempty"`
	MiddleName           string   `json:"middleName,omitempty"`
	AccountHolderName    string   `json:"accountHolderName,omitempty"`
	Email                string   `json:"email,omitempty"`
	Phone                string   `json:"phone,omitempty"`
	AccountNumber        string   `json:"accountNumber,omitempty"`
	WalletAddress        string   `json:"walletAddress,omitempty"`
	DestinationTag       string   `json:"destinationTag,omitempty"`
	Type                 string   `json:"type,omitempty"`
	Country              string   `json:"country,omitempty"`
	BankCode             string   `json:"bankCode,omitempty"`
	BankName             string   `json:"bankName,omitempty"`
	SortCode             string   `json:"sortCode,omitempty"`
	BankSwiftCode        string   `json:"bankSwiftCode,omitempty"`
	MobileMoneyCode      string   `json:"mobileMoneyCode,omitempty"`
	InteracEmail         string   `json:"interacEmail,omitempty"`
	SecurityQuestion     string   `json:"securityQuestion,omitempty"`
	SecurityAnswer       string   `json:"securityAnswer,omitempty"`
	BirthDate            string   `json:"birthDate,omitempty"`
	BirthPlace           string   `json:"birthPlace,omitempty"`
	RegistrationNumber   string   `json:"registrationNumber,omitempty"`
	IncorporationCountry string   `json:"incorporationCountry,omitempty"`
	Address              *Address `json:"address,omitempty"`
	BankAddress          *Address `json:"bankAddress,omitempty"`
}

// PayoutRequest creates a payout. Amount is expressed in the source currency.
type PayoutRequest struct {
	Business            string         `json:"business,omitempty"`
	SourceCurrency      string         `json:"sourceCurrency"`
	DestinationCurrency string         `json:"destinationCurrency"`
	Amount              float64        `json:"amount"`
	QuoteReference      string         `json:"quoteReference,omitempty"`
	Description         string         `json:"description,omitempty"`
	PaymentDestination  string         `json:"paymentDestination"`
	PaymentScheme       string         `json:"paymentScheme,omitempty"`
	CustomerReference   string         `json:"customerReference"`
	Beneficiary         Beneficiary    `json:"beneficiary"`
	AdditionalInfo      map[string]any `json:"additionalInfo,omitempty"`
}

// PayoutListParams filters Service.ListPayouts.
type PayoutListParams struct {
	BusinessID          string
	SourceCurrency      string
	DestinationCurrency string
	SubAccount          string
	Status              []string
	Page                int
	PerPage             int
	DateFrom            string
	DateTo              string
}

// PayoutSummary is one row in a payout list.
type PayoutSummary struct {
	ID                  int64     `json:"id"`
	ObjectID            string    `json:"_id"`
	AmountSent          float64   `json:"amountSent"`
	AmountReceived      float64   `json:"amountReceived"`
	SourceCurrency      string    `json:"sourceCurrency"`
	DestinationCurrency string    `json:"destinationCurrency"`
	Fee                 float64   `json:"fee"`
	BeneficiaryName     string    `json:"beneficiaryName"`
	Status              string    `json:"status"`
	Reference           string    `json:"reference"`
	BusinessName        string    `json:"businessName"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// PayoutListData is the data payload of PayoutListResponse.
type PayoutListData struct {
	Results []PayoutSummary `json:"results"`
	Total   int64           `json:"total"`
}

// PayoutListResponse is returned by Service.ListPayouts.
type PayoutListResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    PayoutListData `json:"data"`
}

// Recipient is the resolved payout beneficiary.
type Recipient struct {
	FirstName            string   `json:"firstName"`
	LastName             string   `json:"lastName"`
	MiddleName           string   `json:"middleName"`
	AccountHolderName    string   `json:"accountHolderName"`
	Email                string   `json:"email"`
	Phone                string   `json:"phone"`
	AccountNumber        string   `json:"accountNumber"`
	WalletAddress        string   `json:"walletAddress"`
	DestinationTag       string   `json:"destinationTag"`
	Type                 string   `json:"type"`
	Country              string   `json:"country"`
	BankCode             string   `json:"bankCode"`
	BankName             string   `json:"bankName"`
	SortCode             string   `json:"sortCode"`
	BankSwiftCode        string   `json:"bankSwiftCode"`
	MobileMoneyCode      string   `json:"mobileMoneyCode"`
	RegistrationNumber   string   `json:"registrationNumber"`
	IncorporationCountry string   `json:"incorporationCountry"`
	BirthDate            string   `json:"birthDate"`
	BirthPlace           string   `json:"birthPlace"`
	BankAddress          *Address `json:"bankAddress"`
}

// Payout is a single payout with full detail.
type Payout struct {
	ID                  int64     `json:"id"`
	AmountSent          float64   `json:"amountSent"`
	AmountReceived      float64   `json:"amountReceived"`
	SourceCurrency      string    `json:"sourceCurrency"`
	DestinationCurrency string    `json:"destinationCurrency"`
	Fee                 float64   `json:"fee"`
	QuoteID             string    `json:"quoteId"`
	Status              string    `json:"status"`
	Reference           string    `json:"reference"`
	CustomerReference   string    `json:"customerReference"`
	Description         string    `json:"description"`
	PaymentDestination  string    `json:"paymentDestination"`
	PaymentScheme       string    `json:"paymentScheme"`
	AdditionalInfo      any       `json:"additionalInfo"`
	TraceID             string    `json:"traceId"`
	Message             string    `json:"message"`
	IsDocumentRequired  int       `json:"isDocumentRequired"`
	IsDocumentUploaded  int       `json:"isDocumentUploaded"`
	DocumentsRequired   []string  `json:"documentsRequired"`
	ValuedAt            time.Time `json:"valuedAt"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
	Recipient           Recipient `json:"recipient"`
}

// PayoutResponse is returned by Service.CreatePayout and the status lookups.
type PayoutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    Payout `json:"data"`
}

// CreatedPayout is the compact result of creating a payout.
type CreatedPayout struct {
	ID                 int64    `json:"id"`
	Reference          string   `json:"reference"`
	CustomerReference  string   `json:"customerReference"`
	Status             string   `json:"status"`
	Message            string   `json:"message"`
	IsDocumentRequired bool     `json:"isDocumentRequired"`
	DocumentsRequired  []string `json:"documentsRequired"`
}

// CreatePayoutResponse is returned by Service.CreatePayout.
type CreatePayoutResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Data    CreatedPayout `json:"data"`
}

// DocumentUploadRequest uploads supporting documents for a payout.
type DocumentUploadRequest struct {
	Reference string
	Name      string
	Type      string
	FileName  string
	File      io.Reader
}

// DocumentUploadResponse is returned by Service.UploadDocument.
type DocumentUploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// WalletTransferRequest moves funds to another Fincra wallet.
type WalletTransferRequest struct {
	Amount                  string `json:"amount"`
	Business                string `json:"business,omitempty"`
	CustomerReference       string `json:"customerReference"`
	Description             string `json:"description"`
	BeneficiaryWalletNumber string `json:"beneficiaryWalletNumber"`
}

// WalletTransfer is the result of a Fincra-to-Fincra transfer.
type WalletTransfer struct {
	ID                int64  `json:"id"`
	Reference         string `json:"reference"`
	CustomerReference string `json:"customerReference"`
	Status            string `json:"status"`
}

// WalletTransferResponse is returned by Service.TransferToFincraAccount.
type WalletTransferResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    WalletTransfer `json:"data"`
}

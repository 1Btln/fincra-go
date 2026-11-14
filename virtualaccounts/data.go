// Package virtualaccounts exposes Fincra virtual-account endpoints.
package virtualaccounts

// Request creates a permanent virtual account. KYC fields vary by currency
// and account type, so KYCInformation and document values remain flexible JSON.
type Request struct {
	Currency                       string         `json:"currency"`
	AccountType                    string         `json:"accountType"`
	KYCInformation                 map[string]any `json:"KYCInformation"`
	MerchantReference              string         `json:"merchantReference,omitempty"`
	PhoneNumber                    string         `json:"phoneNumber,omitempty"`
	Reason                         string         `json:"reason,omitempty"`
	Purpose                        string         `json:"purpose,omitempty"`
	EntityName                     string         `json:"entityName,omitempty"`
	PaymentFlowDescription         string         `json:"paymentFlowDescription,omitempty"`
	MeansOfID                      any            `json:"meansOfId,omitempty"`
	UtilityBill                    any            `json:"utilityBill,omitempty"`
	BankStatement                  any            `json:"bankStatement,omitempty"`
	LeaseAgreement                 any            `json:"leaseAgreement,omitempty"`
	SourceOfFunds                  any            `json:"sourceOfFunds,omitempty"`
	BeneficialOwnershipCertificate any            `json:"beneficialOwnershipCertificate,omitempty"`
	ArticleOfIncorporation         any            `json:"articleOfIncorporation,omitempty"`
	IncorporationDocuments         any            `json:"incorporationDocuments,omitempty"`
	AccountAgreement               any            `json:"accountAgreement,omitempty"`
	RegulatoryEvidence             any            `json:"regulatoryEvidence,omitempty"`
	RequestedCurrencies            []string       `json:"requestedCurrencies,omitempty"`
	Metadata                       map[string]any `json:"metadata,omitempty"`
	AccountOpeningDate             string         `json:"accountOpeningDate,omitempty"`
	IsTermsAccepted                bool           `json:"isTermsAccepted,omitempty"`
	TermsAndConditionAcceptance    map[string]any `json:"termsAndConditionAcceptance,omitempty"`
}

// Account contains issued virtual-account information.
type Account struct {
	ID                 string              `json:"_id"`
	Status             string              `json:"status"`
	IsActive           bool                `json:"isActive"`
	AccountNumber      string              `json:"accountNumber"`
	MerchantReference  string              `json:"merchantReference"`
	Business           string              `json:"business"`
	Currency           string              `json:"currency"`
	AccountType        string              `json:"accountType"`
	AccountName        string              `json:"accountName"`
	BankName           string              `json:"bankName"`
	BankCode           string              `json:"bankCode"`
	AccountInformation *AccountInformation `json:"accountInformation"`
	KYCInformation     map[string]any      `json:"KYCInformation"`
	CreatedAt          string              `json:"createdAt"`
	UpdatedAt          string              `json:"updatedAt"`
}

// AccountInformation contains issued bank-account details. Some currencies,
// including CAD, expose their receiving address in OtherInfo instead of an
// account number.
type AccountInformation struct {
	AccountNumber string            `json:"accountNumber"`
	AccountName   string            `json:"accountName"`
	BankName      string            `json:"bankName"`
	BankCode      string            `json:"bankCode"`
	CountryCode   string            `json:"countryCode"`
	Email         string            `json:"email"`
	BankAddress   string            `json:"bankAddress"`
	OtherInfo     *AccountOtherInfo `json:"otherInfo"`
}

// AccountOtherInfo contains country-specific account details.
type AccountOtherInfo struct {
	IBAN          string `json:"iban"`
	AccountNumber string `json:"accountNumber"`
	CheckNumber   string `json:"checkNumber"`
	SortCode      string `json:"sortCode"`
	BankSwiftCode string `json:"bankSwiftCode"`
	AddressableIn string `json:"addressableIn"`
	Memo          string `json:"memo"`
	InteracEmail  string `json:"interacEmail"`
}

// AccountResponse is returned by Service.Request and Service.Get.
type AccountResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	Data    Account `json:"data"`
}

// AccountListData is the data payload of AccountListResponse.
type AccountListData struct {
	Results []Account `json:"results"`
	Total   int       `json:"total"`
}

// AccountListResponse is returned by Service.List.
type AccountListResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    AccountListData `json:"data"`
}

// RequestOption describes whether an account currency is available.
type RequestOption struct {
	Currency  string `json:"currency"`
	Available bool   `json:"available"`
	Requested bool   `json:"requested"`
}

// RequestOptionsResponse is returned by Service.RequestOptions.
type RequestOptionsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// TemporaryRequest creates a temporary virtual account.
type TemporaryRequest struct {
	Amount            string `json:"amount"`
	ExpiresAt         string `json:"expiresAt"`
	MerchantReference string `json:"merchantReference,omitempty"`
	Channel           string `json:"channel,omitempty"`
}

// TemporaryAccountData contains the payment details for a temporary account.
type TemporaryAccountData struct {
	AccountNumber string         `json:"accountNumber"`
	AccountName   string         `json:"accountName"`
	BankName      string         `json:"bankName"`
	Reference     string         `json:"reference"`
	ExpiresAt     string         `json:"expiresAt"`
	Extra         map[string]any `json:"-"`
}

// TemporaryAccountResponse is returned by Service.RequestTemporary.
type TemporaryAccountResponse struct {
	Success bool                 `json:"success"`
	Status  bool                 `json:"status,omitempty"`
	Message string               `json:"message"`
	Data    TemporaryAccountData `json:"data"`
}

// BVNLookupParams identifies a business and BVN for Service.GetByBVN.
// BusinessID defaults to the ID configured on the client.
type BVNLookupParams struct {
	BVN        string
	BusinessID string
}

// BVNAccountInformation contains the bank details returned by the BVN lookup.
type BVNAccountInformation = AccountInformation

// BVNAccount is a virtual account record returned by the BVN lookup endpoint.
type BVNAccount struct {
	ID                         string                  `json:"_id"`
	Business                   string                  `json:"business"`
	Status                     string                  `json:"status"`
	IsActive                   bool                    `json:"isActive"`
	AccountNumber              string                  `json:"accountNumber"`
	MerchantReference          string                  `json:"merchantReference"`
	Currency                   string                  `json:"currency"`
	AccountType                string                  `json:"accountType"`
	AccountInformation         BVNAccountInformation   `json:"accountInformation"`
	PreviousAccountInformation []BVNAccountInformation `json:"previousAccountInformation"`
	KYCInformation             map[string]any          `json:"KYCInformation"`
	Metadata                   map[string]any          `json:"metadata"`
	IsPermanent                bool                    `json:"isPermanent"`
	ExpiresAt                  *string                 `json:"expiresAt"`
	Reason                     *string                 `json:"reason"`
	CreatedAt                  string                  `json:"createdAt"`
	UpdatedAt                  string                  `json:"updatedAt"`
}

// BVNLookupResponse is returned by Service.GetByBVN.
type BVNLookupResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Data    []BVNAccount `json:"data"`
}

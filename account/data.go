// Package account exposes Fincra business, merchant, bank, account-resolution,
// and BVN verification endpoints.
package account

import "time"

// Settings are a business's payment settings.
type Settings struct {
	FeeBearer             string `json:"feeBearer"`
	SettlementDestination string `json:"settlementDestination"`
	EnableWebhook         bool   `json:"enableWebhook"`
	CallbackURL           string `json:"callbackURL"`
}

// EmailSettings control business notification emails.
type EmailSettings struct {
	CollectionNotification     bool `json:"collectionNotification"`
	TopupNotification          bool `json:"topupNotification"`
	SettlementNotification     bool `json:"settlementNotification"`
	PaymentReceiptNotification bool `json:"paymentReceiptNotification"`
}

// CheckoutSettings control the hosted checkout appearance.
type CheckoutSettings struct {
	UseBrandLogo        bool   `json:"useBrandLogo"`
	ButtonTextColor     string `json:"buttonTextColor"`
	ButtonColor         string `json:"buttonColor"`
	FontColor           string `json:"fontColor"`
	FontFamily          string `json:"fontFamily"`
	BackgroundColor     string `json:"backgroundColor"`
	BackgroundTextColor string `json:"backgroundTextColor"`
}

// Business is the business profile returned by GetBusiness.
type Business struct {
	ID                 string           `json:"_id"`
	Name               string           `json:"name"`
	BusinessType       string           `json:"businessType"`
	Website            string           `json:"website"`
	Email              string           `json:"email"`
	Country            string           `json:"country"`
	Mobile             string           `json:"mobile"`
	BusinessKey        string           `json:"businessKey"`
	Status             string           `json:"status"`
	CountryCallingCode string           `json:"countryCallingCode"`
	BusinessTag        int64            `json:"businessTag"`
	IsKYCApproved      bool             `json:"isKYCApproved"`
	IsBlocked          bool             `json:"isBlocked"`
	Settings           Settings         `json:"settings"`
	EmailSettings      EmailSettings    `json:"emailSettings"`
	CheckoutSettings   CheckoutSettings `json:"checkoutSettings"`
	CreatedAt          time.Time        `json:"createdAt"`
	UpdatedAt          time.Time        `json:"updatedAt"`
}

// BusinessResponse is returned by Service.GetBusiness.
type BusinessResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Data    Business `json:"data"`
}

// MerchantBusiness identifies the business attached to a merchant.
type MerchantBusiness struct {
	ID          string `json:"id"`
	BusinessKey string `json:"businessKey"`
}

// Merchant is the merchant profile returned by GetMerchant.
type Merchant struct {
	ID                   string           `json:"_id"`
	UserID               string           `json:"userId"`
	Name                 string           `json:"name"`
	Email                string           `json:"email"`
	Country              string           `json:"country"`
	Logo                 string           `json:"logo"`
	AccessAllFeatures    bool             `json:"accessAllFeatures"`
	WhitelistedIPs       []string         `json:"whitelistedIpAddresses"`
	HowDidYouHearAboutUs []map[string]any `json:"howDidYouHearAboutUs"`
	Business             MerchantBusiness `json:"business"`
	CreatedAt            time.Time        `json:"createdAt"`
	UpdatedAt            time.Time        `json:"updatedAt"`
}

// MerchantResponse is returned by Service.GetMerchant.
type MerchantResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Data    Merchant `json:"data"`
}

// Bank is a bank or payout provider accepted by Fincra.
type Bank struct {
	ID                 string `json:"id"`
	ObjectID           string `json:"_id"`
	Name               string `json:"name"`
	Code               string `json:"code"`
	Currency           string `json:"currency"`
	Country            string `json:"country"`
	Type               string `json:"type"`
	PaymentDestination string `json:"paymentDestination"`
	SortCode           string `json:"sortCode"`
	SwiftCode          string `json:"swiftCode"`
	Active             bool   `json:"active"`
}

// BanksResponse is returned by Service.ListBanks.
type BanksResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    []Bank `json:"data"`
}

// BankParams filters Service.ListBanks. Currency or Country is required.
type BankParams struct {
	Currency           string
	Country            string
	PaymentDestination string
}

// ResolveAccountRequest verifies an account before a payout. Type selects which
// identifier fields Fincra expects.
type ResolveAccountRequest struct {
	Type            string `json:"type"`
	AccountNumber   string `json:"accountNumber,omitempty"`
	BankCode        string `json:"bankCode,omitempty"`
	BankSwiftCode   string `json:"bankSwiftCode,omitempty"`
	MobileMoneyCode string `json:"mobileMoneyCode,omitempty"`
	IBAN            string `json:"iban,omitempty"`
	Currency        string `json:"currency,omitempty"`
	InteracEmail    string `json:"interacEmail,omitempty"`
}

// ResolvedAccount is the result of account resolution.
type ResolvedAccount struct {
	AccountNumber      string `json:"accountNumber"`
	AccountName        string `json:"accountName"`
	Email              string `json:"email"`
	BankCode           string `json:"bankCode"`
	BankName           string `json:"bankName"`
	AutoDepositEnabled bool   `json:"autoDepositEnabled"`
}

// ResolveAccountResponse is returned by Service.ResolveAccount.
type ResolveAccountResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    ResolvedAccount `json:"data"`
}

// BVNVerificationRequest verifies a Bank Verification Number.
type BVNVerificationRequest struct {
	BVN      string `json:"bvn"`
	Business string `json:"business,omitempty"`
}

// BVNData is the verification result.
type BVNData struct {
	VerificationStatus string      `json:"verificationStatus"`
	Response           BVNIdentity `json:"response"`
}

// BVNIdentity is the identity record returned by a BVN lookup.
type BVNIdentity struct {
	FirstName   string `json:"firstName"`
	MiddleName  string `json:"middleName"`
	LastName    string `json:"lastName"`
	Gender      string `json:"gender"`
	DateOfBirth string `json:"dateOfBirth"`
	PhoneNumber string `json:"phoneNo"`
	PhotoBase64 string `json:"pixBase64"`
}

// BVNVerificationResponse is returned by Service.VerifyBVN.
type BVNVerificationResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	Data    BVNData `json:"data"`
}

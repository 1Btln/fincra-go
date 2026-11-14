// Package collections exposes Fincra deposit, refund, and chargeback endpoints.
package collections

import "time"

// DepositParams filters Service.ListDeposits. VirtualAccount is the virtual
// account ID or number advertised by Fincra.
type DepositParams struct {
	BusinessID     string
	VirtualAccount string
}

// Deposit is a successful pay-in.
type Deposit struct {
	ID                int64     `json:"id"`
	ObjectID          string    `json:"_id"`
	Reference         string    `json:"reference"`
	MerchantReference string    `json:"merchantReference"`
	VirtualAccount    string    `json:"virtualAccount"`
	VirtualAccountID  string    `json:"virtualAccountId"`
	Amount            float64   `json:"amount"`
	Currency          string    `json:"currency"`
	Status            string    `json:"status"`
	PaymentMethod     string    `json:"paymentMethod"`
	CustomerName      string    `json:"customerName"`
	CustomerEmail     string    `json:"customerEmail"`
	Description       string    `json:"description"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// DepositData is the data payload of DepositListResponse.
type DepositData struct {
	Results []Deposit `json:"results"`
	Total   int64     `json:"total"`
}

// DepositListResponse is returned by Service.ListDeposits.
type DepositListResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    DepositData `json:"data"`
}

// DepositResponse is returned by Service.VerifyDeposit.
type DepositResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	Data    Deposit `json:"data"`
}

// Chargeback is a disputed card collection.
type Chargeback struct {
	ID                   int64     `json:"id"`
	Reference            string    `json:"reference"`
	CollectionReference  string    `json:"collectionReference"`
	Amount               float64   `json:"amount"`
	Currency             string    `json:"currency"`
	Status               string    `json:"status"`
	Reason               string    `json:"reason"`
	BusinessRejectReason string    `json:"businessRejectReason"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

// ChargebackData is the data payload of ChargebackListResponse.
type ChargebackData struct {
	Results []Chargeback `json:"results"`
	Total   int64        `json:"total"`
}

// ChargebackListResponse is returned by Service.ListChargebacks.
type ChargebackListResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    ChargebackData `json:"data"`
}

// RejectChargebackRequest carries the reason for rejecting a chargeback.
type RejectChargebackRequest struct {
	BusinessRejectReason string `json:"business_reject_reason"`
}

// ActionResponse is returned by chargeback decision endpoints.
type ActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// InitiateRefundRequest starts a refund for a collection.
type InitiateRefundRequest struct {
	TransReference string `json:"transReference"`
}

// RefundPayee describes where a refund is sent.
type RefundPayee struct {
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	AccountNumber string `json:"accountNumber"`
	BankCode      string `json:"bankCode"`
	BankName      string `json:"bankName"`
	CardScheme    string `json:"cardScheme"`
}

// Refund is a collection refund.
type Refund struct {
	ID                            int64          `json:"id"`
	CollectionReference           string         `json:"collectionReference"`
	Reference                     string         `json:"reference"`
	Currency                      string         `json:"currency"`
	AmountRefunded                string         `json:"amountRefunded"`
	RefundType                    string         `json:"refundType"`
	PayeeAccountNumber            string         `json:"payeeAccountNumber"`
	CollectionExpectedAmount      string         `json:"collectionExpectedAmount"`
	CollectionAmountReceived      string         `json:"collectionAmountReceived"`
	CollectionAmountSettled       string         `json:"collectionAmountSettled"`
	CollectionSettlementReference string         `json:"collectionSettlementReference"`
	CollectionMerchantReference   string         `json:"collectionMerchantReference"`
	CollectionVarianceType        string         `json:"collectionVarianceType"`
	CollectionVirtualAccountID    string         `json:"collectionVirtualAccountId"`
	Metadata                      map[string]any `json:"metadata"`
	RefundStatus                  string         `json:"refundStatus"`
	CompletedAt                   string         `json:"completedAt"`
	InitiatedAt                   string         `json:"initiatedAt"`
	Payee                         RefundPayee    `json:"payee"`
}

// RefundResponse is returned by Service.InitiateRefund and Service.GetRefund.
type RefundResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    Refund `json:"data"`
}

// RefundListData is the data payload of RefundListResponse.
type RefundListData struct {
	Results []Refund `json:"results"`
	Total   int64    `json:"total"`
}

// RefundListResponse is returned by Service.ListRefunds.
type RefundListResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    RefundListData `json:"data"`
}

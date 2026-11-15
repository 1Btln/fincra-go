package payouts

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/1Btln/fincra-go/internal/api"
)

// Service groups Fincra payout and disbursement endpoints. Access it through
// Client.Payouts.
type Service struct {
	backend api.Backend
}

// New returns a payouts service backed by backend.
func New(backend api.Backend) *Service {
	return &Service{backend: backend}
}

// ListPayouts returns payouts for a business. Fincra documents this operation
// with a JSON body on a GET; this SDK sends the filters as query parameters.
func (s *Service) ListPayouts(ctx context.Context, params *PayoutListParams) (*PayoutListResponse, error) {
	if params == nil {
		params = &PayoutListParams{}
	}
	businessID, err := api.BusinessID(params.BusinessID, s.backend.BusinessID())
	if err != nil {
		return nil, err
	}
	if err := api.ValidateDate("dateFrom", params.DateFrom); err != nil {
		return nil, err
	}
	if err := api.ValidateDate("dateTo", params.DateTo); err != nil {
		return nil, err
	}

	query := url.Values{"business": {businessID}}
	api.SetString(query, "sourceCurrency", strings.ToUpper(params.SourceCurrency))
	api.SetString(query, "destinationCurrency", strings.ToUpper(params.DestinationCurrency))
	api.SetString(query, "subAccount", params.SubAccount)
	for _, status := range params.Status {
		if status = strings.TrimSpace(status); status != "" {
			query.Add("status", status)
		}
	}
	api.SetInt(query, "page", params.Page)
	api.SetInt(query, "perPage", params.PerPage)
	api.SetString(query, "dateFrom", params.DateFrom)
	api.SetString(query, "dateTo", params.DateTo)

	var out PayoutListResponse
	if err := s.backend.Get(ctx, "/disbursements/payouts?"+query.Encode(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreatePayout initiates a bank-account or mobile-money payout. A quote
// reference is required when the source and destination currencies differ.
func (s *Service) CreatePayout(ctx context.Context, req *PayoutRequest) (*CreatePayoutResponse, error) {
	if req == nil {
		return nil, api.Missing("request")
	}
	businessID, err := api.BusinessID(req.Business, s.backend.BusinessID())
	if err != nil {
		return nil, err
	}
	if err := api.Require("sourceCurrency", req.SourceCurrency); err != nil {
		return nil, err
	}
	if err := api.Require("destinationCurrency", req.DestinationCurrency); err != nil {
		return nil, err
	}
	if req.Amount <= 0 {
		return nil, api.Invalid("amount", "must be greater than zero")
	}
	if err := api.Require("customerReference", req.CustomerReference); err != nil {
		return nil, err
	}
	if err := api.Require("paymentDestination", req.PaymentDestination); err != nil {
		return nil, err
	}
	source := strings.ToUpper(strings.TrimSpace(req.SourceCurrency))
	destination := strings.ToUpper(strings.TrimSpace(req.DestinationCurrency))
	if source != destination && strings.TrimSpace(req.QuoteReference) == "" {
		return nil, api.Missing("quoteReference (required for cross-currency payouts)")
	}

	payload := *req
	payload.Business = businessID
	payload.SourceCurrency = source
	payload.DestinationCurrency = destination

	var out CreatePayoutResponse
	if err := s.backend.Post(ctx, "/disbursements/payouts", &payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPayoutByReference returns a payout by Fincra's transaction reference.
func (s *Service) GetPayoutByReference(ctx context.Context, transactionReference string) (*PayoutResponse, error) {
	transactionReference = strings.TrimSpace(transactionReference)
	if err := api.Require("transactionReference", transactionReference); err != nil {
		return nil, err
	}

	var out PayoutResponse
	path := "/disbursements/payouts/reference/" + url.PathEscape(transactionReference)
	if err := s.backend.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPayoutByCustomerReference returns a payout by the caller's reference.
func (s *Service) GetPayoutByCustomerReference(ctx context.Context, customerReference string) (*PayoutResponse, error) {
	customerReference = strings.TrimSpace(customerReference)
	if err := api.Require("customerReference", customerReference); err != nil {
		return nil, err
	}

	var out PayoutResponse
	path := "/disbursements/payouts/customer-reference/" + url.PathEscape(customerReference)
	if err := s.backend.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UploadDocument uploads a supporting document for a payout as multipart form
// data.
func (s *Service) UploadDocument(ctx context.Context, req *DocumentUploadRequest) (*DocumentUploadResponse, error) {
	if req == nil {
		return nil, api.Missing("request")
	}
	reference := strings.TrimSpace(req.Reference)
	if err := api.Require("reference", reference); err != nil {
		return nil, err
	}
	if req.File == nil {
		return nil, api.Missing("file")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("reference", reference); err != nil {
		return nil, err
	}
	if name := strings.TrimSpace(req.Name); name != "" {
		if err := writer.WriteField("name", name); err != nil {
			return nil, err
		}
	}
	if docType := strings.TrimSpace(req.Type); docType != "" {
		if err := writer.WriteField("type", docType); err != nil {
			return nil, err
		}
	}
	fileName := strings.TrimSpace(req.FileName)
	if fileName == "" {
		fileName = "document"
	}
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, req.File); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	contentType := writer.FormDataContentType()
	var out DocumentUploadResponse
	err = s.backend.Post(ctx, "/disbursements/payouts/documents-upload", &body, &out,
		func(r *http.Request) { r.Header.Set("Content-Type", contentType) })
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// TransferToFincraAccount moves funds to another Fincra wallet.
func (s *Service) TransferToFincraAccount(ctx context.Context, req *WalletTransferRequest) (*WalletTransferResponse, error) {
	if req == nil {
		return nil, api.Missing("request")
	}
	businessID, err := api.BusinessID(req.Business, s.backend.BusinessID())
	if err != nil {
		return nil, err
	}
	if err := api.Require("amount", req.Amount); err != nil {
		return nil, err
	}
	if err := api.Require("customerReference", req.CustomerReference); err != nil {
		return nil, err
	}
	if err := api.Require("description", req.Description); err != nil {
		return nil, err
	}
	if err := api.Require("beneficiaryWalletNumber", req.BeneficiaryWalletNumber); err != nil {
		return nil, err
	}

	payload := *req
	payload.Business = businessID

	var out WalletTransferResponse
	if err := s.backend.Post(ctx, "/disbursements/payouts/wallets", &payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

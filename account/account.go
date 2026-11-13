package account

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/Ledgenta/fincra-go/internal/api"
)

// Backend is the subset of the Fincra client this package needs.
type Backend interface {
	Do(ctx context.Context, method, path string, body, out any, opts ...func(*http.Request)) error
	BusinessID() string
}

// Service groups Fincra account-management endpoints. Access it through
// Client.Account.
type Service struct {
	backend Backend
}

// New returns an account service backed by backend.
func New(backend Backend) *Service {
	return &Service{backend: backend}
}

// GetBusiness returns the business profile attached to the API key.
func (s *Service) GetBusiness(ctx context.Context) (*BusinessResponse, error) {
	var out BusinessResponse
	if err := s.backend.Do(ctx, http.MethodGet, "/profile/business/me", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetMerchant returns the merchant profile attached to the API key.
func (s *Service) GetMerchant(ctx context.Context) (*MerchantResponse, error) {
	var out MerchantResponse
	if err := s.backend.Do(ctx, http.MethodGet, "/profile/merchants/me", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListBanks returns banks and payout providers. At least one of Currency or
// Country must be supplied.
func (s *Service) ListBanks(ctx context.Context, params *BankParams) (*BanksResponse, error) {
	if params == nil {
		params = &BankParams{}
	}
	currency := strings.ToUpper(strings.TrimSpace(params.Currency))
	country := strings.ToUpper(strings.TrimSpace(params.Country))
	if err := api.RequireAny("currency or country", currency, country); err != nil {
		return nil, err
	}

	query := url.Values{}
	api.SetString(query, "currency", currency)
	api.SetString(query, "country", country)
	api.SetString(query, "paymentDestination", params.PaymentDestination)

	var out BanksResponse
	if err := s.backend.Do(ctx, http.MethodGet, "/core/banks?"+query.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ResolveAccount verifies an account number, IBAN, or mobile-money wallet.
func (s *Service) ResolveAccount(ctx context.Context, req *ResolveAccountRequest) (*ResolveAccountResponse, error) {
	if req == nil {
		return nil, api.Missing("request")
	}
	if err := api.Require("type", req.Type); err != nil {
		return nil, err
	}
	if err := api.RequireAny("accountNumber, iban, or interacEmail", req.AccountNumber, req.IBAN, req.InteracEmail); err != nil {
		return nil, err
	}

	var out ResolveAccountResponse
	if err := s.backend.Do(ctx, http.MethodPost, "/core/accounts/resolve", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// VerifyBVN resolves a Bank Verification Number. When Business is empty the
// client's configured BusinessID is used.
func (s *Service) VerifyBVN(ctx context.Context, req *BVNVerificationRequest) (*BVNVerificationResponse, error) {
	if req == nil {
		return nil, api.Missing("request")
	}
	if err := api.Require("bvn", req.BVN); err != nil {
		return nil, err
	}
	businessID, err := api.BusinessID(req.Business, s.backend.BusinessID())
	if err != nil {
		return nil, err
	}

	payload := *req
	payload.Business = businessID

	var out BVNVerificationResponse
	if err := s.backend.Do(ctx, http.MethodPost, "/core/bvn-verification", &payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

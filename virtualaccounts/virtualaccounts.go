package virtualaccounts

import (
	"context"
	"net/url"
	"strings"

	"github.com/1Btln/fincra-go/internal/api"
)

// backend is the subset of the Fincra client this package needs.
type backend interface {
	api.HTTPClient
	BusinessID() string
}

// Service groups Fincra permanent and temporary virtual-account endpoints.
type Service struct{ backend backend }

// New returns a virtual-account service backed by backend.
func New(backend backend) *Service { return &Service{backend: backend} }

// Request creates a permanent virtual-account request.
func (s *Service) Request(ctx context.Context, req *Request) (*AccountResponse, error) {
	if req == nil {
		return nil, api.Missing("request")
	}
	if err := api.Require("currency", req.Currency); err != nil {
		return nil, err
	}
	if err := api.Require("accountType", req.AccountType); err != nil {
		return nil, err
	}
	if req.KYCInformation == nil {
		return nil, api.Missing("KYCInformation")
	}
	var out AccountResponse
	if err := s.backend.Post(ctx, "/profile/virtual-accounts/requests", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RequestOptions returns the supported account-request currencies and their availability.
func (s *Service) RequestOptions(ctx context.Context) (*RequestOptionsResponse, error) {
	var out RequestOptionsResponse
	if err := s.backend.Get(ctx, "/profile/virtual-accounts/requests", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RequestTemporary creates a short-lived NGN virtual account for one payment.
func (s *Service) RequestTemporary(ctx context.Context, req *TemporaryRequest) (*TemporaryAccountResponse, error) {
	if req == nil {
		return nil, api.Missing("request")
	}
	if err := api.Require("amount", req.Amount); err != nil {
		return nil, err
	}
	if err := api.Require("expiresAt", req.ExpiresAt); err != nil {
		return nil, err
	}
	var out TemporaryAccountResponse
	if err := s.backend.Post(ctx, "/profile/virtual-accounts/transfer", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a virtual account by its Fincra ID.
func (s *Service) Get(ctx context.Context, accountID string) (*AccountResponse, error) {
	accountID = strings.TrimSpace(accountID)
	if err := api.Require("accountID", accountID); err != nil {
		return nil, err
	}
	var out AccountResponse
	path := "/profile/virtual-accounts/" + url.PathEscape(accountID)
	if err := s.backend.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns virtual accounts, optionally filtered by currency.
func (s *Service) List(ctx context.Context, currency string) (*AccountListResponse, error) {
	query := url.Values{}
	api.SetString(query, "currency", strings.ToLower(currency))
	path := "/profile/virtual-accounts/"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	var out AccountListResponse
	if err := s.backend.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetByBVN retrieves virtual-account records associated with a BVN.
func (s *Service) GetByBVN(ctx context.Context, params *BVNLookupParams) (*BVNLookupResponse, error) {
	if params == nil {
		return nil, api.Missing("request")
	}
	bvn := strings.TrimSpace(params.BVN)
	if err := api.Require("bvn", bvn); err != nil {
		return nil, err
	}
	businessID, err := api.BusinessID(params.BusinessID, s.backend.BusinessID())
	if err != nil {
		return nil, err
	}
	query := url.Values{"bvn": {bvn}, "business": {businessID}}
	var out BVNLookupResponse
	if err := s.backend.Get(ctx, "/profile/virtual-accounts/bvn?"+query.Encode(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

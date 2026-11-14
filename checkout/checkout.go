package checkout

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

// Service groups hosted-checkout and fee-calculation endpoints.
type Service struct{ backend backend }

// New returns a checkout service backed by backend.
func New(backend backend) *Service { return &Service{backend: backend} }

// Initiate creates a hosted checkout payment link.
func (s *Service) Initiate(ctx context.Context, req *InitiateRequest) (*PaymentResponse, error) {
	if req == nil {
		return nil, api.Missing("request")
	}
	if err := api.Require("currency", req.Currency); err != nil {
		return nil, err
	}
	if req.Amount <= 0 {
		return nil, api.Invalid("amount", "must be greater than zero")
	}
	if req.Customer == nil {
		return nil, api.Missing("customer")
	}
	if err := api.Require("customer.name", req.Customer.Name); err != nil {
		return nil, err
	}
	if err := api.Require("customer.email", req.Customer.Email); err != nil {
		return nil, err
	}
	id, err := api.BusinessID(req.BusinessID, s.backend.BusinessID())
	if err != nil {
		return nil, err
	}
	var out PaymentResponse
	if err := s.backend.Post(ctx, "/checkout/payments", req, &out, api.BusinessHeader(id)); err != nil {
		return nil, err
	}
	return &out, nil
}

// Verify retrieves a checkout transaction by the merchant reference.
func (s *Service) Verify(ctx context.Context, reference string) (*TransactionResponse, error) {
	reference = strings.TrimSpace(reference)
	if err := api.Require("reference", reference); err != nil {
		return nil, err
	}
	var out TransactionResponse
	path := "/checkout/payments/merchant-reference/" + url.PathEscape(reference)
	if err := s.backend.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CalculateFee estimates a hosted-checkout fee for the requested currency,
// payment method, and amount.
func (s *Service) CalculateFee(ctx context.Context, req *FeeRequest) (*FeeResponse, error) {
	if req == nil {
		return nil, api.Missing("request")
	}
	if err := api.Require("currency", req.Currency); err != nil {
		return nil, err
	}
	if err := api.Require("type", req.Type); err != nil {
		return nil, err
	}
	if req.Amount <= 0 {
		return nil, api.Invalid("amount", "must be greater than zero")
	}
	id, err := api.BusinessID(req.BusinessID, s.backend.BusinessID())
	if err != nil {
		return nil, err
	}
	payload := *req
	payload.BusinessID = ""
	var out FeeResponse
	if err := s.backend.Post(ctx, "/checkout/data/fees", &payload, &out, api.BusinessHeader(id)); err != nil {
		return nil, err
	}
	return &out, nil
}

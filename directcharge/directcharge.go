package directcharge

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/1Btln/fincra-go/internal/api"
)

// Backend is the subset of the Fincra client this package needs.
type Backend interface {
	api.HTTPClient
	BusinessID() string
}

// Service groups Fincra direct-charge endpoints.
type Service struct{ backend Backend }

// New returns a direct-charge service backed by backend.
func New(backend Backend) *Service { return &Service{backend: backend} }

// Initiate starts a card, bank-transfer, mobile-money, EFT, or PayAttitude charge.
func (s *Service) Initiate(ctx context.Context, req *InitiateRequest) (*ChargeResponse, error) {
	if req == nil {
		return nil, api.Missing("request")
	}
	if err := api.Require("type", req.Type); err != nil {
		return nil, err
	}
	if err := api.Require("currency", req.Currency); err != nil {
		return nil, err
	}
	if err := api.Require("reference", req.Reference); err != nil {
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
	payload := *req
	payload.BusinessID = ""
	var out ChargeResponse
	if err := s.backend.Post(ctx, "/checkout/charges", &payload, &out, api.BusinessHeader(id)); err != nil {
		return nil, err
	}
	return &out, nil
}

// Authorize supplies an OTP, PIN, or address-verification data for a pending charge.
func (s *Service) Authorize(ctx context.Context, chargeID string, req *AuthorizeRequest, businessID string) (*ChargeResponse, error) {
	chargeID = strings.TrimSpace(chargeID)
	if err := api.Require("chargeID", chargeID); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, api.Missing("request")
	}
	var opts []func(*http.Request)
	if id := strings.TrimSpace(businessID); id != "" {
		opts = append(opts, api.BusinessHeader(id))
	}
	var out ChargeResponse
	path := "/checkout/charges/" + url.PathEscape(chargeID) + "/authorize"
	if err := s.backend.Post(ctx, path, req, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// Verify fetches a charge by its merchant reference.
func (s *Service) Verify(ctx context.Context, reference string) (*ChargeResponse, error) {
	reference = strings.TrimSpace(reference)
	if err := api.Require("reference", reference); err != nil {
		return nil, err
	}
	var out ChargeResponse
	path := "/checkout/charges/merchant-reference/" + url.PathEscape(reference)
	if err := s.backend.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

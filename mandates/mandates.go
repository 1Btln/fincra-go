package mandates

import (
	"context"
	"net/url"
	"strings"

	"github.com/1Btln/fincra-go/internal/api"
)

// Backend is the subset of the Fincra client this package needs.
type Backend interface {
	api.HTTPClient
	BusinessID() string
}

// Service groups direct-debit mandate endpoints.
type Service struct{ backend Backend }

// New returns a mandate service backed by backend.
func New(backend Backend) *Service { return &Service{backend: backend} }

// Create creates a direct-debit mandate.
func (s *Service) Create(ctx context.Context, req *CreateRequest) (*MandateResponse, error) {
	if req == nil {
		return nil, api.Missing("request")
	}
	if err := api.Require("amount", req.Amount); err != nil {
		return nil, err
	}
	if err := api.Require("description", req.Description); err != nil {
		return nil, err
	}
	if err := api.Require("startDate", req.StartDate); err != nil {
		return nil, err
	}
	if err := api.Require("endDate", req.EndDate); err != nil {
		return nil, err
	}
	if err := api.Require("businessAddress", req.BusinessAddress); err != nil {
		return nil, err
	}
	if err := api.Require("currency", req.Currency); err != nil {
		return nil, err
	}
	if req.Customer == nil {
		return nil, api.Missing("customer")
	}
	var out MandateResponse
	if err := s.backend.Post(ctx, "/mandate-mgt/mandates/", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a mandate by its Fincra mandate reference.
func (s *Service) Get(ctx context.Context, reference string) (*MandateResponse, error) {
	reference = strings.TrimSpace(reference)
	if err := api.Require("reference", reference); err != nil {
		return nil, err
	}
	var out MandateResponse
	path := "/mandate-mgt/mandates/reference/" + url.PathEscape(reference)
	if err := s.backend.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Debit initiates a debit against an approved mandate.
func (s *Service) Debit(ctx context.Context, req *DebitRequest) (*DebitResponse, error) {
	if req == nil {
		return nil, api.Missing("request")
	}
	for field, value := range map[string]string{
		"amount": req.Amount, "description": req.Description, "mandateReference": req.MandateReference,
		"reference": req.Reference, "currency": req.Currency, "bankCode": req.BankCode,
		"accountNumber": req.AccountNumber, "initiatorAccountName": req.InitiatorAccountName,
		"beneficiaryNarration": req.BeneficiaryNarration,
	} {
		if err := api.Require(field, value); err != nil {
			return nil, err
		}
	}
	var out DebitResponse
	if err := s.backend.Post(ctx, "/v2/mandate-mgt/mandates/payment", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// VerifyDebit retrieves a mandate debit by the transaction reference supplied to Debit.
func (s *Service) VerifyDebit(ctx context.Context, reference string) (*DebitResponse, error) {
	reference = strings.TrimSpace(reference)
	if err := api.Require("reference", reference); err != nil {
		return nil, err
	}
	var out DebitResponse
	path := "/v2/mandate-mgt/payments/" + url.PathEscape(reference)
	if err := s.backend.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

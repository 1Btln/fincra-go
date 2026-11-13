package collections

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

// Service groups Fincra deposit, refund, and chargeback endpoints. Access it
// through Client.Collections.
type Service struct {
	backend Backend
}

// New returns a collections service backed by backend.
func New(backend Backend) *Service {
	return &Service{backend: backend}
}

// ListDeposits returns deposits for a business and virtual account.
func (s *Service) ListDeposits(ctx context.Context, params *DepositParams) (*DepositListResponse, error) {
	if params == nil {
		params = &DepositParams{}
	}
	businessID, err := api.BusinessID(params.BusinessID, s.backend.BusinessID())
	if err != nil {
		return nil, err
	}
	virtualAccount := strings.TrimSpace(params.VirtualAccount)
	if err := api.Require("virtualAccount", virtualAccount); err != nil {
		return nil, err
	}

	query := url.Values{
		"business":       {businessID},
		"virtualAccount": {virtualAccount},
	}
	var out DepositListResponse
	if err := s.backend.Do(ctx, http.MethodGet, "/collections?"+query.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListChargebacks returns chargebacks for a business.
func (s *Service) ListChargebacks(ctx context.Context, businessID string) (*ChargebackListResponse, error) {
	id, err := api.BusinessID(businessID, s.backend.BusinessID())
	if err != nil {
		return nil, err
	}
	query := url.Values{"business": {id}}

	var out ChargebackListResponse
	if err := s.backend.Do(ctx, http.MethodGet, "/collections/chargebacks?"+query.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AcceptChargeback accepts a chargeback on behalf of the business.
func (s *Service) AcceptChargeback(ctx context.Context, chargebackID, businessID string) (*ActionResponse, error) {
	id, err := chargebackIDAndBusiness(chargebackID, businessID, s.backend.BusinessID())
	if err != nil {
		return nil, err
	}
	query := url.Values{"business": {id.business}}

	var out ActionResponse
	path := "/collections/chargebacks/" + url.PathEscape(id.chargeback) + "/accept?" + query.Encode()
	if err := s.backend.Do(ctx, http.MethodPatch, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RejectChargeback rejects a chargeback with a reason.
func (s *Service) RejectChargeback(ctx context.Context, chargebackID, businessID, reason string) (*ActionResponse, error) {
	id, err := chargebackIDAndBusiness(chargebackID, businessID, s.backend.BusinessID())
	if err != nil {
		return nil, err
	}
	if err := api.Require("reason", reason); err != nil {
		return nil, err
	}
	query := url.Values{"business": {id.business}}
	payload := RejectChargebackRequest{BusinessRejectReason: reason}

	var out ActionResponse
	path := "/collections/chargebacks/" + url.PathEscape(id.chargeback) + "/reject?" + query.Encode()
	if err := s.backend.Do(ctx, http.MethodPatch, path, &payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type chargebackIDs struct {
	chargeback string
	business   string
}

func chargebackIDAndBusiness(chargebackID, override, fallback string) (chargebackIDs, error) {
	chargebackID = strings.TrimSpace(chargebackID)
	if err := api.Require("chargebackID", chargebackID); err != nil {
		return chargebackIDs{}, err
	}
	businessID, err := api.BusinessID(override, fallback)
	if err != nil {
		return chargebackIDs{}, err
	}
	return chargebackIDs{chargeback: chargebackID, business: businessID}, nil
}

// VerifyDeposit returns a deposit by the merchant reference.
func (s *Service) VerifyDeposit(ctx context.Context, merchantReference string) (*DepositResponse, error) {
	merchantReference = strings.TrimSpace(merchantReference)
	if err := api.Require("merchantReference", merchantReference); err != nil {
		return nil, err
	}

	var out DepositResponse
	path := "/collections/merchant-reference/" + url.PathEscape(merchantReference)
	if err := s.backend.Do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// InitiateRefund starts a refund for a collection.
func (s *Service) InitiateRefund(ctx context.Context, transReference string) (*RefundResponse, error) {
	transReference = strings.TrimSpace(transReference)
	if err := api.Require("transReference", transReference); err != nil {
		return nil, err
	}
	payload := InitiateRefundRequest{TransReference: transReference}

	var out RefundResponse
	if err := s.backend.Do(ctx, http.MethodPost, "/collections/refund", &payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetRefund returns a refund by its Fincra reference.
func (s *Service) GetRefund(ctx context.Context, reference string) (*RefundResponse, error) {
	reference = strings.TrimSpace(reference)
	if err := api.Require("reference", reference); err != nil {
		return nil, err
	}

	var out RefundResponse
	if err := s.backend.Do(ctx, http.MethodGet, "/collections/refund/"+url.PathEscape(reference), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListRefunds returns refunds for a business.
func (s *Service) ListRefunds(ctx context.Context, businessID string) (*RefundListResponse, error) {
	id, err := api.BusinessID(businessID, s.backend.BusinessID())
	if err != nil {
		return nil, err
	}
	query := url.Values{"business": {id}}

	var out RefundListResponse
	if err := s.backend.Do(ctx, http.MethodGet, "/collections/refund?"+query.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

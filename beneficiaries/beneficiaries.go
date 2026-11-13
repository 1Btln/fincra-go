package beneficiaries

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/1Btln/fincra-go/internal/api"
)

// Backend is the subset of the Fincra client this package needs.
type Backend interface {
	api.HTTPClient
	BusinessID() string
}

// Service groups saved-recipient endpoints. Access it through Client.Beneficiaries.
type Service struct{ backend Backend }

// New returns a beneficiary service backed by backend.
func New(backend Backend) *Service { return &Service{backend: backend} }

func (s *Service) businessPath(override string) (string, error) {
	id, err := api.BusinessID(override, s.backend.BusinessID())
	if err != nil {
		return "", err
	}
	return "/profile/beneficiaries/business/" + url.PathEscape(id), nil
}

// List returns beneficiaries attached to a business.
func (s *Service) List(ctx context.Context, params *ListParams) (*BeneficiaryListResponse, error) {
	if params == nil {
		params = &ListParams{}
	}
	path, err := s.businessPath(params.BusinessID)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	if params.Page > 0 {
		query.Set("page", strconv.Itoa(params.Page))
	}
	if params.PerPage > 0 {
		query.Set("perPage", strconv.Itoa(params.PerPage))
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	var out BeneficiaryListResponse
	if err := s.backend.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches one beneficiary by ID.
func (s *Service) Get(ctx context.Context, beneficiaryID string) (*BeneficiaryResponse, error) {
	return s.GetForBusiness(ctx, beneficiaryID, "")
}

// GetForBusiness fetches a beneficiary, optionally overriding the configured business ID.
func (s *Service) GetForBusiness(ctx context.Context, beneficiaryID, businessID string) (*BeneficiaryResponse, error) {
	beneficiaryID = strings.TrimSpace(beneficiaryID)
	if err := api.Require("beneficiaryID", beneficiaryID); err != nil {
		return nil, err
	}
	path, err := s.businessPath(businessID)
	if err != nil {
		return nil, err
	}
	var out BeneficiaryResponse
	path += "/" + url.PathEscape(beneficiaryID)
	if err := s.backend.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Create saves a beneficiary under the configured business.
func (s *Service) Create(ctx context.Context, req *CreateBeneficiaryRequest) (*BeneficiaryResponse, error) {
	return s.CreateForBusiness(ctx, "", req)
}

// CreateForBusiness saves a beneficiary under businessID.
func (s *Service) CreateForBusiness(ctx context.Context, businessID string, req *CreateBeneficiaryRequest) (*BeneficiaryResponse, error) {
	if req == nil {
		return nil, api.Missing("request")
	}
	if err := api.Require("firstName", req.FirstName); err != nil {
		return nil, err
	}
	if err := api.Require("accountHolderName", req.AccountHolderName); err != nil {
		return nil, err
	}
	if err := api.Require("type", req.Type); err != nil {
		return nil, err
	}
	if err := api.Require("currency", req.Currency); err != nil {
		return nil, err
	}
	if err := api.Require("paymentDestination", req.PaymentDestination); err != nil {
		return nil, err
	}
	if err := api.Require("destinationAddress", req.DestinationAddress); err != nil {
		return nil, err
	}
	path, err := s.businessPath(businessID)
	if err != nil {
		return nil, err
	}
	var out BeneficiaryResponse
	if err := s.backend.Post(ctx, path, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update replaces the editable fields on a saved beneficiary.
func (s *Service) Update(ctx context.Context, beneficiaryID string, req *UpdateBeneficiaryRequest) (*BeneficiaryResponse, error) {
	return s.UpdateForBusiness(ctx, "", beneficiaryID, req)
}

// UpdateForBusiness updates a beneficiary under businessID.
func (s *Service) UpdateForBusiness(ctx context.Context, businessID, beneficiaryID string, req *UpdateBeneficiaryRequest) (*BeneficiaryResponse, error) {
	beneficiaryID = strings.TrimSpace(beneficiaryID)
	if err := api.Require("beneficiaryID", beneficiaryID); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, api.Missing("request")
	}
	path, err := s.businessPath(businessID)
	if err != nil {
		return nil, err
	}
	var out BeneficiaryResponse
	path += "/" + url.PathEscape(beneficiaryID)
	if err := s.backend.Patch(ctx, path, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a beneficiary by ID.
func (s *Service) Delete(ctx context.Context, beneficiaryID string) (*ActionResponse, error) {
	return s.DeleteForBusiness(ctx, "", beneficiaryID)
}

// DeleteForBusiness removes a beneficiary under businessID.
func (s *Service) DeleteForBusiness(ctx context.Context, businessID, beneficiaryID string) (*ActionResponse, error) {
	beneficiaryID = strings.TrimSpace(beneficiaryID)
	if err := api.Require("beneficiaryID", beneficiaryID); err != nil {
		return nil, err
	}
	path, err := s.businessPath(businessID)
	if err != nil {
		return nil, err
	}
	var out ActionResponse
	path += "/" + url.PathEscape(beneficiaryID)
	if err := s.backend.Delete(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

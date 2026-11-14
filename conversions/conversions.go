package conversions

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

// Service groups FX quote, rates, and wallet conversion endpoints.
type Service struct{ backend Backend }

// New returns a conversions service backed by backend.
func New(backend Backend) *Service { return &Service{backend: backend} }

// GenerateQuote requests a short-lived FX quote. Fincra quotes expire quickly;
// callers should initiate the associated transaction promptly.
func (s *Service) GenerateQuote(ctx context.Context, req *QuoteRequest) (*QuoteResponse, error) {
	if req == nil {
		return nil, api.Missing("request")
	}
	if err := api.Require("sourceCurrency", req.SourceCurrency); err != nil {
		return nil, err
	}
	if err := api.Require("destinationCurrency", req.DestinationCurrency); err != nil {
		return nil, err
	}
	if err := api.Require("amount", req.Amount); err != nil {
		return nil, err
	}
	if err := api.Require("action", req.Action); err != nil {
		return nil, err
	}
	id, err := api.BusinessID(req.Business, s.backend.BusinessID())
	if err != nil {
		return nil, err
	}
	payload := *req
	payload.Business = id
	var out QuoteResponse
	if err := s.backend.Post(ctx, "/quotes/generate", &payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Initiate converts wallet funds using an unexpired quote reference.
func (s *Service) Initiate(ctx context.Context, req *InitiateRequest) (*ConversionResponse, error) {
	if req == nil {
		return nil, api.Missing("request")
	}
	if err := api.Require("quoteReference", req.QuoteReference); err != nil {
		return nil, err
	}
	if err := api.Require("customerReference", req.CustomerReference); err != nil {
		return nil, err
	}
	id, err := api.BusinessID(req.Business, s.backend.BusinessID())
	if err != nil {
		return nil, err
	}
	payload := *req
	payload.Business = id
	var out ConversionResponse
	if err := s.backend.Post(ctx, "/conversions/initiate", &payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListRates returns available FX rates, optionally filtered by currency pair.
func (s *Service) ListRates(ctx context.Context, params *RateParams) (*RatesResponse, error) {
	if params == nil {
		params = &RateParams{}
	}
	query := url.Values{}
	api.SetString(query, "currencyPair", strings.ToUpper(params.CurrencyPair))
	api.SetString(query, "currency", strings.ToUpper(params.Currency))
	api.SetString(query, "side", strings.ToLower(params.Side))
	api.SetString(query, "baseCurrency", strings.ToUpper(params.BaseCurrency))
	api.SetString(query, "quoteCurrency", strings.ToUpper(params.QuoteCurrency))
	api.SetInt(query, "page", params.Page)
	api.SetInt(query, "perPage", params.PerPage)
	path := "/quotes/treasury-orders/rates"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	var out RatesResponse
	if err := s.backend.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns the conversions for a business.
func (s *Service) List(ctx context.Context, params *ListParams) (*ConversionListResponse, error) {
	if params == nil {
		params = &ListParams{}
	}
	id, err := api.BusinessID(params.BusinessID, s.backend.BusinessID())
	if err != nil {
		return nil, err
	}
	query := url.Values{"business": {id}}
	var out ConversionListResponse
	if err := s.backend.Get(ctx, "/conversions?"+query.Encode(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns a conversion by its Fincra-generated reference.
func (s *Service) Get(ctx context.Context, reference string) (*ConversionResponse, error) {
	reference = strings.TrimSpace(reference)
	if err := api.Require("conversionReference", reference); err != nil {
		return nil, err
	}
	var out ConversionResponse
	path := "/conversions/reference/" + url.PathEscape(reference)
	if err := s.backend.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

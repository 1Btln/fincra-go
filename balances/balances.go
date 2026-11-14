package balances

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/1Btln/fincra-go/internal/api"
)

const defaultStatementPerPage = 20

// backend is the subset of the Fincra client this package needs. The root
// client satisfies it without an import cycle.
type backend interface {
	api.HTTPClient
	BusinessID() string
}

// Service groups Fincra's balance and wallet endpoints. Access it through
// Client.Balances.
type Service struct {
	backend backend
}

// New returns a Balances service backed by backend.
func New(backend backend) *Service {
	return &Service{backend: backend}
}

// ListWallets returns every wallet belonging to businessID. When businessID is
// empty the client's configured BusinessID is used.
func (s *Service) ListWallets(ctx context.Context, businessID string) (*ListWalletsResponse, error) {
	id, err := api.BusinessID(businessID, s.backend.BusinessID())
	if err != nil {
		return nil, err
	}

	query := url.Values{"businessID": {id}}
	var out ListWalletsResponse
	if err := s.backend.Get(ctx, "/wallets?"+query.Encode(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWallet returns a single wallet by its Fincra wallet ID. The ID is the
// "id" field returned by ListWallets, not the wallet number.
func (s *Service) GetWallet(ctx context.Context, walletID string) (*WalletResponse, error) {
	walletID = strings.TrimSpace(walletID)
	if err := api.Require("walletID", walletID); err != nil {
		return nil, err
	}

	var out WalletResponse
	if err := s.backend.Get(ctx, "/wallets/"+url.PathEscape(walletID), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetBalanceHistory returns pay-ins and pay-outs recorded against the business
// wallet.
//
// Fincra documents this operation as a GET with a JSON request body. This SDK
// sends the same fields as query parameters instead, because GET bodies are not
// reliably forwarded by proxies and load balancers. Use Client.Do directly if
// an integration must mirror the documented wire format exactly.
func (s *Service) GetBalanceHistory(ctx context.Context, params *BalanceHistoryParams) (*BalanceHistoryResponse, error) {
	if params == nil {
		params = &BalanceHistoryParams{}
	}

	businessID, err := api.BusinessID(params.BusinessID, s.backend.BusinessID())
	if err != nil {
		return nil, err
	}

	action := strings.ToLower(strings.TrimSpace(params.Action))
	if action != "" && action != "credit" && action != "debit" {
		return nil, api.Invalid("action", `must be "credit" or "debit"`)
	}
	if err := api.ValidateDate("dateFrom", params.DateFrom); err != nil {
		return nil, err
	}
	if err := api.ValidateDate("dateTo", params.DateTo); err != nil {
		return nil, err
	}

	query := url.Values{"business": {businessID}}
	if action != "" {
		query.Set("action", action)
	}
	api.SetString(query, "dateFrom", params.DateFrom)
	api.SetString(query, "dateTo", params.DateTo)
	api.SetInt(query, "page", params.Page)
	api.SetInt(query, "perPage", params.PerPage)
	api.SetString(query, "amount", params.Amount)

	var out BalanceHistoryResponse
	if err := s.backend.Get(ctx, "/wallets/logs?"+query.Encode(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAccountStatement returns a paginated statement for one wallet currency.
func (s *Service) GetAccountStatement(ctx context.Context, params *AccountStatementParams) (*AccountStatementResponse, error) {
	if params == nil {
		return nil, api.Missing("params")
	}

	dateFrom := strings.TrimSpace(params.DateFrom)
	dateTo := strings.TrimSpace(params.DateTo)
	currency := strings.ToUpper(strings.TrimSpace(params.Currency))

	if err := api.ValidateDate("dateFrom", dateFrom); err != nil {
		return nil, err
	}
	if err := api.Require("dateFrom", dateFrom); err != nil {
		return nil, err
	}
	if err := api.ValidateDate("dateTo", dateTo); err != nil {
		return nil, err
	}
	if err := api.Require("dateTo", dateTo); err != nil {
		return nil, err
	}
	if err := api.Require("currency", currency); err != nil {
		return nil, err
	}

	perPage := params.PerPage
	if perPage <= 0 {
		perPage = defaultStatementPerPage
	}

	query := url.Values{
		"dateFrom": {dateFrom},
		"dateTo":   {dateTo},
		"currency": {currency},
		"perPage":  {strconv.Itoa(perPage)},
	}
	api.SetInt(query, "page", params.Page)

	var opts []func(*http.Request)
	if businessID := strings.TrimSpace(params.BusinessID); businessID != "" {
		opts = append(opts, api.BusinessHeader(businessID))
	}

	var out AccountStatementResponse
	if err := s.backend.Get(ctx, "/wallets/logs/statement?"+query.Encode(), &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTopupHistory returns wallet top-ups for the business.
func (s *Service) GetTopupHistory(ctx context.Context, params *TopupHistoryParams) (*TopupHistoryResponse, error) {
	if params == nil {
		params = &TopupHistoryParams{}
	}

	businessID, err := api.BusinessID(params.BusinessID, s.backend.BusinessID())
	if err != nil {
		return nil, err
	}

	query := url.Values{"business": {businessID}}
	api.SetString(query, "reference", params.Reference)
	api.SetInt(query, "page", params.Page)
	api.SetInt(query, "perPage", params.PerPage)

	var out TopupHistoryResponse
	if err := s.backend.Get(ctx, "/wallets/topups?"+query.Encode(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

package fincra

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/1Btln/fincra-go/account"
	"github.com/1Btln/fincra-go/balances"
	"github.com/1Btln/fincra-go/beneficiaries"
	"github.com/1Btln/fincra-go/collections"
	"github.com/1Btln/fincra-go/internal/api"
	"github.com/1Btln/fincra-go/payouts"
)

const (
	headerAPIKey     = api.HeaderAPIKey
	headerPublicKey  = api.HeaderPublicKey
	headerBusinessID = api.HeaderBusinessID

	// maxResponseBodyBytes protects callers from decoding an unbounded body.
	maxResponseBodyBytes = 10 << 20 // 10 MiB
)

// Client is a Fincra API client. It is safe for concurrent use by multiple
// goroutines once constructed.
type Client struct {
	cfg        Config
	httpClient *http.Client

	maskHeaders     map[string]struct{}
	maskJSONFields  map[string]struct{}
	maskQueryParams map[string]struct{}

	// Account exposes business, merchant, bank, account-resolution, and BVN
	// endpoints.
	Account *account.Service

	// Balances exposes wallet, balance history, statement, and top-up
	// endpoints.
	Balances *balances.Service

	// Beneficiaries exposes saved-recipient management endpoints.
	Beneficiaries *beneficiaries.Service

	// Checkout exposes hosted checkout and fee-calculation endpoints.
	Checkout *checkout.Service

	// Collections exposes deposits, refunds, and chargeback endpoints.
	Collections *collections.Service

	// Conversions exposes FX conversions, quotes, and treasury rates.
	Conversions *conversions.Service

	// DirectCharge exposes card, bank-transfer, and mobile-money charge
	// endpoints.
	DirectCharge *directcharge.Service

	// Mandates exposes direct-debit mandate endpoints.
	Mandates *mandates.Service

	// Payouts exposes disbursement and Fincra-to-Fincra transfer endpoints.
	Payouts *payouts.Service

	// VirtualAccounts exposes permanent virtual-account endpoints.
	VirtualAccounts *virtualaccounts.Service
}

// New returns a Client configured from DefaultConfig, then FINCRA_* environment
// variables, then opts. Options always win.
//
// Environment variables are only a fallback for callers that do not build a
// Config. Use NewWithConfig when configuration comes from your own object or
// configuration system. At minimum a secret key is required from one of the
// three sources:
//
//	client, err := fincra.New(fincra.WithSecretKey(os.Getenv("FINCRA_SECRET_KEY")))
func New(opts ...Option) (*Client, error) {
	cfg := DefaultConfig()
	if err := cfg.applyEnv(os.LookupEnv, true); err != nil {
		return nil, err
	}
	return NewWithConfig(cfg, opts...)
}

// NewWithConfig returns a Client based on cfg plus opts. cfg is cloned before
// defaults and options are applied, so the caller's Config is not modified.
//
// This is the primary constructor: it never reads the environment. To use
// environment variables as a fallback for unset fields, call cfg.ApplyEnv
// before passing it here.
func NewWithConfig(cfg *Config, opts ...Option) (*Client, error) {
	if cfg == nil {
		return nil, &ConfigError{Field: "Config", Reason: "must not be nil"}
	}
	resolved := cfg.Clone()
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(resolved); err != nil {
			return nil, err
		}
	}
	resolved.setDefaults()
	if err := resolved.Validate(); err != nil {
		return nil, err
	}

	// Copy the caller's http.Client so the SDK can add a redirect guard without
	// mutating the value they own. The underlying Transport and Jar are shared.
	httpClient := *resolved.HTTPClient
	if httpClient.CheckRedirect == nil {
		httpClient.CheckRedirect = sameHostRedirectPolicy
	}
	resolved.HTTPClient = &httpClient

	client := &Client{
		cfg:             *resolved,
		httpClient:      &httpClient,
		maskHeaders:     newStringSet(resolved.Log.MaskHeaders),
		maskJSONFields:  newStringSet(resolved.Log.MaskJSONFields),
		maskQueryParams: newStringSet(resolved.Log.MaskQueryParams),
	}
	client.Account = account.New(client)
	client.Balances = balances.New(client)
	client.Beneficiaries = beneficiaries.New(client)
	client.Checkout = checkout.New(client)
	client.Collections = collections.New(client)
	client.Conversions = conversions.New(client)
	client.DirectCharge = directcharge.New(client)
	client.Mandates = mandates.New(client)
	client.Payouts = payouts.New(client)
	client.VirtualAccounts = virtualaccounts.New(client)
	return client, nil
}

var errCrossHostRedirect = errors.New("fincra: refusing cross-host redirect")

// sameHostRedirectPolicy prevents the api-key header from being forwarded to a
// different host by a redirect. Fincra does not document cross-host redirects,
// and dropping them is safer than leaking credentials to an unexpected origin.
func sameHostRedirectPolicy(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("fincra: stopped after 10 redirects")
	}
	if len(via) > 0 && !strings.EqualFold(req.URL.Host, via[0].URL.Host) {
		return fmt.Errorf("%w to %s", errCrossHostRedirect, req.URL.Host)
	}
	return nil
}

// Config returns a copy of the client configuration. Logging mask slices are
// copied so callers cannot mutate the running client.
func (c *Client) Config() Config {
	cfg := c.cfg
	cfg.Log = cfg.Log.Clone()
	return cfg
}

// BaseURL returns the base URL requests are sent to.
func (c *Client) BaseURL() string { return c.cfg.ResolveBaseURL() }

// BusinessID returns the configured default business or sub-account identifier.
// Resource packages use it when a request omits a business ID.
func (c *Client) BusinessID() string { return c.cfg.BusinessID }

// RequestOption customises an individual request before it is sent.
//
// It is an alias for func(*http.Request) so resource packages can define small
// backend interfaces without importing this package.
type RequestOption = func(*http.Request)

// WithRequestHeader sets a header on a single request, overriding any default
// value for the same header.
func WithRequestHeader(key, value string) RequestOption {
	return func(r *http.Request) {
		r.Header.Set(key, value)
	}
}

// Do executes an authenticated request against Fincra.
//
// path is relative to the configured base URL and may include a query string,
// for example "/wallets?businessID=123". An absolute http(s) URL is accepted
// only when it shares the configured base URL's origin, so the api-key header
// is never sent to an unrelated host. body is JSON-encoded when non-nil. If out
// is non-nil, a successful JSON response is decoded into it.
//
// Non-2xx responses are returned as *APIError. Retryable failures (network
// errors, 408, 425, 429, and common 5xx responses) are retried up to
// Config.MaxRetries with exponential backoff and jitter, respecting a
// Retry-After header when present.
func (c *Client) Do(ctx context.Context, method, path string, body, out any, opts ...RequestOption) error {
	if ctx == nil {
		ctx = context.Background()
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = http.MethodGet
	}
	if strings.TrimSpace(path) == "" {
		return errors.New("fincra: request path must not be empty")
	}

	var payload []byte
	if body != nil {
		var err error
		if reader, ok := body.(io.Reader); ok {
			payload, err = io.ReadAll(reader)
			if err != nil {
				return fmt.Errorf("fincra: read request body: %w", err)
			}
		} else {
			payload, err = json.Marshal(body)
			if err != nil {
				return fmt.Errorf("fincra: encode request body: %w", err)
			}
		}
	}

	if c.cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.cfg.Timeout)
		defer cancel()
	}

	var (
		lastErr     error
		retryAfter  time.Duration
		retryReason string
	)
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := c.retryDelay(attempt, retryAfter)
			c.logRetry(ctx, method, path, attempt+1, delay, retryReason)
			select {
			case <-ctx.Done():
				return fmt.Errorf("fincra: %s %s: %w", method, path, ctx.Err())
			case <-time.After(delay):
			}
		}

		req, err := c.newRequest(ctx, method, path, payload, opts)
		if err != nil {
			return err
		}

		started := time.Now()
		resp, err := c.httpClient.Do(req)
		duration := time.Since(started)
		if err != nil {
			c.logTransportError(ctx, req, err, duration, attempt)
			lastErr = fmt.Errorf("fincra: %s %s: %w", method, path, err)
			if !isRetryableTransportError(err) || attempt == c.cfg.MaxRetries {
				return lastErr
			}
			retryAfter = 0
			retryReason = "transport error"
			continue
		}

		body, ra, retry, responseErr := c.handleResponse(resp, method, path, out)
		c.logResponse(ctx, req, payload, resp, body, duration, attempt, responseErr)
		if responseErr == nil {
			return nil
		}
		lastErr = responseErr
		if !retry || attempt == c.cfg.MaxRetries {
			return lastErr
		}
		retryAfter = ra
		retryReason = retryReasonFor(responseErr)
	}
	return lastErr
}

func retryReasonFor(err error) string {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return "status " + strconv.Itoa(apiErr.StatusCode)
	}
	return "request failed"
}

func (c *Client) newRequest(ctx context.Context, method, path string, payload []byte, opts []RequestOption) (*http.Request, error) {
	endpoint, err := c.resolveURL(path)
	if err != nil {
		return nil, err
	}

	var reader io.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, fmt.Errorf("fincra: build request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set(headerAPIKey, c.cfg.SecretKey)
	if c.cfg.PublicKey != "" {
		req.Header.Set(headerPublicKey, c.cfg.PublicKey)
	}
	if c.cfg.BusinessID != "" {
		req.Header.Set(headerBusinessID, c.cfg.BusinessID)
	}
	if len(payload) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, opt := range opts {
		if opt != nil {
			opt(req)
		}
	}
	return req, nil
}

func (c *Client) resolveURL(path string) (string, error) {
	path = strings.TrimSpace(path)
	base, err := url.Parse(c.cfg.ResolveBaseURL())
	if err != nil {
		return "", fmt.Errorf("fincra: invalid base URL: %w", err)
	}

	if u, err := url.Parse(path); err == nil && u.IsAbs() {
		if u.Scheme != "http" && u.Scheme != "https" {
			return "", fmt.Errorf("fincra: unsupported URL scheme %q", u.Scheme)
		}
		// Only allow absolute URLs on the configured origin so the api-key
		// header can never be sent to an unrelated host.
		if !strings.EqualFold(u.Scheme, base.Scheme) || !strings.EqualFold(u.Host, base.Host) {
			return "", fmt.Errorf("fincra: absolute URL %q must match the configured base URL %q", u.Redacted(), base.Redacted())
		}
		return u.String(), nil
	}

	return strings.TrimRight(base.String(), "/") + "/" + strings.TrimLeft(path, "/"), nil
}

// handleResponse consumes resp and returns the body for logging, the retry
// delay, whether the failure is retryable, and the resulting error.
func (c *Client) handleResponse(resp *http.Response, method, path string, out any) ([]byte, time.Duration, bool, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodyBytes+1))
	if err != nil {
		return body, 0, false, &APIError{
			StatusCode: resp.StatusCode,
			Method:     method,
			Path:       path,
			Message:    "failed to read response body",
			Err:        err,
		}
	}
	if len(body) > maxResponseBodyBytes {
		return body, 0, false, &APIError{
			StatusCode: resp.StatusCode,
			Method:     method,
			Path:       path,
			Message:    "response body exceeds the 10 MiB limit",
		}
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		if out != nil && len(bytes.TrimSpace(body)) > 0 {
			if err := json.Unmarshal(body, out); err != nil {
				return body, 0, false, &APIError{
					StatusCode: resp.StatusCode,
					Method:     method,
					Path:       path,
					Message:    "failed to decode response body",
					Body:       body,
					Err:        err,
				}
			}
		}
		return body, 0, false, nil
	}

	apiErr := newAPIError(resp, method, path, body)
	return body, parseRetryAfter(resp.Header.Get("Retry-After")), apiErr.Retryable, apiErr
}

type errorResponse struct {
	Message   string          `json:"message"`
	Error     string          `json:"error"`
	ErrorType string          `json:"errorType"`
	Errors    json.RawMessage `json:"errors"`
}

func newAPIError(resp *http.Response, method, path string, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		Method:     method,
		Path:       path,
		Body:       body,
		RequestID: firstNonEmpty(
			resp.Header.Get("x-request-id"),
			resp.Header.Get("x-fincra-request-id"),
		),
		Retryable: isRetryableStatus(resp.StatusCode),
	}

	var parsed errorResponse
	if err := json.Unmarshal(body, &parsed); err == nil {
		apiErr.ErrorType = parsed.ErrorType
		if len(parsed.Errors) > 0 && string(parsed.Errors) != "null" {
			apiErr.Details = parsed.Errors
		}
		apiErr.Message = firstNonEmpty(parsed.Error, parsed.Message)
	}
	if apiErr.Message == "" {
		apiErr.Message = http.StatusText(resp.StatusCode)
	}
	return apiErr
}

func (c *Client) retryDelay(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
		return retryAfter
	}

	base := c.cfg.RetryWaitMin
	for i := 1; i < attempt && base < c.cfg.RetryWaitMax; i++ {
		base *= 2
	}
	if base > c.cfg.RetryWaitMax || base <= 0 {
		base = c.cfg.RetryWaitMax
	}
	if base <= 0 {
		return 0
	}

	half := base / 2
	if half <= 0 {
		return base
	}
	return half + time.Duration(rand.Int64N(int64(half)+1))
}

func isRetryableStatus(code int) bool {
	switch code {
	case http.StatusRequestTimeout, // 408
		http.StatusTooEarly,            // 425
		http.StatusTooManyRequests,     // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	default:
		return false
	}
}

func isRetryableTransportError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	if errors.Is(err, errCrossHostRedirect) {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return false
	}
	return true
}

func parseRetryAfter(value string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds <= 0 {
			return 0
		}
		return time.Duration(seconds) * time.Second
	}
	if when, err := http.ParseTime(value); err == nil {
		if d := time.Until(when); d > 0 {
			return d
		}
	}
	return 0
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

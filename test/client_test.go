package fincra_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	fincra "github.com/1Btln/fincra-go"
)

func newTestClient(t *testing.T, baseURL string, opts ...fincra.Option) *fincra.Client {
	t.Helper()

	cfg := fincra.DefaultConfig()
	cfg.SecretKey = "sk_test_123"
	cfg.PublicKey = "pk_test_456"
	cfg.BusinessID = "biz_789"
	cfg.BaseURL = baseURL
	cfg.RetryWaitMin = time.Millisecond
	cfg.RetryWaitMax = 2 * time.Millisecond

	client, err := fincra.NewWithConfig(cfg, opts...)
	if err != nil {
		t.Fatalf("NewWithConfig: %v", err)
	}
	return client
}

func TestNewRequiresSecretKey(t *testing.T) {
	clearFincraEnv(t)

	_, err := fincra.New()
	if err == nil {
		t.Fatal("New() error = nil, want error")
	}
	var cfgErr *fincra.ConfigError
	if !errors.As(err, &cfgErr) || cfgErr.Field != "SecretKey" {
		t.Fatalf("New() error = %v, want *fincra.ConfigError for SecretKey", err)
	}
}

func TestNewNilConfig(t *testing.T) {
	if _, err := fincra.NewWithConfig(nil); err == nil {
		t.Fatal("NewWithConfig(nil) error = nil, want error")
	}
}

func TestNewWithConfigDoesNotMutateInput(t *testing.T) {
	cfg := fincra.DefaultConfig()
	cfg.SecretKey = "sk_test"

	client, err := fincra.NewWithConfig(cfg, fincra.WithUserAgent("custom-agent"))
	if err != nil {
		t.Fatalf("NewWithConfig: %v", err)
	}
	if cfg.UserAgent != fincra.DefaultUserAgent {
		t.Errorf("input Config.UserAgent = %q, want unchanged", cfg.UserAgent)
	}
	if got := client.Config().UserAgent; got != "custom-agent" {
		t.Errorf("client UserAgent = %q, want custom-agent", got)
	}
}

func TestNewWithConfigDoesNotReadEnvironment(t *testing.T) {
	clearFincraEnv(t)
	t.Setenv(fincra.EnvSecretKey, "sk_env")
	t.Setenv(fincra.EnvBusinessID, "biz_env")
	t.Setenv(fincra.EnvEnvironment, "live")

	cfg := fincra.DefaultConfig()
	cfg.SecretKey = "sk_explicit"
	cfg.BusinessID = "biz_explicit"

	client, err := fincra.NewWithConfig(cfg)
	if err != nil {
		t.Fatalf("NewWithConfig: %v", err)
	}

	got := client.Config()
	if got.SecretKey != "sk_explicit" || got.BusinessID != "biz_explicit" {
		t.Errorf("config = %+v, want explicit values", got)
	}
	if got.Environment != fincra.EnvironmentSandbox {
		t.Errorf("Environment = %q, want sandbox from explicit config", got.Environment)
	}
}

func TestDoSendsHeadersAndDecodesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/wallets" {
			t.Errorf("path = %s, want /wallets", r.URL.Path)
		}
		if got := r.Header.Get("api-key"); got != "sk_test_123" {
			t.Errorf("api-key = %q, want sk_test_123", got)
		}
		if got := r.Header.Get("x-pub-key"); got != "pk_test_456" {
			t.Errorf("x-pub-key = %q, want pk_test_456", got)
		}
		if got := r.Header.Get("x-business-id"); got != "biz_789" {
			t.Errorf("x-business-id = %q, want biz_789", got)
		}
		if got := r.Header.Get("User-Agent"); got != fincra.DefaultUserAgent {
			t.Errorf("User-Agent = %q, want %q", got, fincra.DefaultUserAgent)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q, want application/json", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"success":true,"message":"ok","data":{"name":"fincra"}}`)
	}))
	defer srv.Close()

	var out struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			Name string `json:"name"`
		} `json:"data"`
	}

	client := newTestClient(t, srv.URL)
	if err := client.HTTPClient.Get(context.Background(), "/wallets", &out); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if !out.Success || out.Data.Name != "fincra" {
		t.Errorf("decoded response = %+v", out)
	}
}

func TestDoEncodesJSONBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]string
		if err := json.Unmarshal(body, &got); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if got["reference"] != "ref_1" {
			t.Errorf("body = %v, want reference ref_1", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"success":true}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	body := map[string]string{"reference": "ref_1"}
	if err := client.HTTPClient.Post(context.Background(), "/payments", body, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
}

func TestAPIErrorMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-request-id", "req_123")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"success":false,"error":"Conversion not found","errorType":"RESOURCE_NOT_FOUND","errors":"Conversion not found"}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	err := client.HTTPClient.Get(context.Background(), "/conversions/reference/abc", nil)
	if err == nil {
		t.Fatal("Do() error = nil, want error")
	}

	apiErr, ok := fincra.IsAPIError(err)
	if !ok {
		t.Fatalf("error = %v, want *fincra.APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
	if apiErr.Message != "Conversion not found" {
		t.Errorf("Message = %q", apiErr.Message)
	}
	if apiErr.ErrorType != "RESOURCE_NOT_FOUND" {
		t.Errorf("ErrorType = %q", apiErr.ErrorType)
	}
	if apiErr.RequestID != "req_123" {
		t.Errorf("RequestID = %q, want req_123", apiErr.RequestID)
	}
	if string(apiErr.Details) != `"Conversion not found"` {
		t.Errorf("Details = %s", apiErr.Details)
	}
	if !apiErr.IsNotFound() {
		t.Error("IsNotFound() = false, want true")
	}
	if apiErr.Retryable {
		t.Error("Retryable = true, want false")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("Error() = %q, want status code", err.Error())
	}
}

func TestAPIErrorStatusHelpers(t *testing.T) {
	tests := []struct {
		status int
		check  func(*testing.T, *fincra.APIError)
	}{
		{status: http.StatusUnauthorized, check: func(t *testing.T, e *fincra.APIError) {
			if !e.IsUnauthorized() {
				t.Error("IsUnauthorized() = false")
			}
		}},
		{status: http.StatusForbidden, check: func(t *testing.T, e *fincra.APIError) {
			if !e.IsForbidden() {
				t.Error("IsForbidden() = false")
			}
		}},
		{status: http.StatusTooManyRequests, check: func(t *testing.T, e *fincra.APIError) {
			if !e.IsRateLimited() {
				t.Error("IsRateLimited() = false")
			}
			if !e.Retryable {
				t.Error("Retryable = false, want true")
			}
		}},
		{status: http.StatusBadGateway, check: func(t *testing.T, e *fincra.APIError) {
			if !e.IsServerError() {
				t.Error("IsServerError() = false")
			}
		}},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()

			client := newTestClient(t, srv.URL, fincra.WithMaxRetries(0))
			err := client.HTTPClient.Get(context.Background(), "/x", nil)
			apiErr, ok := fincra.IsAPIError(err)
			if !ok {
				t.Fatalf("error = %v, want *fincra.APIError", err)
			}
			tt.check(t, apiErr)
		})
	}
}

func TestDoRetriesOnServerError(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = io.WriteString(w, `{"success":false,"error":"unavailable"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"success":true}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL, fincra.WithMaxRetries(2))
	if err := client.HTTPClient.Get(context.Background(), "/wallets", nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 3 {
		t.Errorf("attempts = %d, want 3", got)
	}
}

func TestDoRetriesExhausted(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL, fincra.WithMaxRetries(2))
	err := client.HTTPClient.Get(context.Background(), "/wallets", nil)
	if _, ok := fincra.IsAPIError(err); !ok {
		t.Fatalf("error = %v, want *fincra.APIError", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 3 {
		t.Errorf("attempts = %d, want 3", got)
	}
}

func TestDoDoesNotRetryClientError(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL, fincra.WithMaxRetries(3))
	if err := client.HTTPClient.Get(context.Background(), "/wallets", nil); err == nil {
		t.Fatal("Do() error = nil, want error")
	}
	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Errorf("attempts = %d, want 1", got)
	}
}

func TestDoContextCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(t, srv.URL)
	err := client.HTTPClient.Get(ctx, "/wallets", nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestDoEmptyPath(t *testing.T) {
	client := newTestClient(t, "https://example.test")
	if err := client.HTTPClient.Get(context.Background(), "  ", nil); err == nil {
		t.Fatal("Do() error = nil, want error")
	}
}

func TestDoRejectsForeignAbsoluteURL(t *testing.T) {
	client := newTestClient(t, "https://api.example.test")
	err := client.HTTPClient.Get(context.Background(), "https://evil.example.test/steal", nil)
	if err == nil {
		t.Fatal("Do() error = nil, want error for foreign origin")
	}
	if !strings.Contains(err.Error(), "must match the configured base URL") {
		t.Errorf("error = %v, want origin mismatch", err)
	}
}

func TestDoAcceptsSameOriginAbsoluteURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wallets" {
			t.Errorf("path = %s, want /wallets", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"success":true}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if err := client.HTTPClient.Get(context.Background(), srv.URL+"/wallets", nil); err != nil {
		t.Fatalf("Get: %v", err)
	}
}

func TestDoRejectsCrossHostRedirect(t *testing.T) {
	var targetHits int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&targetHits, 1)
	}))
	defer target.Close()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL, fincra.WithMaxRetries(0))
	err := client.HTTPClient.Get(context.Background(), "/wallets", nil)
	if err == nil {
		t.Fatal("Do() error = nil, want redirect error")
	}
	if !strings.Contains(err.Error(), "cross-host redirect") {
		t.Errorf("error = %v, want cross-host redirect error", err)
	}
	if got := atomic.LoadInt32(&targetHits); got != 0 {
		t.Errorf("redirect target hits = %d, want 0", got)
	}
}

func TestRequestOptionHeaderOverride(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("x-business-id"); got != "biz_override" {
			t.Errorf("x-business-id = %q, want biz_override", got)
		}
		_, _ = io.WriteString(w, `{}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	err := client.HTTPClient.Get(
		context.Background(),
		"/wallets",
		nil,
		fincra.WithRequestHeader("x-business-id", "biz_override"),
	)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
}

func TestDoDecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `not-json`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	var out map[string]any
	err := client.HTTPClient.Get(context.Background(), "/wallets", &out)
	var apiErr *fincra.APIError
	if !errors.As(err, &apiErr) || apiErr.Err == nil {
		t.Fatalf("error = %v, want *fincra.APIError with decode cause", err)
	}
}

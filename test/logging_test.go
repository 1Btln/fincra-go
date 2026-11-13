package fincra_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	fincra "github.com/1Btln/fincra-go"
)

func newLoggingClient(t *testing.T, baseURL string, opts ...fincra.Option) (*fincra.Client, *bytes.Buffer) {
	t.Helper()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	all := []fincra.Option{
		fincra.WithLogger(logger),
		fincra.WithLogDepth(fincra.LogDepthBasic),
		fincra.WithLogLevel(slog.LevelDebug),
	}
	all = append(all, opts...)

	return newTestClient(t, baseURL, all...), &buf
}

func TestLogBasicRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"success":true}`)
	}))
	defer srv.Close()

	client, buf := newLoggingClient(t, srv.URL)
	if err := client.Do(context.Background(), http.MethodGet, "/wallets?businessID=biz_789", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}

	logs := buf.String()
	for _, want := range []string{
		`"msg":"fincra request"`,
		`"method":"GET"`,
		`"status":200`,
		`"attempt":1`,
		"/wallets?businessID=biz_789",
	} {
		if !strings.Contains(logs, want) {
			t.Errorf("logs missing %q:\n%s", want, logs)
		}
	}
	for _, unwanted := range []string{"request_headers", "response_headers", "request_body", "response_body", "sk_test_123"} {
		if strings.Contains(logs, unwanted) {
			t.Errorf("basic depth logs contain %q:\n%s", unwanted, logs)
		}
	}
}

func TestLogDepthNoneDisables(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"success":true}`)
	}))
	defer srv.Close()

	client, buf := newLoggingClient(t, srv.URL, fincra.WithLogDepth(fincra.LogDepthNone))
	if err := client.Do(context.Background(), http.MethodGet, "/wallets", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("logs = %q, want none", buf.String())
	}
}

func TestLogLevelFilters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"success":true}`)
	}))
	defer srv.Close()

	client, buf := newLoggingClient(t, srv.URL, fincra.WithLogLevel(slog.LevelWarn))
	if err := client.Do(context.Background(), http.MethodGet, "/wallets", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("info record logged at warn level: %q", buf.String())
	}

	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer failing.Close()

	client, buf = newLoggingClient(t, failing.URL, fincra.WithLogLevel(slog.LevelWarn), fincra.WithMaxRetries(0))
	if err := client.Do(context.Background(), http.MethodGet, "/wallets", nil, nil); err == nil {
		t.Fatal("Do() error = nil, want error")
	}
	if buf.Len() == 0 {
		t.Fatal("warn record was suppressed at warn level")
	}
}

func TestLogHeadersMasked(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-request-id", "req_123")
		_, _ = io.WriteString(w, `{"success":true}`)
	}))
	defer srv.Close()

	client, buf := newLoggingClient(t, srv.URL, fincra.WithLogDepth(fincra.LogDepthHeaders))
	if err := client.Do(context.Background(), http.MethodGet, "/wallets", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}

	logs := buf.String()
	if !strings.Contains(logs, "request_headers") || !strings.Contains(logs, "response_headers") {
		t.Errorf("headers missing from logs:\n%s", logs)
	}
	if strings.Contains(logs, "sk_test_123") {
		t.Errorf("api-key leaked into logs:\n%s", logs)
	}
	if !strings.Contains(logs, `[REDACTED]`) {
		t.Errorf("masked value marker missing:\n%s", logs)
	}
	if !strings.Contains(logs, "biz_789") {
		t.Errorf("non-sensitive x-business-id should be logged:\n%s", logs)
	}
	if !strings.Contains(logs, "req_123") {
		t.Errorf("response request ID should be logged:\n%s", logs)
	}
}

func TestLogQueryParamsMasked(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"success":true}`)
	}))
	defer srv.Close()

	client, buf := newLoggingClient(t, srv.URL)
	if err := client.Do(context.Background(), http.MethodGet, "/wallets?token=super-secret&businessID=biz_789", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}

	logs := buf.String()
	if strings.Contains(logs, "super-secret") {
		t.Errorf("query token leaked into logs:\n%s", logs)
	}
	if !strings.Contains(logs, "token=%5BREDACTED%5D") {
		t.Errorf("query token was not masked:\n%s", logs)
	}
	if !strings.Contains(logs, "businessID=biz_789") {
		t.Errorf("non-sensitive query parameter was removed:\n%s", logs)
	}
}

func TestLogBodyMasked(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"success":true,"data":{"accountNumber":"1234567890","reference":"ref_1"}}`)
	}))
	defer srv.Close()

	client, buf := newLoggingClient(t, srv.URL, fincra.WithLogDepth(fincra.LogDepthBody))
	body := map[string]any{
		"reference": "ref_1",
		"password":  "hunter2",
		"api_key":   "sk_body_secret",
		"nested":    map[string]any{"token": "nested-secret"},
	}
	if err := client.Do(context.Background(), http.MethodPost, "/payments", body, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}

	logs := buf.String()
	for _, secret := range []string{"hunter2", "sk_body_secret", "nested-secret", "1234567890"} {
		if strings.Contains(logs, secret) {
			t.Errorf("body value %q leaked into logs:\n%s", secret, logs)
		}
	}
	for _, want := range []string{"request_body", "response_body", "ref_1", `[REDACTED]`} {
		if !strings.Contains(logs, want) {
			t.Errorf("logs missing %q:\n%s", want, logs)
		}
	}
}

func TestLogBodyTruncated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"success":true,"message":"`+strings.Repeat("x", 512)+`"}`)
	}))
	defer srv.Close()

	client, buf := newLoggingClient(t, srv.URL,
		fincra.WithLogDepth(fincra.LogDepthBody),
		fincra.WithLogBodyLimit(24),
	)
	if err := client.Do(context.Background(), http.MethodGet, "/wallets", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if !strings.Contains(buf.String(), "(truncated)") {
		t.Errorf("body was not truncated:\n%s", buf.String())
	}
}

func TestLogNonJSONBodySuppressed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, "secret=plain-text-token")
	}))
	defer srv.Close()

	client, buf := newLoggingClient(t, srv.URL, fincra.WithLogDepth(fincra.LogDepthBody))
	if err := client.Do(context.Background(), http.MethodGet, "/wallets", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}

	logs := buf.String()
	if strings.Contains(logs, "plain-text-token") {
		t.Errorf("non-JSON body leaked into logs:\n%s", logs)
	}
	if !strings.Contains(logs, "<non-JSON body: 23 bytes>") {
		t.Errorf("non-JSON body was not summarised:\n%s", logs)
	}
}

func TestLogRetry(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = io.WriteString(w, `{"success":true}`)
	}))
	defer srv.Close()

	client, buf := newLoggingClient(t, srv.URL, fincra.WithMaxRetries(2))
	if err := client.Do(context.Background(), http.MethodGet, "/wallets?token=retry-secret", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}

	logs := buf.String()
	if !strings.Contains(logs, `"msg":"fincra retrying"`) {
		t.Errorf("retry record missing:\n%s", logs)
	}
	if !strings.Contains(logs, `"reason":"status 503"`) {
		t.Errorf("retry reason missing:\n%s", logs)
	}
	if strings.Contains(logs, "retry-secret") {
		t.Errorf("retry logs leaked query token:\n%s", logs)
	}
}

func TestLogTransportErrorMasksURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	baseURL := srv.URL
	srv.Close()

	client, buf := newLoggingClient(t, baseURL, fincra.WithMaxRetries(0))
	err := client.Do(context.Background(), http.MethodGet, "/wallets?token=transport-secret", nil, nil)
	if err == nil {
		t.Fatal("Do() error = nil, want transport error")
	}

	logs := buf.String()
	if !strings.Contains(logs, `"msg":"fincra request failed"`) {
		t.Errorf("transport error record missing:\n%s", logs)
	}
	if strings.Contains(logs, "transport-secret") {
		t.Errorf("transport error leaked query token:\n%s", logs)
	}
}

func TestLogMaskOptionsExtendDefaults(t *testing.T) {
	cfg := fincra.DefaultConfig()
	cfg.SecretKey = "sk_test"
	client, err := fincra.NewWithConfig(cfg,
		fincra.WithLogMaskHeaders("x-custom"),
		fincra.WithLogMaskJSONFields("reference"),
		fincra.WithLogMaskQueryParams("businessID"),
	)
	if err != nil {
		t.Fatalf("NewWithConfig: %v", err)
	}

	got := client.Config().Log
	for _, header := range []string{"api-key", "x-custom"} {
		if !slices.Contains(got.MaskHeaders, header) {
			t.Errorf("MaskHeaders = %v, want %q", got.MaskHeaders, header)
		}
	}
	for _, field := range []string{"password", "reference"} {
		if !slices.Contains(got.MaskJSONFields, field) {
			t.Errorf("MaskJSONFields = %v, want %q", got.MaskJSONFields, field)
		}
	}
	for _, param := range []string{"token", "businessID"} {
		if !slices.Contains(got.MaskQueryParams, param) {
			t.Errorf("MaskQueryParams = %v, want %q", got.MaskQueryParams, param)
		}
	}
}

func TestLogConfigClonedByOption(t *testing.T) {
	headers := []string{"x-custom"}
	mask := fincra.LogConfig{
		MaskHeaders:     headers,
		MaskJSONFields:  []string{"reference"},
		MaskQueryParams: []string{"businessID"},
	}
	client, err := fincra.NewWithConfig(func() *fincra.Config {
		cfg := fincra.DefaultConfig()
		cfg.SecretKey = "sk_test"
		return cfg
	}(), fincra.WithLogConfig(mask))
	if err != nil {
		t.Fatalf("NewWithConfig: %v", err)
	}

	headers[0] = "mutated"
	if got := client.Config().Log.MaskHeaders[0]; got != "x-custom" {
		t.Errorf("MaskHeaders[0] = %q, want x-custom", got)
	}
}

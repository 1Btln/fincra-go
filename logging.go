package fincra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

// LogDepth controls how much request and response detail the client logs.
//
// Depth is deliberately separate from Level: Level filters severity, while
// Depth decides which attributes are attached to a record. Credentials and
// configured sensitive fields are masked at every depth.
type LogDepth int

const (
	// LogDepthNone disables logging even when a Logger is configured.
	LogDepthNone LogDepth = -1

	// LogDepthBasic logs one record per request with method, URL, status,
	// duration, and attempt number.
	LogDepthBasic LogDepth = 0

	// LogDepthHeaders adds request and response headers, with sensitive header
	// values masked.
	LogDepthHeaders LogDepth = 1

	// LogDepthBody adds request and response bodies, with configured JSON
	// fields masked and bodies truncated to LogConfig.MaxBodyBytes.
	LogDepthBody LogDepth = 2
)

// String implements fmt.Stringer.
func (d LogDepth) String() string {
	switch d {
	case LogDepthNone:
		return "none"
	case LogDepthBasic:
		return "basic"
	case LogDepthHeaders:
		return "headers"
	case LogDepthBody:
		return "body"
	default:
		return "unknown"
	}
}

// DefaultLogBodyBytes caps the body text attached to a log record.
const DefaultLogBodyBytes = 2048

// LogConfig controls structured logging.
//
// Logging is disabled when Logger is nil or Depth is LogDepthNone. The zero
// value is safe: Level defaults to slog.LevelInfo, Depth to LogDepthBasic, and
// the default masking lists are applied by New and NewWithConfig.
type LogConfig struct {
	// Logger receives records. Pass a slog logger configured with the handler
	// and destination you want.
	Logger *slog.Logger

	// Level is the minimum record severity emitted by the SDK. The zero value
	// is slog.LevelInfo.
	//
	// The SDK emits request outcomes at Info (Warn for 5xx, transport failures,
	// and malformed responses) and retry decisions at Debug.
	Level slog.Level

	// Depth selects how much request and response detail is attached.
	Depth LogDepth

	// MaskHeaders lists header names whose values are replaced with
	// "[REDACTED]". Matching is case-insensitive. A nil slice uses
	// DefaultMaskedHeaders; an empty non-nil slice disables header masking.
	MaskHeaders []string

	// MaskJSONFields lists JSON object keys whose values are replaced with
	// "[REDACTED]" in logged request and response bodies. Matching is
	// case-insensitive. A nil slice uses DefaultMaskedJSONFields; an empty
	// non-nil slice disables field masking.
	MaskJSONFields []string

	// MaskQueryParams lists URL query parameter names whose values are replaced
	// with "[REDACTED]" in logged URLs. Matching is case-insensitive. A nil
	// slice uses DefaultMaskedQueryParams; an empty non-nil slice disables query
	// masking.
	MaskQueryParams []string

	// MaxBodyBytes truncates logged bodies. Zero uses DefaultLogBodyBytes.
	MaxBodyBytes int
}

// Clone returns a copy of the LogConfig with independent mask slices.
func (l LogConfig) Clone() LogConfig {
	clone := l
	clone.MaskHeaders = cloneStrings(l.MaskHeaders)
	clone.MaskJSONFields = cloneStrings(l.MaskJSONFields)
	clone.MaskQueryParams = cloneStrings(l.MaskQueryParams)
	return clone
}

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}
	clone := make([]string, len(values))
	copy(clone, values)
	return clone
}

// DefaultMaskedHeaders returns the header names masked by default.
func DefaultMaskedHeaders() []string {
	return []string{
		"api-key",
		"x-pub-key",
		"authorization",
		"proxy-authorization",
		"cookie",
		"set-cookie",
	}
}

// DefaultMaskedJSONFields returns the JSON keys masked by default. The list
// covers credentials, one-time secrets, and strong personal identifiers.
func DefaultMaskedJSONFields() []string {
	return []string{
		"api_key",
		"apikey",
		"password",
		"pin",
		"cvv",
		"otp",
		"token",
		"access_token",
		"accesstoken",
		"refresh_token",
		"refreshtoken",
		"secret",
		"secret_key",
		"secretkey",
		"bvn",
		"nin",
		"accountnumber",
		"bankaccountnumber",
		"cardnumber",
		"card_number",
		"cardno",
		"pan",
		"primaryaccountnumber",
		"number",
	}
}

// DefaultMaskedQueryParams returns the query parameter names masked by default.
func DefaultMaskedQueryParams() []string {
	return []string{
		"api-key",
		"api_key",
		"token",
		"access_token",
		"secret",
	}
}

func (l *LogConfig) setDefaults() {
	if l.Depth < LogDepthNone {
		l.Depth = LogDepthNone
	}
	if l.MaxBodyBytes <= 0 {
		l.MaxBodyBytes = DefaultLogBodyBytes
	}
	if l.MaskHeaders == nil {
		l.MaskHeaders = DefaultMaskedHeaders()
	}
	if l.MaskJSONFields == nil {
		l.MaskJSONFields = DefaultMaskedJSONFields()
	}
	if l.MaskQueryParams == nil {
		l.MaskQueryParams = DefaultMaskedQueryParams()
	}
}

// logEnabled reports whether a record at level should be emitted. Both the
// configured LogConfig.Level and the slog handler's own level are honoured, and
// LogDepthNone disables every record.
func (c *Client) logEnabled(ctx context.Context, level slog.Level) bool {
	logger := c.cfg.Log.Logger
	return logger != nil &&
		c.cfg.Log.Depth != LogDepthNone &&
		level >= c.cfg.Log.Level &&
		logger.Enabled(ctx, level)
}

func (c *Client) logResponse(
	ctx context.Context,
	req *http.Request,
	payload []byte,
	resp *http.Response,
	body []byte,
	duration time.Duration,
	attempt int,
	responseErr error,
) {
	level := slog.LevelInfo
	if resp.StatusCode >= http.StatusInternalServerError {
		level = slog.LevelWarn
	}
	if responseErr != nil {
		var apiErr *APIError
		if !errors.As(responseErr, &apiErr) || apiErr.StatusCode >= http.StatusInternalServerError {
			level = slog.LevelWarn
		}
	}
	if !c.logEnabled(ctx, level) {
		return
	}
	logger := c.cfg.Log.Logger

	attrs := []slog.Attr{
		slog.String("method", req.Method),
		slog.String("url", c.logURL(req.URL)),
		slog.Int("status", resp.StatusCode),
		slog.Duration("duration", duration),
		slog.Int("attempt", attempt+1),
	}
	if c.cfg.Log.Depth >= LogDepthHeaders {
		attrs = append(attrs,
			slog.Any("request_headers", c.maskedHeaders(req.Header)),
			slog.Any("response_headers", c.maskedHeaders(resp.Header)),
		)
	}
	if c.cfg.Log.Depth >= LogDepthBody {
		if len(payload) > 0 {
			attrs = append(attrs, slog.String("request_body", c.maskBody(payload)))
		}
		if len(body) > 0 {
			attrs = append(attrs, slog.String("response_body", c.maskBody(body)))
		}
	}
	if responseErr != nil {
		attrs = append(attrs, slog.String("error", c.maskedError(responseErr, req)))
	}

	logger.LogAttrs(ctx, level, "fincra request", attrs...)
}

// maskedError renders err for a log record with any URL or request URI replaced
// by the masked form. API and transport errors embed the request path, which may
// carry sensitive query parameters.
func (c *Client) maskedError(err error, req *http.Request) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.Path != "" {
		message = strings.ReplaceAll(message, apiErr.Path, c.maskPath(apiErr.Path))
	}
	if req == nil || req.URL == nil {
		return message
	}
	message = strings.ReplaceAll(message, req.URL.String(), c.logURL(req.URL))
	message = strings.ReplaceAll(message, req.URL.RequestURI(), c.maskPath(req.URL.RequestURI()))
	return message
}

func (c *Client) logTransportError(ctx context.Context, req *http.Request, err error, duration time.Duration, attempt int) {
	if !c.logEnabled(ctx, slog.LevelWarn) {
		return
	}

	// Go's transport errors embed the full URL. Replace it with the masked form
	// so sensitive query parameters never reach the log sink.
	message := c.maskedError(err, req)
	c.cfg.Log.Logger.LogAttrs(ctx, slog.LevelWarn, "fincra request failed",
		slog.String("method", req.Method),
		slog.String("url", c.logURL(req.URL)),
		slog.Duration("duration", duration),
		slog.Int("attempt", attempt+1),
		slog.String("error", message),
	)
}

func (c *Client) logRetry(ctx context.Context, method, path string, attempt int, delay time.Duration, reason string) {
	if !c.logEnabled(ctx, slog.LevelDebug) {
		return
	}

	c.cfg.Log.Logger.LogAttrs(ctx, slog.LevelDebug, "fincra retrying",
		slog.String("method", method),
		slog.String("url", c.maskPath(path)),
		slog.Int("attempt", attempt),
		slog.Duration("delay", delay),
		slog.String("reason", reason),
	)
}

func (c *Client) logURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	clone := *u
	clone.User = nil
	if clone.RawQuery != "" {
		query := clone.Query()
		for key := range query {
			if _, ok := c.maskQueryParams[strings.ToLower(key)]; ok {
				query.Set(key, "[REDACTED]")
			}
		}
		clone.RawQuery = query.Encode()
	}
	return clone.String()
}

func (c *Client) maskPath(path string) string {
	u, err := url.Parse(strings.TrimSpace(path))
	if err != nil {
		return path
	}
	if u.RawQuery != "" {
		query := u.Query()
		for key := range query {
			if _, ok := c.maskQueryParams[strings.ToLower(key)]; ok {
				query.Set(key, "[REDACTED]")
			}
		}
		u.RawQuery = query.Encode()
	}
	if u.IsAbs() {
		u.User = nil
		return u.String()
	}
	return u.RequestURI()
}

func (c *Client) maskedHeaders(headers http.Header) map[string][]string {
	masked := make(map[string][]string, len(headers))
	for key, values := range headers {
		if _, ok := c.maskHeaders[strings.ToLower(key)]; ok {
			masked[key] = []string{"[REDACTED]"}
			continue
		}
		masked[key] = append([]string(nil), values...)
	}
	return masked
}

func (c *Client) maskBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		// Never log an unstructured body: it may contain a form-encoded or
		// multipart secret that cannot be masked reliably.
		return fmt.Sprintf("<non-JSON body: %d bytes>", len(body))
	}

	masked, err := json.Marshal(maskJSONValue(decoded, c.maskJSONFields))
	if err != nil {
		return "<unloggable body>"
	}
	return truncateUTF8(string(masked), c.cfg.Log.MaxBodyBytes)
}

func maskJSONValue(value any, masked map[string]struct{}) any {
	switch v := value.(type) {
	case map[string]any:
		for key, child := range v {
			if _, ok := masked[strings.ToLower(key)]; ok {
				v[key] = "[REDACTED]"
				continue
			}
			v[key] = maskJSONValue(child, masked)
		}
		return v
	case []any:
		for i, child := range v {
			v[i] = maskJSONValue(child, masked)
		}
		return v
	default:
		return value
	}
}

func truncateUTF8(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	cut := max
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + "…(truncated)"
}

func newStringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != "" {
			set[value] = struct{}{}
		}
	}
	return set
}

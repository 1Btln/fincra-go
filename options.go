package fincra

import (
	"log/slog"
	"net/http"
	"time"
)

// Option customises a Config when passed to New or NewWithConfig. Options are
// applied in order after DefaultConfig (for New) or after the supplied Config
// (for NewWithConfig), so an option always wins.
type Option func(*Config) error

// WithSecretKey sets the api-key credential used to authenticate API calls.
func WithSecretKey(key string) Option {
	return func(c *Config) error {
		c.SecretKey = key
		return nil
	}
}

// WithPublicKey sets the x-pub-key header value used by endpoints such as
// checkout.
func WithPublicKey(key string) Option {
	return func(c *Config) error {
		c.PublicKey = key
		return nil
	}
}

// WithWebhookSecret sets the key used by Client.VerifyWebhookSignature.
func WithWebhookSecret(secret string) Option {
	return func(c *Config) error {
		c.WebhookSecret = secret
		return nil
	}
}

// WithBusinessID sets the default business or sub-account identifier.
func WithBusinessID(id string) Option {
	return func(c *Config) error {
		c.BusinessID = id
		return nil
	}
}

// WithEnvironment selects the Fincra environment.
func WithEnvironment(env Environment) Option {
	return func(c *Config) error {
		c.Environment = env
		return nil
	}
}

// WithEnvironmentName parses name and selects the matching environment.
func WithEnvironmentName(name string) Option {
	return func(c *Config) error {
		env, err := ParseEnvironment(name)
		if err != nil {
			return err
		}
		c.Environment = env
		return nil
	}
}

// WithBaseURL overrides the environment base URL. It is mainly useful for
// tests and proxies.
func WithBaseURL(baseURL string) Option {
	return func(c *Config) error {
		c.BaseURL = baseURL
		return nil
	}
}

// WithHTTPClient sets the HTTP client used to perform requests.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) error {
		c.HTTPClient = client
		return nil
	}
}

// WithTimeout bounds a logical API call, including retries.
func WithTimeout(d time.Duration) Option {
	return func(c *Config) error {
		c.Timeout = d
		return nil
	}
}

// WithMaxRetries sets the number of retries after the initial attempt. Zero
// disables retries.
func WithMaxRetries(n int) Option {
	return func(c *Config) error {
		c.MaxRetries = n
		return nil
	}
}

// WithRetryWait bounds the exponential backoff between retries.
func WithRetryWait(min, max time.Duration) Option {
	return func(c *Config) error {
		c.RetryWaitMin = min
		c.RetryWaitMax = max
		return nil
	}
}

// WithUserAgent overrides the User-Agent header.
func WithUserAgent(userAgent string) Option {
	return func(c *Config) error {
		c.UserAgent = userAgent
		return nil
	}
}

// WithLogger attaches a slog logger. Logging starts at the configured depth
// (LogDepthBasic by default); use WithLogDepth(LogDepthNone) to disable it.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Config) error {
		c.Log.Logger = logger
		return nil
	}
}

// WithLogConfig replaces the client's logging configuration.
func WithLogConfig(logConfig LogConfig) Option {
	return func(c *Config) error {
		c.Log = logConfig.Clone()
		return nil
	}
}

// WithLogLevel sets the minimum severity emitted by the SDK. The slog handler's
// own level is applied as well, so both must allow a record.
func WithLogLevel(level slog.Level) Option {
	return func(c *Config) error {
		c.Log.Level = level
		return nil
	}
}

// WithLogDepth selects how much request and response detail is logged. Use
// LogDepthNone to disable logging while keeping the logger configured.
func WithLogDepth(depth LogDepth) Option {
	return func(c *Config) error {
		c.Log.Depth = depth
		return nil
	}
}

// WithLogMaskHeaders adds header names whose values are masked. Matching is
// case-insensitive. Default masks are kept, so adding a custom header cannot
// accidentally expose credentials. Use WithLogConfig to replace the list.
func WithLogMaskHeaders(headers ...string) Option {
	return func(c *Config) error {
		c.Log.MaskHeaders = appendMask(c.Log.MaskHeaders, DefaultMaskedHeaders, headers)
		return nil
	}
}

// WithLogMaskJSONFields adds JSON object keys whose values are masked in logged
// bodies. Matching is case-insensitive. Default masks are kept, so adding a
// field cannot accidentally expose credentials. Use WithLogConfig to replace the
// list.
func WithLogMaskJSONFields(fields ...string) Option {
	return func(c *Config) error {
		c.Log.MaskJSONFields = appendMask(c.Log.MaskJSONFields, DefaultMaskedJSONFields, fields)
		return nil
	}
}

// WithLogMaskQueryParams adds query parameter names whose values are masked in
// logged URLs. Matching is case-insensitive. Default masks are kept, so adding a
// parameter cannot accidentally expose credentials. Use WithLogConfig to
// replace the list.
func WithLogMaskQueryParams(params ...string) Option {
	return func(c *Config) error {
		c.Log.MaskQueryParams = appendMask(c.Log.MaskQueryParams, DefaultMaskedQueryParams, params)
		return nil
	}
}

// appendMask extends a mask list without dropping secure defaults. A nil list
// starts from defaults; an explicitly empty list stays empty, which lets callers
// opt out through LogConfig.
func appendMask(existing []string, defaults func() []string, extra []string) []string {
	if len(extra) == 0 {
		return existing
	}
	base := existing
	if base == nil {
		base = defaults()
	}
	result := make([]string, 0, len(base)+len(extra))
	result = append(result, base...)
	return append(result, extra...)
}

// WithLogBodyLimit caps the request and response body text attached to a log
// record. Values <= 0 restore the default limit.
func WithLogBodyLimit(bytes int) Option {
	return func(c *Config) error {
		c.Log.MaxBodyBytes = bytes
		return nil
	}
}

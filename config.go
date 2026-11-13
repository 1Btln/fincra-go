package fincra

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Defaults applied by DefaultConfig, New, and NewWithConfig.
const (
	// DefaultTimeout bounds a single logical API call, including retries.
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the number of retries after the initial attempt.
	DefaultMaxRetries = 2

	// DefaultRetryWaitMin and DefaultRetryWaitMax bound the exponential
	// backoff delay between retries.
	DefaultRetryWaitMin = 500 * time.Millisecond
	DefaultRetryWaitMax = 5 * time.Second

	// DefaultUserAgent identifies this SDK to Fincra.
	DefaultUserAgent = "fincra-go/" + Version
)

// Environment variables read by ConfigFromEnv.
const (
	EnvSecretKey     = "FINCRA_SECRET_KEY"
	EnvPublicKey     = "FINCRA_PUBLIC_KEY"
	EnvWebhookSecret = "FINCRA_WEBHOOK_SECRET"
	EnvBusinessID    = "FINCRA_BUSINESS_ID"
	EnvEnvironment   = "FINCRA_ENVIRONMENT"
	EnvBaseURL       = "FINCRA_BASE_URL"
	EnvTimeout       = "FINCRA_TIMEOUT"
	EnvMaxRetries    = "FINCRA_MAX_RETRIES"
)

// Config controls a Client.
//
// A Config object is the primary source of settings. Construct a Config literal
// (or start from DefaultConfig), pass it to NewWithConfig, and set anything you
// need directly. Environment loading is an optional fallback: call ConfigFromEnv
// when there is no Config at all, or Config.ApplyEnv to fill only the fields you
// left unset.
//
// NewWithConfig copies the Config before applying defaults and options, so the
// caller's value is never mutated.
type Config struct {
	// SecretKey authenticates API calls through the "api-key" header. It grants
	// full access to the account and must never be exposed to untrusted code.
	// Required.
	SecretKey string

	// PublicKey identifies the business through the "x-pub-key" header. It is
	// only needed by endpoints that explicitly ask for it, such as checkout.
	// Safe to publish.
	PublicKey string

	// WebhookSecret verifies inbound webhook signatures. Fincra signs the raw
	// request body with HMAC-SHA512 using this key.
	WebhookSecret string

	// BusinessID is sent as the "x-business-id" header and used as the default
	// business for endpoints that accept a business identifier. For sub-account
	// integrations set the sub-account ID; otherwise use the parent business ID.
	BusinessID string

	// Environment selects the Fincra environment. Empty means sandbox.
	Environment Environment

	// BaseURL overrides the environment base URL. It is primarily useful for
	// tests, proxies, and local mocks. When empty, Environment.BaseURL is used.
	BaseURL string

	// HTTPClient performs requests. When nil a new http.Client is used. The
	// client is shared, not mutated, and its Transport is left untouched.
	HTTPClient *http.Client

	// Timeout bounds a logical call, including retries. Zero uses
	// DefaultTimeout. Cancellation of the supplied context also applies.
	Timeout time.Duration

	// MaxRetries is the number of retries after the initial attempt for
	// retryable failures. Zero disables retries.
	MaxRetries int

	// RetryWaitMin and RetryWaitMax bound exponential backoff. Zero values use
	// DefaultRetryWaitMin and DefaultRetryWaitMax.
	RetryWaitMin time.Duration
	RetryWaitMax time.Duration

	// UserAgent is sent with every request. Empty uses DefaultUserAgent.
	UserAgent string

	// Log controls structured request logging. Logging is disabled when
	// Log.Logger is nil or Log.Depth is LogDepthNone.
	Log LogConfig
}

// DefaultConfig returns a Config populated with safe defaults. The environment
// is sandbox and no credentials are set.
func DefaultConfig() *Config {
	return &Config{
		Environment:  EnvironmentSandbox,
		HTTPClient:   &http.Client{},
		Timeout:      DefaultTimeout,
		MaxRetries:   DefaultMaxRetries,
		RetryWaitMin: DefaultRetryWaitMin,
		RetryWaitMax: DefaultRetryWaitMax,
		UserAgent:    DefaultUserAgent,
		Log: LogConfig{
			Level:           slog.LevelInfo,
			Depth:           LogDepthBasic,
			MaskHeaders:     DefaultMaskedHeaders(),
			MaskJSONFields:  DefaultMaskedJSONFields(),
			MaskQueryParams: DefaultMaskedQueryParams(),
			MaxBodyBytes:    DefaultLogBodyBytes,
		},
	}
}

// ConfigFromEnv builds a Config from FINCRA_* environment variables, applies
// defaults, and validates the result.
//
// This is a fallback helper for programs that do not construct a Config. Prefer
// setting fields on Config directly and passing it to NewWithConfig; use
// Config.ApplyEnv when only some fields should come from the environment.
//
// Recognised variables are the Env* constants in this package. An error is
// returned when a variable is malformed or when the required secret key is
// missing.
func ConfigFromEnv() (*Config, error) {
	return configFromEnv(os.LookupEnv)
}

// ApplyEnv fills configuration fields that are still unset with values from
// FINCRA_* environment variables.
//
// It is a fallback helper: values already set on Config win, so environment
// variables only fill the gaps. Fields considered unset are empty strings, a
// zero Environment, a zero Timeout, and a zero MaxRetries. Apply defaults by
// passing the Config to NewWithConfig afterwards.
//
// ApplyEnv does not validate the result; NewWithConfig does that.
func (c *Config) ApplyEnv() error {
	if c == nil {
		return &ConfigError{Field: "Config", Reason: "must not be nil"}
	}
	return c.applyEnv(os.LookupEnv, false)
}

func configFromEnv(lookup func(string) (string, bool)) (*Config, error) {
	cfg := DefaultConfig()
	if err := cfg.applyEnv(lookup, true); err != nil {
		return nil, err
	}
	cfg.setDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// applyEnv reads FINCRA_* variables into c. When override is true, values from
// the environment replace existing fields; when false, only unset fields are
// filled, which is the fallback behaviour used by Config.ApplyEnv.
func (c *Config) applyEnv(lookup func(string) (string, bool), override bool) error {
	setString := func(name string, dst *string) {
		if !override && strings.TrimSpace(*dst) != "" {
			return
		}
		if v, ok := lookup(name); ok {
			*dst = strings.TrimSpace(v)
		}
	}
	setString(EnvSecretKey, &c.SecretKey)
	setString(EnvPublicKey, &c.PublicKey)
	setString(EnvWebhookSecret, &c.WebhookSecret)
	setString(EnvBusinessID, &c.BusinessID)
	setString(EnvBaseURL, &c.BaseURL)

	if override || c.Environment == "" {
		if v, ok := lookup(EnvEnvironment); ok && strings.TrimSpace(v) != "" {
			env, err := ParseEnvironment(v)
			if err != nil {
				return err
			}
			c.Environment = env
		}
	}

	if override || c.Timeout == 0 {
		if v, ok := lookup(EnvTimeout); ok && strings.TrimSpace(v) != "" {
			d, err := time.ParseDuration(strings.TrimSpace(v))
			if err != nil {
				return &ConfigError{Field: EnvTimeout, Value: v, Reason: "must be a Go duration such as 30s"}
			}
			c.Timeout = d
		}
	}

	if override || c.MaxRetries == 0 {
		if v, ok := lookup(EnvMaxRetries); ok && strings.TrimSpace(v) != "" {
			n, err := strconv.Atoi(strings.TrimSpace(v))
			if err != nil {
				return &ConfigError{Field: EnvMaxRetries, Value: v, Reason: "must be an integer"}
			}
			c.MaxRetries = n
		}
	}

	return nil
}

// Validate reports whether the configuration can be used to build a Client.
//
// SecretKey is required. Environment may be empty (sandbox). BaseURL, when set,
// must be an absolute http or https URL.
func (c *Config) Validate() error {
	if c == nil {
		return &ConfigError{Field: "Config", Reason: "must not be nil"}
	}
	if strings.TrimSpace(c.SecretKey) == "" {
		return &ConfigError{
			Field:  "SecretKey",
			Reason: "is required (set " + EnvSecretKey + " or use WithSecretKey)",
		}
	}
	if c.Environment != "" {
		if _, err := ParseEnvironment(string(c.Environment)); err != nil {
			return err
		}
	}
	if c.MaxRetries < 0 {
		return &ConfigError{Field: "MaxRetries", Value: strconv.Itoa(c.MaxRetries), Reason: "must not be negative"}
	}
	if c.Timeout < 0 {
		return &ConfigError{Field: "Timeout", Value: c.Timeout.String(), Reason: "must not be negative"}
	}
	if c.RetryWaitMin < 0 {
		return &ConfigError{Field: "RetryWaitMin", Value: c.RetryWaitMin.String(), Reason: "must not be negative"}
	}
	if c.RetryWaitMax < 0 {
		return &ConfigError{Field: "RetryWaitMax", Value: c.RetryWaitMax.String(), Reason: "must not be negative"}
	}
	if c.BaseURL != "" {
		u, err := url.Parse(c.BaseURL)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return &ConfigError{Field: "BaseURL", Value: c.BaseURL, Reason: "must be an absolute URL"}
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return &ConfigError{Field: "BaseURL", Value: c.BaseURL, Reason: `must use the "http" or "https" scheme`}
		}
	}
	return nil
}

// ResolveBaseURL returns the base URL requests are sent to. An explicit BaseURL
// wins; otherwise the environment URL is used.
func (c *Config) ResolveBaseURL() string {
	if c == nil {
		return SandboxBaseURL
	}
	if c.BaseURL != "" {
		return strings.TrimRight(c.BaseURL, "/")
	}
	return c.Environment.BaseURL()
}

// Clone returns a copy of the Config. The HTTPClient pointer is shared, while
// the logging mask lists are copied so later changes cannot affect the client.
func (c *Config) Clone() *Config {
	if c == nil {
		return nil
	}
	clone := *c
	clone.Log = c.Log.Clone()
	return &clone
}

// String implements fmt.Stringer with credentials redacted.
func (c Config) String() string {
	return fmt.Sprintf(
		"fincra.Config{Environment:%s BaseURL:%s BusinessID:%s SecretKey:%s PublicKey:%s WebhookSecret:%s Timeout:%s MaxRetries:%d LogLevel:%s LogDepth:%s}",
		c.Environment,
		c.ResolveBaseURL(),
		c.BusinessID,
		redact(c.SecretKey),
		redact(c.PublicKey),
		redact(c.WebhookSecret),
		c.Timeout,
		c.MaxRetries,
		c.Log.Level,
		c.Log.Depth,
	)
}

func (c *Config) setDefaults() {
	if c.Environment == "" {
		c.Environment = EnvironmentSandbox
	}
	if c.Timeout <= 0 {
		c.Timeout = DefaultTimeout
	}
	if c.RetryWaitMin <= 0 {
		c.RetryWaitMin = DefaultRetryWaitMin
	}
	if c.RetryWaitMax <= 0 {
		c.RetryWaitMax = DefaultRetryWaitMax
	}
	if c.RetryWaitMax < c.RetryWaitMin {
		c.RetryWaitMax = c.RetryWaitMin
	}
	if strings.TrimSpace(c.UserAgent) == "" {
		c.UserAgent = DefaultUserAgent
	}
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{}
	}
	c.Log.setDefaults()
	c.BaseURL = strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
}

func redact(s string) string {
	if s == "" {
		return "<unset>"
	}
	return "<redacted>"
}

// ConfigError describes an invalid configuration value. It deliberately never
// carries secret values.
type ConfigError struct {
	Field  string
	Value  string
	Reason string
}

// Error implements error.
func (e *ConfigError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("fincra: invalid %s %q: %s", e.Field, e.Value, e.Reason)
	}
	return fmt.Sprintf("fincra: invalid %s: %s", e.Field, e.Reason)
}

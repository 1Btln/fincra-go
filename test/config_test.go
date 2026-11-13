package fincra_test

import (
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	fincra "github.com/1Btln/fincra-go"
)

func TestDefaultConfig(t *testing.T) {
	cfg := fincra.DefaultConfig()

	if cfg.Environment != fincra.EnvironmentSandbox {
		t.Errorf("Environment = %q, want %q", cfg.Environment, fincra.EnvironmentSandbox)
	}
	if cfg.Timeout != fincra.DefaultTimeout {
		t.Errorf("Timeout = %s, want %s", cfg.Timeout, fincra.DefaultTimeout)
	}
	if cfg.MaxRetries != fincra.DefaultMaxRetries {
		t.Errorf("MaxRetries = %d, want %d", cfg.MaxRetries, fincra.DefaultMaxRetries)
	}
	if cfg.HTTPClient == nil {
		t.Error("HTTPClient is nil")
	}
	if got := cfg.ResolveBaseURL(); got != fincra.SandboxBaseURL {
		t.Errorf("ResolveBaseURL() = %q, want %q", got, fincra.SandboxBaseURL)
	}

	if cfg.Log.Logger != nil {
		t.Error("Log.Logger is set, want logging disabled by default")
	}
	if cfg.Log.Level != slog.LevelInfo {
		t.Errorf("Log.Level = %s, want %s", cfg.Log.Level, slog.LevelInfo)
	}
	if cfg.Log.Depth != fincra.LogDepthBasic {
		t.Errorf("Log.Depth = %s, want %s", cfg.Log.Depth, fincra.LogDepthBasic)
	}
	if cfg.Log.MaxBodyBytes != fincra.DefaultLogBodyBytes {
		t.Errorf("Log.MaxBodyBytes = %d, want %d", cfg.Log.MaxBodyBytes, fincra.DefaultLogBodyBytes)
	}
	if len(cfg.Log.MaskHeaders) == 0 || len(cfg.Log.MaskJSONFields) == 0 || len(cfg.Log.MaskQueryParams) == 0 {
		t.Errorf("Log mask lists = %+v, want defaults", cfg.Log)
	}
}

func TestNewUsesEnvironmentAsFallback(t *testing.T) {
	clearFincraEnv(t)
	t.Setenv(fincra.EnvSecretKey, "sk_env_fallback")
	t.Setenv(fincra.EnvBusinessID, "biz_env_fallback")
	t.Setenv(fincra.EnvEnvironment, "live")

	client, err := fincra.New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	cfg := client.Config()
	if cfg.SecretKey != "sk_env_fallback" {
		t.Errorf("SecretKey = %q, want environment fallback", cfg.SecretKey)
	}
	if cfg.BusinessID != "biz_env_fallback" {
		t.Errorf("BusinessID = %q, want environment fallback", cfg.BusinessID)
	}
	if cfg.Environment != fincra.EnvironmentLive {
		t.Errorf("Environment = %q, want live", cfg.Environment)
	}
}

func TestNewOptionsOverrideEnvironment(t *testing.T) {
	clearFincraEnv(t)
	t.Setenv(fincra.EnvSecretKey, "sk_env_fallback")
	t.Setenv(fincra.EnvBusinessID, "biz_env_fallback")
	t.Setenv(fincra.EnvEnvironment, "live")

	client, err := fincra.New(
		fincra.WithSecretKey("sk_explicit"),
		fincra.WithBusinessID("biz_explicit"),
		fincra.WithEnvironment(fincra.EnvironmentSandbox),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	cfg := client.Config()
	if cfg.SecretKey != "sk_explicit" || cfg.BusinessID != "biz_explicit" {
		t.Errorf("config = %+v, want explicit values", cfg)
	}
	if cfg.Environment != fincra.EnvironmentSandbox {
		t.Errorf("Environment = %q, want explicit sandbox", cfg.Environment)
	}
}

func TestParseEnvironment(t *testing.T) {
	tests := []struct {
		in      string
		want    fincra.Environment
		wantErr bool
	}{
		{in: "sandbox", want: fincra.EnvironmentSandbox},
		{in: "  SANDBOX ", want: fincra.EnvironmentSandbox},
		{in: "test", want: fincra.EnvironmentSandbox},
		{in: "live", want: fincra.EnvironmentLive},
		{in: "Production", want: fincra.EnvironmentLive},
		{in: "staging", wantErr: true},
		{in: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := fincra.ParseEnvironment(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseEnvironment(%q) = %q, want error", tt.in, got)
				}
				var cfgErr *fincra.ConfigError
				if !errors.As(err, &cfgErr) || cfgErr.Field != "Environment" {
					t.Fatalf("error = %v, want *fincra.ConfigError for Environment", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseEnvironment(%q): %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("ParseEnvironment(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestEnvironmentBaseURL(t *testing.T) {
	if got := fincra.EnvironmentSandbox.BaseURL(); got != fincra.SandboxBaseURL {
		t.Errorf("sandbox BaseURL() = %q, want %q", got, fincra.SandboxBaseURL)
	}
	if got := fincra.EnvironmentLive.BaseURL(); got != fincra.LiveBaseURL {
		t.Errorf("live BaseURL() = %q, want %q", got, fincra.LiveBaseURL)
	}
	if got := fincra.Environment("unknown").BaseURL(); got != fincra.SandboxBaseURL {
		t.Errorf("unknown BaseURL() = %q, want sandbox %q", got, fincra.SandboxBaseURL)
	}
}

func TestConfigValidate(t *testing.T) {
	valid := func() *fincra.Config {
		cfg := fincra.DefaultConfig()
		cfg.SecretKey = "sk_test_123"
		return cfg
	}

	tests := []struct {
		name      string
		mutate    func(*fincra.Config)
		wantField string
	}{
		{name: "valid", mutate: func(*fincra.Config) {}},
		{name: "missing secret key", mutate: func(c *fincra.Config) { c.SecretKey = "" }, wantField: "SecretKey"},
		{name: "whitespace secret key", mutate: func(c *fincra.Config) { c.SecretKey = "   " }, wantField: "SecretKey"},
		{name: "invalid environment", mutate: func(c *fincra.Config) { c.Environment = "staging" }, wantField: "Environment"},
		{name: "negative retries", mutate: func(c *fincra.Config) { c.MaxRetries = -1 }, wantField: "MaxRetries"},
		{name: "negative timeout", mutate: func(c *fincra.Config) { c.Timeout = -time.Second }, wantField: "Timeout"},
		{name: "negative retry wait", mutate: func(c *fincra.Config) { c.RetryWaitMin = -time.Second }, wantField: "RetryWaitMin"},
		{name: "relative base URL", mutate: func(c *fincra.Config) { c.BaseURL = "/api" }, wantField: "BaseURL"},
		{name: "unsupported base URL scheme", mutate: func(c *fincra.Config) { c.BaseURL = "ftp://example.com" }, wantField: "BaseURL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := valid()
			tt.mutate(cfg)

			err := cfg.Validate()
			if tt.wantField == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}

			var cfgErr *fincra.ConfigError
			if !errors.As(err, &cfgErr) {
				t.Fatalf("Validate() error = %v, want *fincra.ConfigError", err)
			}
			if cfgErr.Field != tt.wantField {
				t.Errorf("ConfigError.Field = %q, want %q", cfgErr.Field, tt.wantField)
			}
		})
	}
}

func TestConfigValidateNil(t *testing.T) {
	var cfg *fincra.Config
	var cfgErr *fincra.ConfigError
	if err := cfg.Validate(); !errors.As(err, &cfgErr) {
		t.Fatalf("Validate() error = %v, want *fincra.ConfigError", err)
	}
}

func clearFincraEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		fincra.EnvSecretKey,
		fincra.EnvPublicKey,
		fincra.EnvWebhookSecret,
		fincra.EnvBusinessID,
		fincra.EnvEnvironment,
		fincra.EnvBaseURL,
		fincra.EnvTimeout,
		fincra.EnvMaxRetries,
	} {
		t.Setenv(key, "")
	}
}

func TestConfigFromEnv(t *testing.T) {
	clearFincraEnv(t)
	t.Setenv(fincra.EnvSecretKey, "  sk_test_env ")
	t.Setenv(fincra.EnvPublicKey, "pk_test_env")
	t.Setenv(fincra.EnvWebhookSecret, "whsec_env")
	t.Setenv(fincra.EnvBusinessID, "biz_env")
	t.Setenv(fincra.EnvEnvironment, "live")
	t.Setenv(fincra.EnvBaseURL, "https://api.example.test/")
	t.Setenv(fincra.EnvTimeout, "12s")
	t.Setenv(fincra.EnvMaxRetries, "4")

	cfg, err := fincra.ConfigFromEnv()
	if err != nil {
		t.Fatalf("ConfigFromEnv: %v", err)
	}

	if cfg.SecretKey != "sk_test_env" {
		t.Errorf("SecretKey = %q, want trimmed value", cfg.SecretKey)
	}
	if cfg.Environment != fincra.EnvironmentLive {
		t.Errorf("Environment = %q, want live", cfg.Environment)
	}
	if cfg.BaseURL != "https://api.example.test" {
		t.Errorf("BaseURL = %q, want trailing slash trimmed", cfg.BaseURL)
	}
	if cfg.Timeout != 12*time.Second {
		t.Errorf("Timeout = %s, want 12s", cfg.Timeout)
	}
	if cfg.MaxRetries != 4 {
		t.Errorf("MaxRetries = %d, want 4", cfg.MaxRetries)
	}
}

func TestConfigFromEnvErrors(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]string
	}{
		{name: "missing secret key", values: map[string]string{}},
		{name: "invalid environment", values: map[string]string{fincra.EnvSecretKey: "sk", fincra.EnvEnvironment: "staging"}},
		{name: "invalid timeout", values: map[string]string{fincra.EnvSecretKey: "sk", fincra.EnvTimeout: "soon"}},
		{name: "invalid retries", values: map[string]string{fincra.EnvSecretKey: "sk", fincra.EnvMaxRetries: "many"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearFincraEnv(t)
			for key, value := range tt.values {
				t.Setenv(key, value)
			}
			if _, err := fincra.ConfigFromEnv(); err == nil {
				t.Fatal("ConfigFromEnv() error = nil, want error")
			}
		})
	}
}

func TestConfigApplyEnvFillsOnlyUnsetFields(t *testing.T) {
	clearFincraEnv(t)
	t.Setenv(fincra.EnvSecretKey, "sk_from_env")
	t.Setenv(fincra.EnvBusinessID, "biz_from_env")
	t.Setenv(fincra.EnvEnvironment, "live")
	t.Setenv(fincra.EnvTimeout, "45s")
	t.Setenv(fincra.EnvMaxRetries, "7")

	cfg := &fincra.Config{
		BusinessID: "biz_explicit",
		Timeout:    5 * time.Second,
	}
	if err := cfg.ApplyEnv(); err != nil {
		t.Fatalf("ApplyEnv: %v", err)
	}

	if cfg.SecretKey != "sk_from_env" {
		t.Errorf("SecretKey = %q, want environment fallback", cfg.SecretKey)
	}
	if cfg.BusinessID != "biz_explicit" {
		t.Errorf("BusinessID = %q, want explicit value preserved", cfg.BusinessID)
	}
	if cfg.Timeout != 5*time.Second {
		t.Errorf("Timeout = %s, want explicit value preserved", cfg.Timeout)
	}
	if cfg.Environment != fincra.EnvironmentLive {
		t.Errorf("Environment = %q, want environment fallback", cfg.Environment)
	}
	if cfg.MaxRetries != 7 {
		t.Errorf("MaxRetries = %d, want environment fallback", cfg.MaxRetries)
	}
}

func TestConfigApplyEnvNil(t *testing.T) {
	var cfg *fincra.Config
	var cfgErr *fincra.ConfigError
	if err := cfg.ApplyEnv(); !errors.As(err, &cfgErr) {
		t.Fatalf("ApplyEnv() error = %v, want *fincra.ConfigError", err)
	}
}

func TestConfigApplyEnvDoesNotOverrideDefaults(t *testing.T) {
	clearFincraEnv(t)
	t.Setenv(fincra.EnvEnvironment, "live")
	t.Setenv(fincra.EnvTimeout, "45s")

	cfg := fincra.DefaultConfig()
	if err := cfg.ApplyEnv(); err != nil {
		t.Fatalf("ApplyEnv: %v", err)
	}
	if cfg.Environment != fincra.EnvironmentSandbox {
		t.Errorf("Environment = %q, want default sandbox preserved", cfg.Environment)
	}
	if cfg.Timeout != fincra.DefaultTimeout {
		t.Errorf("Timeout = %s, want default timeout preserved", cfg.Timeout)
	}
}

func TestConfigClone(t *testing.T) {
	cfg := fincra.DefaultConfig()
	cfg.SecretKey = "sk_original"
	cfg.BusinessID = "biz_original"

	clone := cfg.Clone()
	clone.SecretKey = "sk_changed"
	clone.BusinessID = "biz_changed"

	if cfg.SecretKey != "sk_original" {
		t.Errorf("original SecretKey = %q, want unchanged", cfg.SecretKey)
	}
	if cfg.BusinessID != "biz_original" {
		t.Errorf("original BusinessID = %q, want unchanged", cfg.BusinessID)
	}
	if clone.HTTPClient != cfg.HTTPClient {
		t.Error("Clone() should share the HTTPClient pointer")
	}

	clone.Log.MaskHeaders[0] = "mutated"
	if cfg.Log.MaskHeaders[0] == "mutated" {
		t.Error("Clone() should copy Log mask slices")
	}
}

func TestConfigStringRedactsSecrets(t *testing.T) {
	cfg := fincra.DefaultConfig()
	cfg.SecretKey = "sk_live_super_secret"
	cfg.PublicKey = "pk_live_public"
	cfg.WebhookSecret = "whsec_live_super_secret"

	got := cfg.String()
	for _, secret := range []string{"sk_live_super_secret", "pk_live_public", "whsec_live_super_secret"} {
		if strings.Contains(got, secret) {
			t.Errorf("String() leaked %q: %s", secret, got)
		}
	}
	if !strings.Contains(got, fincra.EnvironmentSandbox.String()) {
		t.Errorf("String() = %q, want environment", got)
	}
}

func TestConfigResolveBaseURL(t *testing.T) {
	cfg := fincra.DefaultConfig()
	cfg.Environment = fincra.EnvironmentLive
	if got := cfg.ResolveBaseURL(); got != fincra.LiveBaseURL {
		t.Errorf("ResolveBaseURL() = %q, want %q", got, fincra.LiveBaseURL)
	}

	cfg.BaseURL = "https://proxy.example.test/"
	if got := cfg.ResolveBaseURL(); got != "https://proxy.example.test" {
		t.Errorf("ResolveBaseURL() = %q, want override without trailing slash", got)
	}
}

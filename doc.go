// Package fincra provides an idiomatic Go client for the Fincra payments API.
//
// The client supports the two Fincra environments: the sandbox used for testing
// (https://sandboxapi.fincra.com) and live production (https://api.fincra.com).
// Sandbox is the default so that a misconfigured client never moves real money.
//
// The primary way to configure a client is to build a Config and pass it to
// NewWithConfig. A Config object is authoritative and is never read from the
// environment implicitly:
//
//	cfg := fincra.DefaultConfig()
//	cfg.SecretKey = os.Getenv("FINCRA_SECRET_KEY")
//	cfg.BusinessID = "64f1c9ef838af8cda0211d9d"
//
//	client, err := fincra.NewWithConfig(cfg)
//	if err != nil {
//		return err
//	}
//
//	client.Balances.ListWallets(ctx, "")
//
// New is a convenience constructor that starts from DefaultConfig, fills unset
// fields from FINCRA_* environment variables, and then applies functional
// options. Options always win. Config.ApplyEnv is available when a Config is
// partly populated and only missing values should come from the environment;
// ConfigFromEnv builds a complete Config from the environment alone.
//
// Logging uses log/slog and is disabled by default. Configure Client.Log or use
// WithLogger and, optionally, WithLogDepth and WithLogLevel. Depth controls how
// much detail is attached (basic request information, headers, or bodies), while
// Level filters severity. Credentials and configured sensitive fields are masked
// at every depth.
//
// Fincra requires the originating server IP to be whitelisted in the dashboard.
// IP whitelisting is an account-level setting and cannot be configured here.
package fincra

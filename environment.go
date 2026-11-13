package fincra

import "strings"

// Environment identifies the Fincra API environment a Client talks to.
type Environment string

const (
	// EnvironmentSandbox is Fincra's isolated test environment. No real money
	// moves. Sandbox accounts, keys, and webhooks are separate from live.
	EnvironmentSandbox Environment = "sandbox"

	// EnvironmentLive is Fincra's production environment.
	EnvironmentLive Environment = "live"
)

// Base URLs for the Fincra environments.
const (
	SandboxBaseURL = "https://sandboxapi.fincra.com"
	LiveBaseURL    = "https://api.fincra.com"
)

// ParseEnvironment normalises an environment name into an Environment.
//
// It accepts "sandbox"/"test" for the sandbox and "live"/"production" for live.
// Matching is case-insensitive and surrounding whitespace is ignored.
func ParseEnvironment(s string) (Environment, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "sandbox", "test", "testing":
		return EnvironmentSandbox, nil
	case "live", "production", "prod":
		return EnvironmentLive, nil
	default:
		return "", &ConfigError{
			Field:  "Environment",
			Value:  s,
			Reason: `must be "sandbox" or "live"`,
		}
	}
}

// BaseURL returns the public API base URL for the environment.
//
// Unknown or empty environments resolve to the sandbox so that a zero-value
// configuration cannot accidentally target production.
func (e Environment) BaseURL() string {
	if e == EnvironmentLive {
		return LiveBaseURL
	}
	return SandboxBaseURL
}

// String implements fmt.Stringer.
func (e Environment) String() string { return string(e) }

# fincra-go

An idiomatic Go SDK for the [Fincra](https://docs.fincra.com) payments API.

> **Status:** active development. The SDK provides a reusable Go client and
> resource APIs; it does not orchestrate application-specific payment flows.

## Install

```sh
go get github.com/1Btln/fincra-go
```

Requires Go 1.24 or newer. The SDK has no third-party dependencies.

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	fincra "github.com/1Btln/fincra-go"
)

func main() {
	cfg := fincra.DefaultConfig()
	cfg.SecretKey = os.Getenv("FINCRA_SECRET_KEY")
	cfg.BusinessID = os.Getenv("FINCRA_BUSINESS_ID")

	client, err := fincra.NewWithConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	wallets, err := client.Balances.ListWallets(context.Background(), "")
	if err != nil {
		log.Fatal(err)
	}

	for _, wallet := range wallets.Data {
		fmt.Printf("%s: available=%f ledger=%f\n",
			wallet.Currency, wallet.AvailableBalance, wallet.LedgerBalance)
	}
}
```

The environment defaults to `sandbox`. Point the client at production
explicitly:

```go
cfg := fincra.DefaultConfig()
cfg.SecretKey = os.Getenv("FINCRA_SECRET_KEY")
cfg.Environment = fincra.EnvironmentLive

client, err := fincra.NewWithConfig(cfg)
```

## Configuration

A `Config` object is the primary source of settings. Start from
`fincra.DefaultConfig()` or a literal, set the fields you need, and pass it to
`fincra.NewWithConfig(cfg, opts...)`. `NewWithConfig` never reads the
environment, and the config is cloned before use, so the caller's value is never
modified.

```go
cfg := fincra.DefaultConfig()
cfg.SecretKey = os.Getenv("FINCRA_SECRET_KEY")
cfg.BusinessID = os.Getenv("FINCRA_BUSINESS_ID")
cfg.Timeout = 20 * time.Second

client, err := fincra.NewWithConfig(cfg)
```

Environment variables are a fallback, not the primary source:

- `fincra.New(opts...)` starts from `DefaultConfig`, then applies `FINCRA_*`
  variables, then options. This is convenient for small programs.
- `cfg.ApplyEnv()` fills only fields that are still unset (empty string, zero
  `Environment`, `Timeout`, or `MaxRetries`). Values already set on `Config`
  win.
- `fincra.ConfigFromEnv()` builds a complete `Config` from the environment. It
  validates the result and fails fast when the secret key is missing.

Precedence depends on the constructor:

- `NewWithConfig`: `Config` fields, then options, then defaults. No environment
  variables are read.
- `New`: defaults, then environment variables, then options.
- `ApplyEnv`: explicit `Config` fields win; environment variables only fill
  fields that are still unset.

| Field | Environment variable | Default | Purpose |
| --- | --- | --- | --- |
| `SecretKey` | `FINCRA_SECRET_KEY` | — (required) | Sent as `api-key`. Grants full account access. |
| `PublicKey` | `FINCRA_PUBLIC_KEY` | empty | Sent as `x-pub-key`. Needed by checkout endpoints. |
| `WebhookSecret` | `FINCRA_WEBHOOK_SECRET` | empty | Verifies inbound webhook signatures. |
| `BusinessID` | `FINCRA_BUSINESS_ID` | empty | Sent as `x-business-id`; default business/sub-account. |
| `Environment` | `FINCRA_ENVIRONMENT` | `sandbox` | `sandbox`/`test` or `live`/`production`. |
| `BaseURL` | `FINCRA_BASE_URL` | environment URL | Override for tests, proxies, or mocks. |
| `HTTPClient` | — | new `http.Client` | Shared, never mutated by the SDK. |
| `Timeout` | `FINCRA_TIMEOUT` | `30s` | Bounds one logical call including retries. |
| `MaxRetries` | `FINCRA_MAX_RETRIES` | `2` | Retries after the first attempt. |
| `RetryWaitMin`/`RetryWaitMax` | — | `500ms`/`5s` | Exponential backoff bounds. |
| `UserAgent` | — | `fincra-go/<version>` | Request identification. |
| `Log` | — | disabled | `LogConfig` for `slog` request logging. |

Loading from the environment validates the result and fails fast when the
secret key is missing:

```go
cfg, err := fincra.ConfigFromEnv()
if err != nil {
	log.Fatal(err)
}
client, err := fincra.NewWithConfig(cfg)
```

### Credentials

Fincra issues three keys per environment, and sandbox keys do not work in live:

- **Secret key** authenticates API requests (`api-key`).
- **Public key** identifies the business (`x-pub-key`).
- **Webhook secret** signs webhook payloads (HMAC-SHA512).

`Config.String()` redacts all three.

## Logging

Logging is built on `log/slog` and is disabled by default. Attach a logger and
choose how much detail to record:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelDebug,
}))

client, err := fincra.NewWithConfig(cfg,
	fincra.WithLogger(logger),
	fincra.WithLogLevel(slog.LevelInfo),
	fincra.WithLogDepth(fincra.LogDepthBasic),
)
```

`Depth` controls what is attached to a record; `Level` filters severity. Both
the SDK level and the `slog` handler level must allow a record.

| Depth | Value | Records include |
| --- | --- | --- |
| `LogDepthNone` | `-1` | Nothing, even with a logger configured. |
| `LogDepthBasic` | `0` | Method, URL, status, duration, attempt. |
| `LogDepthHeaders` | `1` | Request and response headers, masked. |
| `LogDepthBody` | `2` | Request and response bodies, masked and truncated. |

Credentials are masked at every depth. By default the SDK masks the `api-key`,
`x-pub-key`, authorization, and cookie headers; credential and personal
identifier JSON keys such as `password`, `token`, `bvn`, `nin`,
`accountNumber`, and `bankAccountNumber`; and sensitive query parameters such as
`token` and `secret`. Bodies that are not JSON are never logged in full. Bodies
are capped at 2048 bytes by default.

Customise masking or the body limit with
`WithLogMaskHeaders`, `WithLogMaskJSONFields`, `WithLogMaskQueryParams`, and
`WithLogBodyLimit`. The mask options add to the secure defaults; use
`WithLogConfig` (or the `Log` field on `Config`) to replace a list entirely. The
`Log` field accepts the same settings.

## Balances

```go
ctx := context.Background()

// All wallets for the configured business.
wallets, err := client.Balances.ListWallets(ctx, "")

// One wallet by Fincra wallet ID.
wallet, err := client.Balances.GetWallet(ctx, "105146")

// Credits and debits with pagination and filters.
history, err := client.Balances.GetBalanceHistory(ctx, &fincra.BalanceHistoryParams{
	Action:   "debit",
	DateFrom: "2024-01-01",
	DateTo:   "2024-10-10",
	PerPage:  50,
})

// A paginated statement for one currency.
statement, err := client.Balances.GetAccountStatement(ctx, &fincra.AccountStatementParams{
	DateFrom: "2024-01-01",
	DateTo:   "2024-10-31",
	Currency: "NGN",
})
```

Amounts are decoded as `float64` because Fincra mixes integers and decimals.
Convert to a decimal type before accounting or persisting balances.

## Webhooks

Fincra signs the **raw** request body with HMAC-SHA512 using the webhook secret
and sends the lowercase hex digest in the `signature` header. Verify before
decoding:

```go
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}

	if !client.VerifyWebhookSignature(payload, r.Header.Get(fincra.WebhookSignatureHeader)) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// Only now decode and process payload.
}
```

Re-marshalling a decoded struct changes key order and whitespace and will
invalidate the signature. Use `fincra.SignWebhookPayload` in tests and local
mocks.

## Errors

Every non-2xx response becomes an `*fincra.APIError`:

```go
wallets, err := client.Balances.ListWallets(ctx, businessID)
if err != nil {
	var apiErr *fincra.APIError
	if errors.As(err, &apiErr) {
		switch {
		case apiErr.IsUnauthorized():
			// bad or missing api-key
		case apiErr.IsForbidden():
			// IP not whitelisted, or action not permitted
		case apiErr.IsNotFound():
			// unknown resource
		case apiErr.IsRateLimited():
			// back off and retry later
		}
		log.Printf("fincra %s %s failed: %s (request %s)",
			apiErr.Method, apiErr.Path, apiErr.Message, apiErr.RequestID)
	}
	return err
}
```

Client-side argument problems wrap `ErrMissingRequiredParameter` or
`ErrInvalidParameter` and can be checked with `errors.Is`.

### Retries

The client retries network errors, `408`, `425`, `429`, `500`, `502`, `503`, and
`504` up to `MaxRetries` with exponential backoff and jitter. A `Retry-After`
header is honoured. `POST` requests are retried too, so operations that are not
safe to repeat should carry a caller-generated unique reference; Fincra rejects
duplicate customer references for payouts and charges.

Tune the policy globally:

```go
client, err := fincra.New(
	fincra.WithSecretKey(key),
	fincra.WithTimeout(20*time.Second),
	fincra.WithMaxRetries(0),
)
```

## Advanced usage

`Client.Do` is available for endpoints this SDK does not yet wrap. It applies
authentication, headers, retries, and error mapping:

```go
var out struct {
	Data []fincra.Wallet `json:"data"`
}
err := client.Do(ctx, http.MethodGet, "/wallets?businessID="+url.QueryEscape(id), nil, &out)
```

Resource services use an internal verb-oriented HTTP client backed by the same
request pipeline. Use `Client.Do` for requests that do not have a service method.

Per-request headers, such as a sub-account override, can be supplied with
`fincra.WithRequestHeader`.

For safety, absolute URLs are only accepted on the configured origin and
cross-host redirects are rejected, so the `api-key` header cannot be forwarded
to an unexpected host.

## Project layout

```
.
├── client.go          Public client and shared HTTP engine
├── config.go          Public configuration and environment fallback
├── options.go         Public functional options
├── environment.go     Public sandbox/live environment values
├── errors.go          Public API and validation error types
├── logging.go         Public slog settings and private logging helpers
├── webhook.go         Webhook signing and signature verification
├── webhook/           Event names, typed payloads, and signature helpers
├── account/           Business, banks, account resolution, BVN
├── balances/          Wallets, history, statements, top-ups
├── beneficiaries/     Saved-recipient management
├── checkout/          Hosted checkout and fee calculation
├── collections/       Deposits, refunds, chargebacks
├── conversions/       FX rates, quotes, wallet conversion
├── directcharge/      Card, bank-transfer, mobile-money, EFT charges
├── mandates/          Direct-debit mandates
├── payouts/           Disbursements and transfers
├── virtualaccounts/   Permanent, temporary, and BVN virtual-account lookup
├── internal/api/      Shared validation and request helpers
├── internal/httpclient/ Verb adapter used by resource services
├── test/              External (fincra_test) tests
└── .github/           CI workflow
```

New resources should follow the same shape: a `XxxService` field on `Client`,
typed request/response structs, methods that build a relative path, and a call
through the injected verb-based HTTP client.

## Roadmap

- [x] Client core, configuration, retries, errors, webhooks
- [x] Balances
- [x] Structured `slog` logging with configurable depth and masking
- [x] Business information and rates
- [x] Quotes and conversions
- [x] Checkout and direct charges
- [x] Virtual accounts and deposits
- [x] Payouts, beneficiaries, and transfers
- [x] Refunds, chargebacks, and mandates
- [x] Webhook event types and payload structs

Conversion listing and status lookup, BVN virtual-account lookup, charge OTP
resend, mandate listing, and typed stablecoin payout beneficiary fields are
available on their existing services. Virtual-account responses expose nested
account information, including CAD Interac aliases.

Webhook payloads can be decoded with `webhook.WebhookEnvelope[T]`, using the
matching payload type such as `webhook.PayoutWebhookData` or
`webhook.CollectionWebhookData` after verifying the signature against the
untouched raw request body. Root-level signature helpers remain available for
compatibility.

## Development

```sh
make check      # gofmt, vet, tests
make test-race  # race detector
make cover      # coverage report
```

All tests and runnable examples live in `test/` as the external `fincra_test`
package. Please add an `httptest` contract test for every endpoint, covering the
request path, query parameters, headers, success decoding, and API error mapping.

See [CONTRIBUTING.md](CONTRIBUTING.md) for the contribution workflow and
[SECURITY.md](SECURITY.md) for reporting vulnerabilities.

## Operational notes

- **IP whitelisting.** Fincra rejects requests from non-whitelisted server IPs
  with HTTP 403. Configure this in the dashboard; it cannot be done from the SDK.
- **Sandbox isolation.** Sandbox accounts, keys, balances, and webhook settings
  are separate from live and never carry over.
- **Balance history wire format.** Fincra documents
  `GET /wallets/logs` with a JSON request body. This SDK sends those fields as
  query parameters because GET bodies are not reliably forwarded by proxies. Use
  `Client.Do` directly if you need to mirror the documented form exactly.
- **Money.** Use decimal arithmetic for accounting; `float64` is only the
  transport representation.

## License

MIT. See [LICENSE](LICENSE).

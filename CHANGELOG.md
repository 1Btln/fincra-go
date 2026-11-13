# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-11-13

Initial public release.

### Added

- `Client` with authentication, JSON encoding and decoding, timeouts, retries
  with exponential backoff and jitter, `Retry-After` support, and a 10 MiB
  response limit.
- `Config` as the primary configuration source, with `DefaultConfig`,
  `Config.ApplyEnv`, and `ConfigFromEnv` helpers.
- Sandbox and live environments, with sandbox as the safe default.
- Structured logging through `log/slog`, configurable level and depth, and
  masking for credentials, personal identifiers, headers, query parameters, and
  JSON fields.
- Balances service: list wallets, get a wallet, balance history, and account
  statements.
- HMAC-SHA512 webhook signing and verification.
- `APIError` with status helpers, request IDs, and retry metadata.
- Origin checks for absolute URLs and cross-host redirect protection.
- External test suite in `test/` with `httptest` contract tests.

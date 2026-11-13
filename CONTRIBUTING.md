# Contributing to fincra-go

Thanks for taking the time to contribute. This project aims to be a small,
dependency-free, idiomatic Go client, so changes are reviewed with an emphasis
on correctness, security, and API stability.

## Development setup

You need Go 1.24 or newer.

```sh
git clone https://github.com/1Btln/fincra-go.git
cd fincra-go
make check
```

Useful targets:

```sh
make fmt        # gofmt
make vet        # go vet
make test       # unit and contract tests
make test-race  # race detector
make cover      # coverage across the library
```

## Tests

Tests live in `test/` and use the external `fincra_test` package. This keeps the
public API honest: tests can only use what a third-party caller can use.

- Put endpoint contract tests in `test/`. Use `httptest` and assert the request
  path, query parameters, headers, success decoding, and `*fincra.APIError`
  mapping.
- Keep runnable examples in `test/` alongside the other tests.
- Clear `FINCRA_*` variables with `t.Setenv` when a test could read them
  implicitly.

## Code style

- Run `gofmt` before submitting.
- Prefer the standard library. The SDK has no third-party dependencies.
- Keep exported identifiers documented.
- Do not log credentials, tokens, personal identifiers, or raw non-JSON bodies.
  Add new sensitive fields to the default mask lists in `logging.go`.
- Treat the sandbox as the default environment. Never make an operation target
  live unless the caller explicitly asks for it.

## Pull requests

1. Open an issue for substantial changes so the approach can be agreed first.
2. Keep each pull request focused and include tests.
3. Update `README.md` and `CHANGELOG.md` for user-visible changes.
4. Make sure `make check` and `go test -race ./...` pass.

## Commit messages

Use short imperative subjects, for example:

```
add payouts service
mask account numbers in request logs
```

## Reporting security issues

Do not open a public issue for a suspected vulnerability. Follow
[SECURITY.md](SECURITY.md).

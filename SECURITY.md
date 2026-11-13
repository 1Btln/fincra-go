# Security policy

## Supported versions

Security fixes are applied to the latest release. Until the project reaches
`v1.0.0`, the public API may still change between minor versions.

| Version | Supported |
| --- | --- |
| latest `0.x` | yes |

## Reporting a vulnerability

Please do not open a public issue for a security problem.

Report it privately through GitHub Security Advisories for
[`Ledgenta/fincra-go`](https://github.com/Ledgenta/fincra-go/security/advisories/new),
or contact the maintainers through the repository's private channels.

Include:

- the affected version or commit,
- a minimal reproduction or proof of concept,
- the impact you believe is possible,
- any suggested fix.

You can expect an acknowledgement within a few business days. We will keep you
informed while the issue is investigated and will credit you in the release notes
unless you prefer to remain anonymous.

## Scope

The SDK handles Fincra API credentials and webhook secrets. In-scope examples:

- leaking `api-key`, `x-pub-key`, or webhook secrets through logs, errors, or
  `Config.String()`,
- forwarding credentials to an unexpected host through redirects or absolute
  URLs,
- signature verification bypasses,
- decoding or reader bugs that expose memory or allow unbounded input.

Operational account issues such as IP whitelisting, dashboard access, or
compromised Fincra keys should be handled with Fincra support. This project
cannot rotate or revoke credentials.

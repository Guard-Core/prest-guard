# Security Policy

prest-guard is a security layer; reports about it are treated as
security-critical.

## Supported versions

| Version | Supported |
| ------- | --------- |
| 1.x | :white_check_mark: bug and security fixes |

## Reporting a vulnerability

Email security@guard-core.com or open a GitHub security advisory
(Security > Report a vulnerability). Please do not open a public issue for
anything that looks exploitable.

Include the plugin version, the guard-core-go version, the affected
configuration (env vars, with secrets redacted), and a reproduction.

## Scope notes

- The guard runs inside the prestd process. A plugin crash takes pREST
  down; engine panics are recovered by pREST's recovery middleware, but
  treat crash reports as in scope.
- Known accepted trade: content past PREST_GUARD_MAX_BODY_BYTES is not
  scanned and reaches the handler (documented in the README).

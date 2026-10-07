# Changelog

All notable changes to this project are documented here. The format follows
Keep a Changelog; versions follow SemVer.

v1.0.0 (2026-10-07)
-------------------

The first tagged release: the family Actions ruleset, the parity CI, and the guard-core-go v4.3.1 floor
-------------------------------------------------------------------------------------------------------

### Added

- **The guard-core family Actions ruleset** (`#1`): issue-link (PR bodies
  must deliver an open issue or carry `no-issue`), CodeQL Go analysis on
  every push/PR plus a weekly schedule, greetings, path-based labeler with a
  synced label inventory, scheduled lint (weekly go vet + staticcheck +
  govulncheck), stale, and on-demand AI issue summary. Brings prest-guard's
  CI surface in line with the rest of the family (nethttp-guard,
  guard-core-go, guard-core-ts, guard-agent-ts, guard-core-rs,
  guard-core-php).
- **Parity CI** (`#1`): the gofmt/vet/test/plugin-build matrix gains the
  race detector and govulncheck. Deliberately not adopted yet: docs.yml (no
  docs site), live-smoke.yml (needs a live pREST + Redis environment),
  release.yml and container-release.yml (releases are gated; they land with
  the first tagged release).

### Changed

- **The engine floor moves to `github.com/rennf93/guard-core-go/v4
  v4.3.1`**: prest-guard now requires the 4.3.1 engine release (live on the
  Go module proxy), picking up the engine's 4.3.1 fix train. The
  nethttp-guard floor stays at v1.3.0.

### Compatibility

- **No API change.** The plugin surface (`GuardMiddlewareLoad`, the
  `PREST_GUARD_*` configuration keys) is unchanged from 0.1.0; 1.0.0 marks
  production readiness with the family CI baseline and the current engine
  floor.

## [0.1.0] - 2026-10-05

Initial release.

### Added

- pREST middleware plugin (`GuardMiddlewareLoad`) wiring the guard-core-go
  v4.3.0 engine into pREST's CRUD stack via the #1039 plugin loader.
- Full `PREST_GUARD_*` configuration surface with raw-string validation:
  enabled, passive, rate limit (+window), payload inspection (off by
  default, category-restrictable), auto-ban (opt-in, threshold and
  duration), blacklist/whitelist/exempt IPs, exclude paths, trusted
  proxies, cloud provider blocking, body bound, Redis URL/prefix/fail-open.
- Fail-safe defaults: health and docs paths always excluded, bans opt-in,
  inspection opt-in, loud startup warnings (missing trusted proxies,
  config errors while disabled).
- Redis fail-open (per-instance fallback) and fail-closed (503 with
  background init retry and live swap-in on recovery) modes.
- Trusted-proxy client resolution: rightmost-untrusted X-Forwarded-For
  walk, multi-line header join, malformed-chain voiding, untrusted peers
  ignored.
- Fail-closed behavior for config errors while enabled, with marshaled
  generic JSON error bodies (details only in server logs).
- Tests for the plugin contract and realistic pREST traffic; benchmarks
  for inspection cost across body sizes and the rate-limit-only path.

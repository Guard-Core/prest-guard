# Changelog

All notable changes to this project are documented here. The format follows
Keep a Changelog; versions follow SemVer.

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

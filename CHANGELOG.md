# Changelog

All notable changes to this project are documented here. The format follows
Keep a Changelog; versions follow SemVer.

v1.1.0 (2026-10-16)
-------------------

The full engine config, engine-side exclusions and client identity, and the guard-core-go v4.4.0 floor
------------------------------------------------------------------------------------------------------

### Added

- **The plugin accepts the full engine config** (#17): `PREST_GUARD_ENGINE_CONFIG` (inline JSON) or `PREST_GUARD_ENGINE_CONFIG_FILE` decode a strict JSON object of `guardcore.SecurityConfig` field names over the existing env subset (unknown keys, bad JSON, func fields and values the engine rejects fail closed with the generic 500). The new `GuardConfig.EngineConfigJSON` field carries it, and `GuardConfig.Configure` gives a plugin built from a fork the fully built config last, for the func and interface fields JSON cannot carry (`CustomResponseModifier`, `AgentHandler`, `GeoIPHandler`, `CustomRequestCheck`). `PREST_GUARD_TRUSTED_PROXY_DEPTH` and `GuardConfig.TrustedProxyDepth` map the engine's `trusted_proxy_depth` (0 keeps the engine default of 1). The README documents the surface.
- **The containerized example app** (#8): a compose stack running prestd built from a pinned upstream commit with the prest-guard plugin compiled and mounted, in front of Postgres, published to GHCR.

### Changed

- **Raised the engine floor to `github.com/rennf93/guard-core-go/v4 v4.4.0`, the reference-alignment release.** The v4.4.0 engine adds the per-request scan time budget (`Detection.MaxScanSeconds`, `pattern_timeout`), honors `Detection.ThreatScoreThreshold`, canonicalizes client identity (IPv4-mapped addresses, brackets, zones), follows the reference IP list precedence, answers an unresolved client address with 403 under `FailSecure`, applies dynamic rule versions over the base config, fixes the filtered-value scan bypass (GHSA-w9j2-2m3v-5hcc) and scans the Cookie header one pair at a time. The `nethttp-guard` floor stays at v1.4.0 in this change and moves to the v1.5.0 release (which applies the `CustomResponseModifier`) in a follow-up once that tag is published.
- **Exclusions and client identity run inside the engine** (#17): `proxy.go` and the path short-circuit in the handler are gone. Excluded paths still skip detection but, as in the reference, `ip_security` and `rate_limit` stay enforced on them, so a probe source that shares an identity with a rate limit should go in `PREST_GUARD_EXEMPT_IPS`. Client resolution follows the engine's `trusted_proxy_depth` (default 1) instead of the old rightmost-untrusted walk. The pre-engine matcher stays only for the 503 window before an engine exists, and it now cleans the path so `/_health/../articles` cannot ride an excluded prefix.
- **Behavior to know before upgrading.** A request that reaches the engine with an empty remote address is now answered 403 `Client address could not be determined` under `FailSecure`. IPv4 blacklist, whitelist, exempt and emergency entries now also match clients that connect as `::ffff:a.b.c.d`, and an IPv4 CIDR no longer trusts a mapped peer as a proxy. A configured whitelist now decides alone (a route whitelist overrides the route blacklist and skips the global lists, a global whitelist overrides the global blacklist), so an entry that relied on the blacklist beating the whitelist must move. A dynamic rule version no longer stacks on the previous one. A request that spends `Detection.MaxScanSeconds` (default 10.0) of regex CPU time is blocked with a `pattern_timeout` threat, and a set `Detection.ThreatScoreThreshold` now changes verdicts. A passive-mode log match on the old penetration wording needs updating.

### Fixed

- **`PREST_GUARD_AUTO_BAN` maps to `EnableRateLimitAutoBan`** (#19): `buildEngineConfig` sets it together with `EnableIPBanning` (on when the variable is set, off in passive mode). The engine feeds rate-limit violations into the autoban counter from v4.4.0 (guard-core-go#100), so the knob now takes effect on this release's floor.
- **A timing flake in `TestDeferredHandlerSwapsLiveGuardIn`**: the background engine retry could win the race against the first request (about 2 runs in 12 on master and on v4.4.0), so the test now holds the retry until the 503 is observed.
- **Docs and CI**: the README is refreshed against the current engine (large-value truncation, benchmark table), the docs site deploys to GitHub Pages with `site_url` set, and the weekly lint runs on Go 1.26.x.

___

v1.0.1 (2026-10-09)
-------------------

The engine floor moves to guard-core-go v4.3.2 (the adapter-parity release)
---------------------------------------------------------------------------

### Changed

- **The engine floor moves to `github.com/rennf93/guard-core-go/v4
  v4.3.2`**: prest-guard now requires the 4.3.2 engine release (live on
  the Go module proxy), the adapter-parity release carrying the ReDoS
  static-safety trio (the pattern_safety corpus goes registry-free,
  94/94 with 0 divergences), the sus-patterns runtime registry with
  pattern events and real dynamic-rules application, the
  `custom_response_modifier` response pass, the websocket guard surface,
  the fifteen SecurityConfig knobs with the performance-monitor wiring,
  and the lifecycle/state surface (manager exports, `IsInitialized`, the
  cross-instance middleware state registry, `MarkInitialized`/
  `AgentStats`). The floor also brings `redis/go-redis/v9` v9.23.0
  transitively, clearing the stdlib-adjacent x/sys exposure;
  govulncheck stays clean. The `nethttp-guard` floor stays at v1.3.0.
- **Post-transfer metadata** (e6b262a): repo URLs, docs links, and
  ecosystem references point at the Guard-Core org. Module paths and Go
  imports are deliberately unchanged.

### Compatibility

- **No API change.** The plugin surface (`GuardMiddlewareLoad`, the
  `PREST_GUARD_*` configuration keys) is unchanged from 1.0.0; the 1.0.1
  floor bump tracks the engine release the plugin is tested against, per
  the repo's versioning rule (engine bumps follow the guard-core-go
  releases the plugin is tested against).

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

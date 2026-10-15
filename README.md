<p align="center">
    <a href="https://guard-core.github.io/guard-core/latest/">
        <img src="https://guard-core.github.io/guard-core/latest/assets/guard_core_legend.svg" alt="Guard Core">
    </a>
</p>

___

<p align="center">
    <strong>A [pREST](https://github.com/prest/prest) middleware plugin that puts the [guard-core](https://github.com/Guard-Core/guard-core-go) engine in front of pREST's CRUD routes: per-client rate limits, IP policy (blacklist, whitelist, exemptions), optional payload inspection across 19 attack categories, and opt-in auto-banning.</strong>
</p>

<p align="center">
    <a href="https://github.com/Guard-Core/prest-guard/releases">
        <img src="https://img.shields.io/github/v/tag/Guard-Core/prest-guard?label=release&color=0080ff" alt="Release tag">
    </a>
    <a href="https://opensource.org/licenses/MIT">
        <img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License">
    </a>
    <a href="https://github.com/Guard-Core/prest-guard/actions/workflows/ci.yml">
        <img src="https://github.com/Guard-Core/prest-guard/actions/workflows/ci.yml/badge.svg" alt="CI">
    </a>
    <a href="https://github.com/Guard-Core/prest-guard/actions/workflows/code-ql.yml">
        <img src="https://github.com/Guard-Core/prest-guard/actions/workflows/code-ql.yml/badge.svg" alt="CodeQL">
    </a>
</p>

<p align="center">
    <img src="https://img.shields.io/badge/Go-00ADD8.svg?style=flat&logo=go&logoColor=white" alt="Go"> <img src="https://img.shields.io/badge/PostgreSQL-4169E1.svg?style=flat&logo=postgresql&logoColor=white" alt="PostgreSQL">
</p>

<p align="center">
    <a href="https://guard-core.com">Website</a> &middot;
    <a href="https://playground.guard-core.com">Playground</a> &middot;
    <a href="https://app.guard-core.com">Dashboard</a> &middot;
    <a href="https://discord.gg/ZW7ZJbjMkK">Discord</a>
</p>

---



## Ecosystem

Guard Core is the Python engine. Framework adapters are thin wrappers that translate native request/response types into Guard Core's protocols. The telemetry agents ship security events and metrics to the monitoring backend. Parallel engine implementations exist for Go, PHP, TypeScript (on npm), and Rust (on crates.io) - all ports of the same reference semantics, conformance-tested against the shared adversarial corpus.

### Python

| Package | Role | PyPI |
|---|---|---|
| [guard-core](https://github.com/Guard-Core/guard-core) | Framework-agnostic security engine | [![PyPI](https://img.shields.io/pypi/v/guard-core)](https://pypi.org/project/guard-core/) |
| [guard-agent](https://github.com/Guard-Core/guard-agent) | Telemetry agent | [![PyPI](https://img.shields.io/pypi/v/guard-agent)](https://pypi.org/project/guard-agent/) |
| [fastapi-guard](https://github.com/Guard-Core/fastapi-guard) | FastAPI / Starlette adapter | [![PyPI](https://img.shields.io/pypi/v/fastapi-guard)](https://pypi.org/project/fastapi-guard/) |
| [flaskapi-guard](https://github.com/Guard-Core/flaskapi-guard) | Flask adapter | [![PyPI](https://img.shields.io/pypi/v/flaskapi-guard)](https://pypi.org/project/flaskapi-guard/) |
| [djapi-guard](https://github.com/Guard-Core/djapi-guard) | Django adapter | [![PyPI](https://img.shields.io/pypi/v/djapi-guard)](https://pypi.org/project/djapi-guard/) |
| [tornadoapi-guard](https://github.com/Guard-Core/tornadoapi-guard) | Tornado adapter | [![PyPI](https://img.shields.io/pypi/v/tornadoapi-guard)](https://pypi.org/project/tornadoapi-guard/) |

### Go

Go modules published via GitHub releases. **Production-ready.**

| Package | Role | Release |
|---|---|---|
| [guard-core-go](https://github.com/Guard-Core/guard-core-go) | Go engine | [![release](https://img.shields.io/github/v/tag/Guard-Core/guard-core-go?label=tag)](https://github.com/Guard-Core/guard-core-go/releases) |
| [nethttp-guard](https://github.com/Guard-Core/nethttp-guard) | net/http adapter | [![release](https://img.shields.io/github/v/tag/Guard-Core/nethttp-guard?label=tag)](https://github.com/Guard-Core/nethttp-guard/releases) |
| [gin-guard](https://github.com/Guard-Core/gin-guard) | Gin adapter | [![release](https://img.shields.io/github/v/tag/Guard-Core/gin-guard?label=tag)](https://github.com/Guard-Core/gin-guard/releases) |
| [echo-guard](https://github.com/Guard-Core/echo-guard) | Echo (v4) adapter | [![release](https://img.shields.io/github/v/tag/Guard-Core/echo-guard?label=tag)](https://github.com/Guard-Core/echo-guard/releases) |
| [fiber-guard](https://github.com/Guard-Core/fiber-guard) | Fiber (v3) adapter | [![release](https://img.shields.io/github/v/tag/Guard-Core/fiber-guard?label=tag)](https://github.com/Guard-Core/fiber-guard/releases) |
| [guard-agent-go](https://github.com/Guard-Core/guard-agent-go) | Telemetry agent | [![release](https://img.shields.io/github/v/tag/Guard-Core/guard-agent-go?label=tag)](https://github.com/Guard-Core/guard-agent-go/releases) |

### PHP

Published on [Packagist](https://packagist.org/) under the `rennf93` vendor. **Production-ready.**

| Package | Role | Packagist |
|---|---|---|
| [guard-core-php](https://github.com/Guard-Core/guard-core-php) | PHP engine | [![Packagist](https://img.shields.io/packagist/v/rennf93/guard-core-php)](https://packagist.org/packages/rennf93/guard-core-php) |
| [laravel-guard](https://github.com/Guard-Core/laravel-guard) | Laravel adapter | [![Packagist](https://img.shields.io/packagist/v/rennf93/laravel-guard)](https://packagist.org/packages/rennf93/laravel-guard) |
| [symfony-guard](https://github.com/Guard-Core/symfony-guard) | Symfony adapter | [![Packagist](https://img.shields.io/packagist/v/rennf93/symfony-guard)](https://packagist.org/packages/rennf93/symfony-guard) |
| [psr15-guard](https://github.com/Guard-Core/psr15-guard) | PSR-15 adapter | [![Packagist](https://img.shields.io/packagist/v/rennf93/psr15-guard)](https://packagist.org/packages/rennf93/psr15-guard) |
| [slim-guard](https://github.com/Guard-Core/slim-guard) | Slim 4 adapter | [![Packagist](https://img.shields.io/packagist/v/rennf93/slim-guard)](https://packagist.org/packages/rennf93/slim-guard) |
| [guard-agent-php](https://github.com/Guard-Core/guard-agent-php) | Telemetry agent | [![Packagist](https://img.shields.io/packagist/v/rennf93/guard-agent-php)](https://packagist.org/packages/rennf93/guard-agent-php) |

### TypeScript / JavaScript

Published under the [`@guardcore`](https://www.npmjs.com/org/guardcore) npm scope; source in the [guard-core-ts](https://github.com/Guard-Core/guard-core-ts) monorepo. **Production-ready.**

| Package | Role | npm |
|---|---|---|
| | [@guardcore/core](https://github.com/Guard-Core/guard-core-ts/tree/master/packages/core) | Core engine | [![npm](https://img.shields.io/npm/v/@guardcore%2Fcore)](https://www.npmjs.com/package/@guardcore/core) |
| [@guardcore/express](https://github.com/Guard-Core/guard-core-ts/tree/master/packages/express) | Express adapter | [![npm](https://img.shields.io/npm/v/@guardcore%2Fexpress)](https://www.npmjs.com/package/@guardcore/express) |
| [@guardcore/nestjs](https://github.com/Guard-Core/guard-core-ts/tree/master/packages/nestjs) | NestJS adapter | [![npm](https://img.shields.io/npm/v/@guardcore%2Fnestjs)](https://www.npmjs.com/package/@guardcore/nestjs) |
| [@guardcore/fastify](https://github.com/Guard-Core/guard-core-ts/tree/master/packages/fastify) | Fastify adapter | [![npm](https://img.shields.io/npm/v/@guardcore%2Ffastify)](https://www.npmjs.com/package/@guardcore/fastify) |
| [@guardcore/hono](https://github.com/Guard-Core/guard-core-ts/tree/master/packages/hono) | Hono (edge) adapter | [![npm](https://img.shields.io/npm/v/@guardcore%2Fhono)](https://www.npmjs.com/package/@guardcore/hono) |
| [guardagent](https://github.com/Guard-Core/guard-agent-ts) | Telemetry agent | [![npm](https://img.shields.io/npm/v/guardagent)](https://www.npmjs.com/package/guardagent) |

### Rust

Published on crates.io. **Production-ready.**

| Package | Role | crates.io |
|---|---|---|
| [guard-core-engine](https://github.com/Guard-Core/guard-core-rs) | Core engine crate | [![crates.io](https://img.shields.io/crates/v/guard-core-engine)](https://crates.io/crates/guard-core-engine) |
| [guard-core-rs](https://github.com/Guard-Core/guard-core-rs) | Facade crate (consumer entry point) | [![crates.io](https://img.shields.io/crates/v/guard-core-rs)](https://crates.io/crates/guard-core-rs) |
| [actix-guard-rs](https://github.com/Guard-Core/actix-guard-rs) | Actix Web adapter | [![crates.io](https://img.shields.io/crates/v/actix-guard-rs)](https://crates.io/crates/actix-guard-rs) |
| [axum-guard-rs](https://github.com/Guard-Core/axum-guard-rs) | Axum adapter | [![crates.io](https://img.shields.io/crates/v/axum-guard-rs)](https://crates.io/crates/axum-guard-rs) |
| [tower-guard-rs](https://github.com/Guard-Core/tower-guard-rs) | Tower adapter | [![crates.io](https://img.shields.io/crates/v/tower-guard-rs)](https://crates.io/crates/tower-guard-rs) |
| [rocket-guard-rs](https://github.com/Guard-Core/rocket-guard-rs) | Rocket adapter | [![crates.io](https://img.shields.io/crates/v/rocket-guard-rs)](https://crates.io/crates/rocket-guard-rs) |
| [guard-agent-rs](https://github.com/Guard-Core/guard-agent-rs) | Telemetry agent | [![crates.io](https://img.shields.io/crates/v/guard-agent-rs)](https://crates.io/crates/guard-agent-rs) |

### AI Coding Agents

| Package | Role | PyPI |
|---|---|---|
| [guard-core-mcp](https://github.com/Guard-Core/guard-core-mcp) | MCP server: config validation, docs search, detection sandbox | [![PyPI](https://img.shields.io/pypi/v/guard-core-mcp)](https://pypi.org/project/guard-core-mcp/) |

___

## Features

- **Rate limiting** per client IP with Redis-shared state
- **IP policy**: blacklist, whitelist, and exemptions
- **Payload inspection** across 19 attack categories, bounded by a measured default body budget
- **Opt-in auto-banning** with per-category thresholds
- **Strict fail-closed config**: every raw value validated, invalid config is a loud error
- **Passive preview mode**: log what would be blocked, reject nothing

___

## Documentation

📚 **[Documentation](https://guard-core.github.io/prest-guard/)** - full technical documentation for this package.

🛡️ **[Guard Core](https://guard-core.github.io/guard-core/latest/)** - the engine's reference documentation.

🤖 **[Monitoring Agent Integration](https://github.com/Guard-Core/guard-agent)** - monitor your Guard instance with a monitoring agent.
___

## Requirements and the native-plugin catch

pREST loads Go native plugins (`-buildmode=plugin`), and the Go runtime
requires a plugin to share its build identity with the host binary: build
prest-guard with the **same Go toolchain minor** prestd was built with
(1.26.x today) and keep shared dependencies at the same versions (the
plugin's `go.mod` pins negroni v3.1.1, matching prest main).

## Quick start

```bash
# 1. Build the plugin (CGO is required for -buildmode=plugin)
CGO_ENABLED=1 go build -buildmode=plugin -trimpath -ldflags "-s -w" \
  -o lib/middlewares/guard.so .

# 2. Register it in your prest config
#    [[pluginmiddlewarelist]]
#    file = "guard"
#    func = "Guard"

# 3. Turn it on
export PREST_GUARD_ENABLED=true
export PREST_GUARD_RATE_LIMIT=100
```

`file` and `func` produce the `GuardMiddlewareLoad` symbol the pREST plugin
loader looks up, from `lib/middlewares/guard.so`.

## Configuration

All settings come from `PREST_GUARD_*` environment variables (pREST passes
no configuration to middleware plugins). List values are comma-separated.
Every value is validated as its raw string: a typo like
`PREST_GUARD_RATE_LIMIT=100/min` is a loud config error, never a silent
disable. With the guard enabled, config errors fail closed (500 with a
generic JSON body); with it disabled they only log a warning.

| Variable | Default | Meaning |
|---|---|---|
| `PREST_GUARD_ENABLED` | `false` | Mount the guard at all |
| `PREST_GUARD_PASSIVE` | `false` | Log what would be blocked, reject nothing (rule preview) |
| `PREST_GUARD_RATE_LIMIT` | `0` | Requests per window per client IP; 0 disables |
| `PREST_GUARD_RATE_LIMIT_WINDOW` | `60` | Window seconds |
| `PREST_GUARD_INSPECT_PAYLOADS` | `false` | Payload inspection (query params, URL path, headers, bodies) |
| `PREST_GUARD_INSPECT_CATEGORIES` | all | Restrict inspection to categories (`sqli,xss,cmd_injection`, ...) |
| `PREST_GUARD_AUTO_BAN` | `false` | Ban IPs after repeated violations |
| `PREST_GUARD_AUTO_BAN_THRESHOLD` | `10` | Violations per category before a ban |
| `PREST_GUARD_AUTO_BAN_DURATION` | `3600` | Ban length in seconds |
| `PREST_GUARD_BLACKLIST` | empty | Blocked IPs/CIDRs (403) |
| `PREST_GUARD_WHITELIST` | empty | Allowed IPs/CIDRs; when set, everyone else is denied |
| `PREST_GUARD_EXEMPT_IPS` | empty | IPs/CIDRs that skip checks (still subject to active bans) |
| `PREST_GUARD_EXCLUDE_PATHS` | empty | Extra paths whose subtree the engine scopes as excluded (detection and policy checks skip; IP bans, IP policy and rate limits still run, see below) |
| `PREST_GUARD_TRUSTED_PROXIES` | empty | Proxy IPs/CIDRs whose `X-Forwarded-For` is trusted |
| `PREST_GUARD_TRUSTED_PROXY_DEPTH` | engine default (`1`) | How many forwarded hops from the right belong to your own proxies |
| `PREST_GUARD_BLOCK_CLOUD_PROVIDERS` | empty | Block datacenter ranges (`AWS,GCP,Azure`) |
| `PREST_GUARD_MAX_BODY_BYTES` | `8192` | Body prefix the inspection layer reads |
| `PREST_GUARD_REDIS_URL` | empty | Share rate limit and ban state across instances |
| `PREST_GUARD_REDIS_PREFIX` | `prest_guard` | Key namespace in Redis |
| `PREST_GUARD_REDIS_FAIL_OPEN` | `true` | Keep serving from per-instance memory when Redis is down |
| `PREST_GUARD_ENGINE_CONFIG` | empty | Inline JSON object over the full engine config (see below) |
| `PREST_GUARD_ENGINE_CONFIG_FILE` | empty | Path to a JSON file with the same content; set only one of the two |

### Full engine config

The variables above are a subset of the engine's `SecurityConfig`. Everything else (excluded detection headers, params and body fields, `Detection` tuning, `FailSecure`, endpoint rate limits, `EnforceHTTPS`, security headers, CORS, custom error responses, global behavior rules, blocked user agents, country lists, dynamic rules, emergency mode, agent settings, and the rest) is reachable through `PREST_GUARD_ENGINE_CONFIG` or `PREST_GUARD_ENGINE_CONFIG_FILE`: a JSON object whose keys are `guardcore.SecurityConfig` field names. It is decoded over the config the `PREST_GUARD_*` variables produce, so a key named in the JSON wins over the matching variable and the keys it does not name keep their values. Nested structs and maps merge into the engine defaults, slices and scalars replace them, durations are integer nanoseconds, and `ExcludePaths` can only add to the built-in defaults. Unknown keys, malformed JSON, func and interface fields, and values the engine rejects are loud config errors: an enabled guard answers 500 instead of running with a config you did not ask for.

```bash
export PREST_GUARD_ENGINE_CONFIG='{"BlockedUserAgents":["sqlmap"],"CustomErrorResponses":{"403":"blocked"},"ExcludedDetectionParams":{"_where":true}}'
```

Fields JSON cannot carry (`CustomResponseModifier`, `AgentHandler`, `GeoIPHandler`, `CustomRequestCheck`, `OnError`) are set from Go: a plugin built from a fork wraps `Load` with its own `GuardConfig.Configure`, which receives the fully built engine config last.

```go
func GuardMiddlewareLoad() negroni.Handler {
	return Load(func() GuardConfig {
		cfg := ParseGuardConfigFromEnv()
		cfg.Configure = func(c *guardcore.SecurityConfig) {
			c.CustomRequestCheck = myCheck
		}
		return cfg
	})
}
```

`CustomResponseModifier` takes effect once the pinned `nethttp-guard` release applies it (the nethttp adapter wires the engine's response modifier from the release after v1.4.0).

### Defaults that prevent outages

Three defaults exist specifically because most pREST deployments sit behind
a load balancer:

- `/_health`, `/_ready`, and the docs/static paths are always excluded,
  and operator entries in `EXCLUDE_PATHS` are merged with them, not
  replacing them. Exclusion is the engine's own scoping, the same as in the
  other adapters and the reference: detection and the route and request
  policy checks skip these paths, and their responses still carry the
  security headers and count in the metrics. IP bans, the IP blacklist and
  whitelist, and rate limits are still enforced on them. Health probes share
  the proxy's identity with all user traffic, so list the probe source in
  `EXEMPT_IPS` (and set `TRUSTED_PROXIES`) to keep a rate limit from taking
  pods out of rotation.
- Auto-banning is opt-in. With an empty `TRUSTED_PROXIES`, every client
  shares the load balancer's IP, so a handful of violations would ban all
  users at once. Starting the guard without `TRUSTED_PROXIES` logs a loud
  warning saying exactly that.
- Rate limiting works without Redis (per instance) and without payload
  inspection. The WAF is a separate switch.

### Trusted proxies

The engine resolves the client identity (the reference `extract_client_ip`), the plugin only hands it the connection's `RemoteAddr`. When the direct peer is in `TRUSTED_PROXIES`, the client is the hop `TRUSTED_PROXY_DEPTH` places from the right of `X-Forwarded-For` (default 1, the entry your proxy appended; set it to the number of proxies in front of pREST). A depth that over-counts the hops falls back to the right-to-left walk past trusted entries, and a chain shorter than the depth keeps the peer. Forwarded headers from untrusted peers are ignored, and the engine flags them as spoof attempts. Earlier versions walked the chain themselves and took the rightmost untrusted hop; with several trusted proxies in a row, set the depth instead.

### Redis modes

- `redis_fail_open = true` (default): if Redis is unreachable, the guard
  serves from per-instance memory (limits become per-process) and logs a
  warning. pREST stays up.
- `redis_fail_open = false`: requests answer 503 while initialization
  retries with backoff in the background, and the live guard swaps in when
  Redis returns. Health probes keep flowing either way. This replaces the
  older "500 forever after a failed startup" behavior with one that
  recovers.

The guard reads its own `PREST_GUARD_REDIS_URL` only. A `REDIS_URL` set for
pREST's own cache is never picked up by the guard.

### Coverage scope

pREST mounts plugin middlewares in the CRUD stack, after JWT auth. The
guard therefore covers the CRUD routes; the `/auth` endpoint, custom query
routes, the MCP endpoint, and the health endpoints are outside it (the
health endpoints by design). Global-stack coverage would be a small
follow-up in prest itself, and is the natural next step if wanted.

## Body scanning cost

Payload inspection is off by default, and body size bounds its cost. The
inspection layer reads at most `MAX_BODY_BYTES` of the body; content past
the bound reaches your handler unscanned, which is a deliberate trade, not
an oversight. Measured on an Apple M2 Pro (single request, guard enabled,
inspection on, full middleware path):

| Body size | Pass path (prose JSON) | Reject path (threat) |
|---|---|---|
| 8 KiB (default) | ~28 ms | ~20 ms |
| 16 KiB | ~47 ms | ~39 ms |
| 64 KiB | ~0.22 s | ~0.20 s |
| 256 KiB | ~1.2 s | ~1.2 s |

Cost still grows with body size, so keep the default unless you have measured acceptance for the cost at your bound. The engine caps what it scans (its large-value truncation, `Detection.MaxBodyInspectBytes` and the scan budget), which is why the 1 MiB bound costs about the same as 256 KiB. Treat `MAX_BODY_BYTES` as a CPU budget, not a feature dial. Rate limits and IP policy alone do not read bodies and cost about 9 microseconds per request.

## Verdict codes

| Situation | Status |
|---|---|
| Blacklisted IP, banned IP, non-whitelisted peer | 403 |
| Rate limit exceeded | 429 |
| Detected attack payload (inspection on) | 400 |
| Guard misconfigured while enabled | 500, generic JSON body |
| Redis down with `redis_fail_open = false` | 503, recovers automatically |

Blocked (and, in passive mode, would-have-blocked) requests are logged with
check name, reason, client IP, method, and path. Request headers, query
strings, and body content never reach the logs.

## Development

```bash
make fmt      # gofmt
make vet      # go vet
make test     # go test -race
make bench    # benchmarks (inspection cost table)
make so       # build lib/middlewares/guard.so
```

The conformance story lives in guard-core-go: the engine is CI-tested
against the same 448-case adversarial corpus as the Python reference
(byte-identical fixtures, verdicts generated by running the reference
engine). This repo tests the plugin contract: config parsing, engine
mapping, proxy resolution, exclusions, Redis modes, fail-closed/fail-open
behavior, and realistic pREST traffic passing with inspection on.

## License

MIT

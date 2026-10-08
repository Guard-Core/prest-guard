# prest-guard

A [pREST](https://github.com/prest/prest) middleware plugin that puts the
[guard-core](https://github.com/Guard-Core/guard-core-go) engine in front of
pREST's CRUD routes: per-client rate limits, IP policy (blacklist, whitelist,
exemptions), optional payload inspection across 19 attack categories, and
opt-in auto-banning. Everything is off by default; with no environment
variables set, pREST behaves exactly as before.

It ships as a pREST middleware plugin (see prest#1039) rather than a core
dependency, so none of its code or dependencies (regex engine, go-redis,
cloud-range data) enter the default prestd binary.

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
| `PREST_GUARD_EXCLUDE_PATHS` | empty | Extra paths whose subtree skips all guard checks |
| `PREST_GUARD_TRUSTED_PROXIES` | empty | Proxy IPs/CIDRs whose `X-Forwarded-For` is trusted |
| `PREST_GUARD_BLOCK_CLOUD_PROVIDERS` | empty | Block datacenter ranges (`AWS,GCP,Azure`) |
| `PREST_GUARD_MAX_BODY_BYTES` | `8192` | Body prefix the inspection layer reads |
| `PREST_GUARD_REDIS_URL` | empty | Share rate limit and ban state across instances |
| `PREST_GUARD_REDIS_PREFIX` | `prest_guard` | Key namespace in Redis |
| `PREST_GUARD_REDIS_FAIL_OPEN` | `true` | Keep serving from per-instance memory when Redis is down |

### Defaults that prevent outages

Three defaults exist specifically because most pREST deployments sit behind
a load balancer:

- `/_health`, `/_ready`, and the docs/static paths are excluded from every
  guard check, always. Health probes share the proxy's identity with all
  user traffic; rate limiting or banning them takes pods out of rotation.
  Operator entries in `EXCLUDE_PATHS` are merged with these, not replacing
  them.
- Auto-banning is opt-in. With an empty `TRUSTED_PROXIES`, every client
  shares the load balancer's IP, so a handful of violations would ban all
  users at once. Starting the guard without `TRUSTED_PROXIES` logs a loud
  warning saying exactly that.
- Rate limiting works without Redis (per instance) and without payload
  inspection. The WAF is a separate switch.

### Trusted proxies

When the direct peer is a trusted proxy, the client IP is resolved by
walking `X-Forwarded-For` from the right and taking the rightmost hop that
is not itself trusted (the leftmost entry when every hop is trusted, every
header line joined). Forwarded headers from untrusted peers are ignored
completely, and a malformed hop voids the whole chain. This mirrors the
reference engine's adapter-side contract; the engine separately flags
untrusted peers that send `X-Forwarded-For` as spoof attempts.

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
| 8 KiB (default) | ~59 ms | ~59 ms |
| 16 KiB | ~140 ms | ~139 ms |
| 64 KiB | ~1.16 s | ~1.25 s |
| 256 KiB | ~13 s | ~16 s |

Rejection is not cheaper than passing; cost is body-size driven. Keep the
default unless you have measured acceptance for the cost at your bound:
eight concurrent clients posting 64 KiB bodies will saturate one core.

The numbers above are large because guard-core-go does not yet port the
reference engine's large-value truncation (Python guard-core caps
full-pattern scans at 256 KiB and rebuilds a bounded scan budget around
detected attack regions). Until that lands engine-side, treat
`MAX_BODY_BYTES` as a CPU budget, not a feature dial. Rate limits and IP
policy alone do not read bodies and cost about 4 microseconds per request.

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

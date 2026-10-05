// prest-guard is a pREST middleware plugin that puts the guard-core engine
// (per-client rate limits, IP policy, optional payload inspection) in front
// of pREST's CRUD routes.
//
// The plugin exports the GuardMiddlewareLoad symbol that pREST's plugin
// loader looks up for a [[pluginmiddlewarelist]] entry with file = "guard"
// and func = "Guard". pREST passes no configuration to middleware plugins,
// so every setting is read from PREST_GUARD_* environment variables.
package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/rennf93/guard-core-go/v4/guardcore"
)

const (
	envPrefix = "PREST_GUARD_"

	defaultRateLimitWindow   = 60
	defaultMaxBodyBytes      = 8 << 10 // 8 KiB; see README "body scanning cost"
	defaultRedisPrefix       = "prest_guard"
	defaultAutoBanThreshold  = 10
	defaultAutoBanDuration   = 3600 // seconds
	defaultRedisInitFailOpen = true
)

// defaultExcludedPaths are the paths the guard skips even when the operator
// lists none: the health probes a load balancer fans out, plus the engine's
// documentation and static defaults. Rate limiting or banning a probe whose
// identity is shared by every client behind a proxy is an outage waiting to
// happen.
var defaultExcludedPaths = append([]string{"/_health", "/_ready"}, guardcore.DefaultExcludePaths...)

// GuardConfig carries every PREST_GUARD_* setting. Values that fail
// validation keep their defaults and the error is recorded in Invalid: a
// typo must never silently disable a control the operator asked for.
type GuardConfig struct {
	// Enabled turns the guard on. Default false.
	Enabled bool
	// EnabledInvalid marks a PREST_GUARD_ENABLED value that is set but not a
	// valid boolean. The operator clearly reached for the switch, so the
	// middleware fails closed even though Enabled parsed as false.
	EnabledInvalid bool
	// Invalid carries the first config error found while the guard is meant
	// to be active. Nil means the config parsed fine.
	Invalid error

	// Passive logs what the guard would have blocked instead of rejecting.
	Passive bool
	// InspectPayloads enables the detection engine (query params, URL path,
	// headers, and request bodies) across its 19 attack categories. Default
	// false: rate limits and IP policy work without the WAF.
	InspectPayloads bool
	// InspectCategories restricts detection to the named categories (for
	// example "sqli,xss,cmd_injection"). Empty means all categories.
	InspectCategories []string
	// AutoBan enables automatic IP bans after repeated violations. Default
	// false.
	AutoBan bool

	// RateLimit is the maximum number of requests per RateLimitWindow
	// seconds per client IP. 0 (default) disables rate limiting.
	RateLimit int
	// RateLimitWindow is the rate limit window in seconds. Default 60.
	RateLimitWindow int
	// AutoBanThreshold is the number of violations per category that
	// triggers a ban. Default 10.
	AutoBanThreshold int
	// AutoBanDuration is the ban duration in seconds. Default 3600.
	AutoBanDuration int
	// MaxBodyBytes caps how much of a request body the inspection layer
	// reads. Default 8 KiB. Content past the bound reaches the handler
	// unscanned, and detection cost scales with this bound (see the README's
	// body-scanning-cost table before raising it).
	MaxBodyBytes int64

	// Blacklist blocks these IPs/CIDRs; Whitelist allows them unconditionally.
	Blacklist []string
	Whitelist []string
	// ExemptIPs skips guard checks for these IPs/CIDRs (maintenance probes,
	// internal services). Unlike whitelist entries they are still subject to
	// active bans.
	ExemptIPs []string
	// ExcludePaths are operator-provided paths to skip guard checks for,
	// merged with DefaultExcludedPaths at use time.
	ExcludePaths []string
	// TrustedProxies lists proxy IPs/CIDRs whose X-Forwarded-For is trusted
	// when resolving the real client IP. Empty means forwarded headers are
	// ignored and the direct peer is the client.
	TrustedProxies []string
	// BlockCloudProviders blocks datacenter ranges ("AWS", "GCP", "Azure").
	BlockCloudProviders []string

	// RedisURL shares rate limit and ban state across instances when set.
	RedisURL string
	// RedisPrefix namespaces guard keys in Redis. Default "prest_guard".
	RedisPrefix string
	// RedisFailOpen keeps pREST serving from per-instance memory when Redis
	// is unreachable. Default true. With false, requests answer 503 while
	// initialization retries in the background, and the guard goes live when
	// Redis returns.
	RedisFailOpen bool
}

// ParseGuardConfig reads every PREST_GUARD_* variable through lookup. Each
// value is validated in its raw string form so that viper-style silent
// conversions ("100/min" becoming 0, "yes" becoming false) cannot happen.
func ParseGuardConfig(lookup func(string) (string, bool)) GuardConfig {
	cfg := GuardConfig{
		RateLimitWindow:  defaultRateLimitWindow,
		MaxBodyBytes:     defaultMaxBodyBytes,
		RedisPrefix:      defaultRedisPrefix,
		RedisFailOpen:    defaultRedisInitFailOpen,
		AutoBanThreshold: defaultAutoBanThreshold,
		AutoBanDuration:  defaultAutoBanDuration,
	}

	var errs []string
	fail := func(format string, args ...any) {
		errs = append(errs, fmt.Sprintf(format, args...))
	}

	if raw, ok := lookup(envPrefix + "ENABLED"); ok {
		parsed, err := parseBool(raw)
		if err != nil {
			cfg.EnabledInvalid = true
			fail("PREST_GUARD_ENABLED: %q is not a valid boolean", raw)
		} else {
			cfg.Enabled = parsed
		}
	}

	if raw, ok := lookup(envPrefix + "PASSIVE"); ok {
		if parsed, err := parseBool(raw); err != nil {
			fail("PREST_GUARD_PASSIVE: %q is not a valid boolean", raw)
		} else {
			cfg.Passive = parsed
		}
	}

	if raw, ok := lookup(envPrefix + "INSPECT_PAYLOADS"); ok {
		if parsed, err := parseBool(raw); err != nil {
			fail("PREST_GUARD_INSPECT_PAYLOADS: %q is not a valid boolean", raw)
		} else {
			cfg.InspectPayloads = parsed
		}
	}

	if raw, ok := lookup(envPrefix + "AUTO_BAN"); ok {
		if parsed, err := parseBool(raw); err != nil {
			fail("PREST_GUARD_AUTO_BAN: %q is not a valid boolean", raw)
		} else {
			cfg.AutoBan = parsed
		}
	}

	if raw, ok := lookup(envPrefix + "RATE_LIMIT"); ok {
		if parsed, err := parseInt(raw); err != nil {
			fail("PREST_GUARD_RATE_LIMIT: %q is not a valid integer", raw)
		} else if parsed < 0 {
			fail("PREST_GUARD_RATE_LIMIT: %d is negative; use 0 to disable rate limiting", parsed)
		} else {
			cfg.RateLimit = parsed
		}
	}

	if raw, ok := lookup(envPrefix + "RATE_LIMIT_WINDOW"); ok {
		if parsed, err := parseInt(raw); err != nil {
			fail("PREST_GUARD_RATE_LIMIT_WINDOW: %q is not a valid integer", raw)
		} else if parsed <= 0 {
			fail("PREST_GUARD_RATE_LIMIT_WINDOW: %d is not positive", parsed)
		} else {
			cfg.RateLimitWindow = parsed
		}
	}

	if raw, ok := lookup(envPrefix + "AUTO_BAN_THRESHOLD"); ok {
		if parsed, err := parseInt(raw); err != nil {
			fail("PREST_GUARD_AUTO_BAN_THRESHOLD: %q is not a valid integer", raw)
		} else if parsed <= 0 {
			fail("PREST_GUARD_AUTO_BAN_THRESHOLD: %d is not positive", parsed)
		} else {
			cfg.AutoBanThreshold = parsed
		}
	}

	if raw, ok := lookup(envPrefix + "AUTO_BAN_DURATION"); ok {
		if parsed, err := parseInt(raw); err != nil {
			fail("PREST_GUARD_AUTO_BAN_DURATION: %q is not a valid integer", raw)
		} else if parsed <= 0 {
			fail("PREST_GUARD_AUTO_BAN_DURATION: %d is not positive", parsed)
		} else {
			cfg.AutoBanDuration = parsed
		}
	}

	if raw, ok := lookup(envPrefix + "MAX_BODY_BYTES"); ok {
		if parsed, err := parseInt(raw); err != nil {
			fail("PREST_GUARD_MAX_BODY_BYTES: %q is not a valid integer", raw)
		} else if parsed <= 0 {
			fail("PREST_GUARD_MAX_BODY_BYTES: %d is not positive", parsed)
		} else {
			cfg.MaxBodyBytes = int64(parsed)
		}
	}

	if raw, ok := lookup(envPrefix + "BLACKLIST"); ok {
		cfg.Blacklist = splitList(raw)
	}
	if raw, ok := lookup(envPrefix + "WHITELIST"); ok {
		cfg.Whitelist = splitList(raw)
	}
	if raw, ok := lookup(envPrefix + "EXEMPT_IPS"); ok {
		cfg.ExemptIPs = splitList(raw)
	}
	if raw, ok := lookup(envPrefix + "INSPECT_CATEGORIES"); ok {
		cfg.InspectCategories = splitList(raw)
	}
	if raw, ok := lookup(envPrefix + "EXCLUDE_PATHS"); ok {
		cfg.ExcludePaths = splitList(raw)
	}
	if raw, ok := lookup(envPrefix + "TRUSTED_PROXIES"); ok {
		cfg.TrustedProxies = splitList(raw)
	}
	if raw, ok := lookup(envPrefix + "BLOCK_CLOUD_PROVIDERS"); ok {
		cfg.BlockCloudProviders = splitList(raw)
	}

	if raw, ok := lookup(envPrefix + "REDIS_URL"); ok {
		cfg.RedisURL = strings.TrimSpace(raw)
	}
	if raw, ok := lookup(envPrefix + "REDIS_PREFIX"); ok {
		if trimmed := strings.TrimSpace(raw); trimmed != "" {
			cfg.RedisPrefix = trimmed
		} else {
			cfg.RedisPrefix = defaultRedisPrefix
		}
	}
	if raw, ok := lookup(envPrefix + "REDIS_FAIL_OPEN"); ok {
		if parsed, err := parseBool(raw); err != nil {
			fail("PREST_GUARD_REDIS_FAIL_OPEN: %q is not a valid boolean", raw)
		} else {
			cfg.RedisFailOpen = parsed
		}
	}

	if len(errs) > 0 {
		cfg.Invalid = fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return cfg
}

// active reports whether the guard is meant to run: either it is explicitly
// enabled or the enable switch itself failed to parse. Only an active guard
// escalates config errors to fail-closed behavior.
func (c GuardConfig) active() bool {
	return c.Enabled || c.EnabledInvalid
}

// ParseGuardConfigFromEnv is ParseGuardConfig against the process
// environment.
func ParseGuardConfigFromEnv() GuardConfig {
	return ParseGuardConfig(os.LookupEnv)
}

func parseBool(raw string) (bool, error) {
	return strconv.ParseBool(strings.TrimSpace(raw))
}

func parseInt(raw string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(raw))
}

// splitList splits a comma-separated environment value. pREST passes no
// config to plugins, so there is no native list type here: commas are the
// documented separator, entries are trimmed and empties dropped.
func splitList(raw string) []string {
	parts := strings.Split(raw, ",")
	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			kept = append(kept, trimmed)
		}
	}
	return kept
}

// warnStartupProblems logs the loud warnings an operator should see before
// an incident, not after: typos in a disabled guard, and the proxy-blind
// default that makes every client behind a load balancer share one identity.
func (c GuardConfig) warnStartupProblems() {
	if c.Invalid != nil && !c.active() {
		slog.Warn("prest-guard: guard is disabled but its configuration has errors; "+
			"fix them before enabling", "err", c.Invalid)
	}
	if c.active() && len(c.TrustedProxies) == 0 {
		slog.Warn("prest-guard: enabled with no trusted proxies; behind a load balancer " +
			"every client shares the proxy's IP, so rate limits and bans apply to all " +
			"users at once. List your proxy in PREST_GUARD_TRUSTED_PROXIES")
	}
	if c.active() && c.Passive {
		slog.Info("prest-guard: passive mode is on; the guard logs what it would block and rejects nothing")
	}
}

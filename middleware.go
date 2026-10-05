package main

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	guardcore "github.com/rennf93/guard-core-go/v4/guardcore"
	nethttpguard "github.com/rennf93/nethttp-guard"
	"github.com/urfave/negroni/v3"
)

// GuardMiddlewareLoad is the symbol pREST's plugin loader looks up for a
// [[pluginmiddlewarelist]] entry with file = "guard" and func = "Guard".
func GuardMiddlewareLoad() negroni.Handler {
	return Load(ParseGuardConfigFromEnv)
}

// Load builds the guard middleware from PREST_GUARD_* configuration. It
// never returns nil (the pREST loader rejects nil handlers):
//
//   - guard disabled: a pass-through;
//   - config errors while the guard is meant to be active: a handler that
//     answers 500 with a generic JSON body, because a security layer the
//     operator enabled must not silently degrade to absent (the specific
//     error goes to the server log);
//   - engine initialization failure, such as Redis unreachable under
//     redis_fail_open = false: a handler that answers 503 while
//     initialization retries with backoff in the background, swapping in the
//     live guard when it succeeds, so the outage recovers instead of
//     latching.
func Load(parse func() GuardConfig) negroni.Handler {
	cfg := parse()
	cfg.warnStartupProblems()

	if cfg.Invalid != nil && cfg.active() {
		slog.Error("prest-guard: invalid config, failing closed", "err", cfg.Invalid)
		return misconfiguredHandler()
	}
	if !cfg.Enabled {
		return passThroughHandler()
	}

	trusted, err := newTrustedProxySet(cfg.TrustedProxies)
	if err != nil {
		slog.Error("prest-guard: invalid config, failing closed", "err", err)
		return misconfiguredHandler()
	}

	engineCfg, err := buildEngineConfig(cfg)
	if err != nil {
		slog.Error("prest-guard: invalid config, failing closed", "err", err)
		return misconfiguredHandler()
	}

	exclusions := newPathExclusions(mergeExclusions(cfg.ExcludePaths))

	handler, err := engineFactoryFor(engineCfg, cfg, exclusions, trusted)()
	if err == nil {
		return handler
	}
	slog.Error("prest-guard: engine initialization failed; answering 503 until it succeeds", "err", err)
	return deferredHandler(
		engineFactoryFor(engineCfg, cfg, exclusions, trusted),
		exclusions,
		time.Second,
	)
}

// buildEngineConfig maps GuardConfig onto the engine's SecurityConfig.
// Non-positive values coerce to the documented defaults so a hand-built
// GuardConfig (tests, embedding) cannot trip engine validation on a field
// the parser would have defaulted.
func buildEngineConfig(cfg GuardConfig) (*guardcore.SecurityConfig, error) {
	if cfg.RateLimitWindow <= 0 {
		cfg.RateLimitWindow = defaultRateLimitWindow
	}
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = defaultMaxBodyBytes
	}
	if cfg.AutoBan {
		if cfg.AutoBanThreshold <= 0 {
			cfg.AutoBanThreshold = defaultAutoBanThreshold
		}
		if cfg.AutoBanDuration <= 0 {
			cfg.AutoBanDuration = defaultAutoBanDuration
		}
	}
	return guardcore.NewSecurityConfig(func(c *guardcore.SecurityConfig) {
		c.OnBlock = onBlockLogger()
		c.PassiveMode = cfg.Passive

		// EnableRateLimiting is the switch; the engine's RateLimit field
		// must stay at its valid default even when disabled (its validation
		// requires >= 1 regardless of the switch).
		c.EnableRateLimiting = cfg.RateLimit > 0
		if cfg.RateLimit > 0 {
			c.RateLimit = cfg.RateLimit
			c.RateLimitWindow = cfg.RateLimitWindow
		}

		// Payload inspection is its own switch: rate limits and IP policy
		// run without the WAF unless the operator asks for it. Body scanning
		// is disabled explicitly too, so no request body is buffered when
		// inspection is off.
		c.EnablePenetrationDetection = cfg.InspectPayloads
		scanBody := cfg.InspectPayloads
		c.DetectionScanBody = &scanBody

		// Bans are opt-in: with trusted proxies unset, one shared identity
		// accumulates every client's violations, so auto-banning by default
		// would let a handful of flagged requests lock out all users behind
		// a load balancer.
		// NOTE: the engine exposes EnableRateLimitAutoBan, but its
		// route-aware CheckRateLimit path (the one every adapter uses)
		// never feeds the autoban counter, so the knob is inert through
		// any middleware. It stays unmapped here until the engine wires
		// it; advertising it would be a silent no-op.
		c.EnableIPBanning = cfg.AutoBan && !cfg.Passive
		if cfg.AutoBan {
			c.AutoBanThreshold = cfg.AutoBanThreshold
			c.AutoBanDuration = cfg.AutoBanDuration
		}

		c.Whitelist = cfg.Whitelist
		c.Blacklist = cfg.Blacklist
		c.ExemptIPs = cfg.ExemptIPs
		c.BlockCloudProviders = cfg.BlockCloudProviders
		c.ExcludePaths = mergeExclusions(cfg.ExcludePaths)
		if len(cfg.InspectCategories) > 0 {
			c.EnabledDetectionCategories = cfg.InspectCategories
		}
		if len(cfg.TrustedProxies) > 0 {
			// The engine arms spoof detection for untrusted peers carrying
			// X-Forwarded-For; the chain walk itself stays in this adapter
			// (the engine's own comment pins that split).
			c.TrustedProxies = cfg.TrustedProxies
		}

		// Pin the Redis fields before the conditional: the engine defaults
		// read REDIS_URL from the environment, and pREST deployments often
		// set that variable for their own cache. The guard shares Redis only
		// when the operator says so via PREST_GUARD_REDIS_URL.
		c.EnableRedis = false
		c.RedisURL = ""
		c.RedisPrefix = cfg.RedisPrefix
		if cfg.RedisURL != "" {
			c.EnableRedis = true
			c.RedisURL = cfg.RedisURL
			c.RedisFailOpen = cfg.RedisFailOpen
		}
	})
}

// engineFactory builds and fully initializes one engine plus its request
// handler. A failed attempt closes the half-built engine before returning
// the error.
type engineFactory func() (negroni.Handler, error)

func engineFactoryFor(engineCfg *guardcore.SecurityConfig, cfg GuardConfig, exclusions *pathExclusions, trusted *trustedProxySet) engineFactory {
	return func() (negroni.Handler, error) {
		engine, err := guardcore.NewEngine(engineCfg)
		if err != nil {
			return nil, err
		}
		if err := engine.Initialize(); err != nil {
			_ = engine.Close()
			return nil, err
		}
		handler, err := newGuardHandler(engine, cfg, exclusions, trusted)
		if err != nil {
			_ = engine.Close()
			return nil, err
		}
		return handler, nil
	}
}

// newGuardHandler composes the live request path: exclusion short-circuit,
// trusted-proxy resolution, then the engine (rate limits, IP policy, cloud
// providers, and, when inspect_payloads is on, payload detection including
// the engine's native body scan).
func newGuardHandler(engine *guardcore.Engine, cfg GuardConfig, exclusions *pathExclusions, trusted *trustedProxySet) (negroni.Handler, error) {
	wrap, err := nethttpguard.New(engine,
		nethttpguard.WithMaxBodyBytes(cfg.MaxBodyBytes),
		nethttpguard.WithLogger(bridgeLogger()),
	)
	if err != nil {
		return nil, err
	}
	return negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		// exclude_paths promises "skips guard checks"; the engine still runs
		// some checks on excluded paths, so this short-circuits first.
		if exclusions.matches(r.URL.Path) {
			next(w, r)
			return
		}
		applyTrustedProxy(r, trusted)
		wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next(w, r)
		})).ServeHTTP(w, r)
	}), nil
}

// deferredHandler answers 503 (excluded paths still pass) while a background
// goroutine retries the factory with capped exponential backoff; the first
// success swaps the live guard in, so a startup outage recovers instead of
// latching.
func deferredHandler(factory engineFactory, exclusions *pathExclusions, firstBackoff time.Duration) negroni.Handler {
	var mu sync.Mutex
	var live negroni.Handler

	go func() {
		backoff := firstBackoff
		for {
			handler, err := factory()
			if err == nil {
				mu.Lock()
				live = handler
				mu.Unlock()
				slog.Info("prest-guard: engine initialized after retry; serving")
				return
			}
			slog.Warn("prest-guard: engine initialization retry failed", "err", err)
			time.Sleep(backoff)
			if backoff < 30*time.Second {
				backoff *= 2
			}
		}
	}()

	return negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		// Health probes keep flowing even while the guard is not ready: a
		// probe that answers 503 takes the pod out of rotation, which is the
		// outage this mode exists to avoid.
		if exclusions.matches(r.URL.Path) {
			next(w, r)
			return
		}
		mu.Lock()
		handler := live
		mu.Unlock()
		if handler == nil {
			writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
			return
		}
		handler.ServeHTTP(w, r, next)
	})
}

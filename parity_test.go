package main

// Parity tests against the reference adapter (fastapi-guard): excluded paths
// still go through the engine (exclusion scoping, not a bypass), the engine
// owns client identity resolution, and the full engine config is reachable
// from the plugin's environment.

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/negroni/v3"
)

func serveWithHeaders(tb testing.TB, h negroni.Handler, target, clientIP string, headers map[string]string) (int, http.Header, string) {
	tb.Helper()
	rec := serveRaw(tb, h, target, clientIP, headers)
	return rec.Code, rec.Header(), rec.Body.String()
}

// Excluded paths are scoped by the engine, not short-circuited before it:
// the reference keeps ip_security and rate_limit enforced on them and the
// pass-through response carries the security headers.
func TestExcludedPathsGoThroughTheEngine(t *testing.T) {
	h := Load(parseConst(baseEnabled(func(c *GuardConfig) {
		c.Blacklist = []string{"203.0.113.0/24"}
	})))
	if code, _, _ := serveWithHeaders(t, h, "/_health", "203.0.113.7", nil); code != http.StatusForbidden {
		t.Fatalf("a blacklisted IP must be refused on an excluded path (reference ip_security), got %d", code)
	}
	code, header, _ := serveWithHeaders(t, h, "/_health", "198.51.100.2", nil)
	if code != http.StatusOK {
		t.Fatalf("a clean client must reach an excluded path, got %d", code)
	}
	if header.Get("X-Frame-Options") == "" {
		t.Fatalf("the engine's security headers must reach excluded-path responses: %v", header)
	}
}

func TestExcludedPathsStillSkipDetection(t *testing.T) {
	h := Load(parseConst(baseEnabled(func(c *GuardConfig) { c.InspectPayloads = true })))
	if code, _, _ := serveWithHeaders(t, h, "/_health?q=<script>alert(1)</script>", "198.51.100.2", nil); code != http.StatusOK {
		t.Fatalf("an excluded path must skip payload detection, got %d", code)
	}
	if code, _, _ := serveWithHeaders(t, h, "/articles?q=<script>alert(1)</script>", "198.51.100.2", nil); code != http.StatusBadRequest {
		t.Fatalf("the same payload must still be blocked off the excluded path, got %d", code)
	}
}

func TestPathTraversalOutOfAnExcludedSubtreeIsNotExcluded(t *testing.T) {
	h := Load(parseConst(baseEnabled(func(c *GuardConfig) {
		c.Blacklist = []string{"203.0.113.0/24"}
		c.InspectPayloads = true
	})))
	if code, _, _ := serveWithHeaders(t, h, "/_health/../articles", "203.0.113.7", nil); code != http.StatusForbidden {
		t.Fatalf("a dot-dot path leaving the excluded subtree must be guarded, got %d", code)
	}
}

// The engine resolves the client from X-Forwarded-For (reference
// extract_client_ip) under trusted_proxy_depth; the plugin no longer walks
// the chain itself.
func TestTrustedProxyDepthComesFromTheEngine(t *testing.T) {
	env := map[string]string{
		"PREST_GUARD_ENABLED":         "true",
		"PREST_GUARD_BLACKLIST":       "203.0.113.0/24",
		"PREST_GUARD_TRUSTED_PROXIES": "10.0.0.0/8",
	}
	xff := map[string]string{"X-Forwarded-For": "203.0.113.7, 10.0.0.2"}

	h := Load(parseEnv(env))
	if code, _, _ := serveWithHeaders(t, h, "/articles", "10.0.0.1", xff); code != http.StatusOK {
		t.Fatalf("with the default depth 1 the last hop is the client (reference semantics), got %d", code)
	}

	env["PREST_GUARD_TRUSTED_PROXY_DEPTH"] = "2"
	h = Load(parseEnv(env))
	if code, _, _ := serveWithHeaders(t, h, "/articles", "10.0.0.1", xff); code != http.StatusForbidden {
		t.Fatalf("depth 2 must resolve the second hop from the right, got %d", code)
	}
}

func TestUntrustedPeerForwardedHeaderIsIgnored(t *testing.T) {
	h := Load(parseEnv(map[string]string{
		"PREST_GUARD_ENABLED":         "true",
		"PREST_GUARD_BLACKLIST":       "203.0.113.0/24",
		"PREST_GUARD_TRUSTED_PROXIES": "10.0.0.0/8",
	}))
	xff := map[string]string{"X-Forwarded-For": "203.0.113.7"}
	if code, _, _ := serveWithHeaders(t, h, "/articles", "198.51.100.9", xff); code != http.StatusOK {
		t.Fatalf("a spoofed header from an untrusted peer must not move the identity, got %d", code)
	}
}

func TestInvalidTrustedProxyFailsClosed(t *testing.T) {
	h := Load(parseEnv(map[string]string{
		"PREST_GUARD_ENABLED":         "true",
		"PREST_GUARD_TRUSTED_PROXIES": "not-an-ip",
	}))
	if code, _, _ := serveWithHeaders(t, h, "/articles", "198.51.100.9", nil); code != http.StatusInternalServerError {
		t.Fatalf("an unusable trusted proxy entry must fail closed, got %d", code)
	}
}

func TestInvalidTrustedProxyDepthFailsLoud(t *testing.T) {
	cfg := ParseGuardConfig(lookupMap(map[string]string{
		"PREST_GUARD_ENABLED":             "true",
		"PREST_GUARD_TRUSTED_PROXY_DEPTH": "0",
	}))
	if cfg.Invalid == nil {
		t.Fatal("a non-positive trusted proxy depth must be recorded as a config error")
	}
}

// The full engine config: every SecurityConfig field the env subset does not
// cover is reachable through PREST_GUARD_ENGINE_CONFIG (inline JSON) or
// PREST_GUARD_ENGINE_CONFIG_FILE, applied over the env subset.
func TestEngineConfigJSONReachesFieldsOutsideTheEnvSubset(t *testing.T) {
	h := Load(parseEnv(map[string]string{
		"PREST_GUARD_ENABLED": "true",
		"PREST_GUARD_ENGINE_CONFIG": `{
			"BlockedUserAgents": ["badbot"],
			"CustomErrorResponses": {"403": "denied by policy"},
			"EnforceHTTPS": false,
			"SecurityHeaders": {"FrameOptions": "DENY"}
		}`,
	}))
	code, header, body := serveWithHeaders(t, h, "/articles", "198.51.100.2", map[string]string{"User-Agent": "badbot/1.0"})
	if code != http.StatusForbidden || !strings.Contains(body, "denied by policy") {
		t.Fatalf("a JSON-configured blocked user agent and custom error body must apply, got %d %q", code, body)
	}
	if header.Get("X-Frame-Options") != "DENY" {
		t.Fatalf("JSON-configured security headers must apply, got %v", header)
	}
}

func TestEngineConfigFileMatchesInlineJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "engine.json")
	if err := os.WriteFile(path, []byte(`{"BlockedUserAgents": ["badbot"]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	h := Load(parseEnv(map[string]string{
		"PREST_GUARD_ENABLED":            "true",
		"PREST_GUARD_ENGINE_CONFIG_FILE": path,
	}))
	if code, _, _ := serveWithHeaders(t, h, "/articles", "198.51.100.2", map[string]string{"User-Agent": "badbot/1.0"}); code != http.StatusForbidden {
		t.Fatalf("the engine config file must apply, got %d", code)
	}
}

func TestEngineConfigJSONKeepsDefaultExclusions(t *testing.T) {
	// The default /_health exclusion survives the JSON list: detection is
	// skipped there, so the payload passes.
	h := Load(parseEnv(map[string]string{
		"PREST_GUARD_ENABLED":          "true",
		"PREST_GUARD_INSPECT_PAYLOADS": "true",
		"PREST_GUARD_ENGINE_CONFIG":    `{"ExcludePaths": ["/reports"]}`,
	}))
	for _, path := range []string{"/_health", "/reports"} {
		if code, _, _ := serveWithHeaders(t, h, path+"?q=<script>alert(1)</script>", "198.51.100.2", nil); code != http.StatusOK {
			t.Fatalf("%s must stay excluded from detection, got %d", path, code)
		}
	}
}

func TestEngineConfigJSONFailsClosed(t *testing.T) {
	for name, env := range map[string]map[string]string{
		"unknown field": {"PREST_GUARD_ENGINE_CONFIG": `{"NoSuchField": 1}`},
		"bad json":      {"PREST_GUARD_ENGINE_CONFIG": `{`},
		"func field":    {"PREST_GUARD_ENGINE_CONFIG": `{"OnBlock": "x"}`},
		"invalid value": {"PREST_GUARD_ENGINE_CONFIG": `{"RateLimit": -5, "EnableRateLimiting": true}`},
		"missing file":  {"PREST_GUARD_ENGINE_CONFIG_FILE": "/nonexistent/engine.json"},
		"both sources":  {"PREST_GUARD_ENGINE_CONFIG": `{}`, "PREST_GUARD_ENGINE_CONFIG_FILE": "/nonexistent/engine.json"},
	} {
		t.Run(name, func(t *testing.T) {
			env["PREST_GUARD_ENABLED"] = "true"
			h := Load(parseEnv(env))
			if code, _, _ := serveWithHeaders(t, h, "/articles", "198.51.100.2", nil); code != http.StatusInternalServerError {
				t.Fatalf("a bad engine config must fail closed, got %d", code)
			}
		})
	}
}

func TestEngineConfigJSONIgnoredWhileGuardDisabled(t *testing.T) {
	h := Load(parseEnv(map[string]string{"PREST_GUARD_ENGINE_CONFIG": `{`}))
	if code, _, _ := serveWithHeaders(t, h, "/articles", "198.51.100.2", nil); code != http.StatusOK {
		t.Fatalf("a disabled guard must keep serving despite a bad engine config, got %d", code)
	}
}

// Reference behavior: rate_limit and ip_security stay enforced on excluded
// paths, so a probe source that shares an identity must be exempted
// explicitly (PREST_GUARD_EXEMPT_IPS) rather than relying on the path list.
func TestExcludedPathsStayRateLimitedUnlessExempt(t *testing.T) {
	h := Load(parseConst(baseEnabled(func(c *GuardConfig) {
		c.RateLimit = 1
		c.ExemptIPs = []string{"198.51.100.1"}
	})))
	if code, _, _ := serveWithHeaders(t, h, "/_health", "203.0.113.5", nil); code != http.StatusOK {
		t.Fatalf("first probe must pass, got %d", code)
	}
	if code, _, _ := serveWithHeaders(t, h, "/_health", "203.0.113.5", nil); code != http.StatusTooManyRequests {
		t.Fatalf("an excluded path is still rate limited by the engine (reference), got %d", code)
	}
	for i := 0; i < 3; i++ {
		if code, _, _ := serveWithHeaders(t, h, "/_health", "198.51.100.1", nil); code != http.StatusOK {
			t.Fatalf("an exempt probe source must never be limited, got %d", code)
		}
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/urfave/negroni/v3"
)

const okBody = "reached-next"

// serve runs one request through the middleware with a fixed client address
// and returns the recorder.
func serve(tb testing.TB, h negroni.Handler, method, target, clientIP string, body string) *httptest.ResponseRecorder {
	tb.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.RemoteAddr = clientIP + ":12345"
	rec := httptest.NewRecorder()
	var next http.HandlerFunc = func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(okBody))
	}
	if h == nil {
		next(rec, req)
		return rec
	}
	h.ServeHTTP(rec, req, next)
	return rec
}

func parseConst(cfg GuardConfig) func() GuardConfig {
	return func() GuardConfig { return cfg }
}

// envLookup parses through ParseGuardConfig so tests exercise the same path
// the plugin uses in production.
func parseEnv(env map[string]string) func() GuardConfig {
	return func() GuardConfig { return ParseGuardConfig(lookupMap(env)) }
}

func baseEnabled(mutate func(*GuardConfig)) GuardConfig {
	cfg := GuardConfig{Enabled: true}
	if mutate != nil {
		mutate(&cfg)
	}
	return cfg
}

func TestDisabledGuardPassesThrough(t *testing.T) {
	h := Load(parseConst(GuardConfig{}))
	rec := serve(t, h, http.MethodGet, "/articles", "203.0.113.5", "")
	if rec.Code != http.StatusOK || rec.Body.String() != okBody {
		t.Fatalf("disabled guard must pass through, got %d %q", rec.Code, rec.Body.String())
	}
}

func TestInvalidConfigFailsClosedWithGenericJSON(t *testing.T) {
	h := Load(parseEnv(map[string]string{
		"PREST_GUARD_ENABLED":    "true",
		"PREST_GUARD_RATE_LIMIT": "100/min",
	}))
	rec := serve(t, h, http.MethodGet, "/articles", "203.0.113.5", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("invalid active config must fail closed with 500, got %d", rec.Code)
	}
	var parsed map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("error body must be valid JSON, got %q: %v", rec.Body.String(), err)
	}
	if msg, _ := parsed["error"].(string); msg != "guard misconfigured" {
		t.Fatalf("error body must be generic (no config echo), got %q", msg)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content type must be application/json, got %q", ct)
	}
}

func TestDisabledGuardWithConfigErrorsStillServes(t *testing.T) {
	// A typo with the guard off must not take the API down; it is warned
	// about at startup instead.
	h := Load(parseEnv(map[string]string{"PREST_GUARD_RATE_LIMIT": "100/min"}))
	rec := serve(t, h, http.MethodGet, "/articles", "203.0.113.5", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("disabled guard with bad config must still serve, got %d", rec.Code)
	}
}

func TestInvalidBlacklistEntryFailsClosed(t *testing.T) {
	h := Load(parseConst(baseEnabled(func(c *GuardConfig) {
		c.Blacklist = []string{"not-an-ip"}
	})))
	rec := serve(t, h, http.MethodGet, "/articles", "203.0.113.5", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("engine config validation must fail closed, got %d", rec.Code)
	}
}

func TestBlacklistBlocksAndExemptPasses(t *testing.T) {
	h := Load(parseConst(baseEnabled(func(c *GuardConfig) {
		c.Blacklist = []string{"203.0.113.0/24"}
		c.ExemptIPs = []string{"198.51.100.1"}
	})))

	rec := serve(t, h, http.MethodGet, "/articles", "203.0.113.7", "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("blacklisted IP must get 403, got %d", rec.Code)
	}

	rec = serve(t, h, http.MethodGet, "/articles", "198.51.100.2", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("unlisted IP must pass, got %d", rec.Code)
	}
}

func TestWhitelistDeniesUnlistedPeers(t *testing.T) {
	h := Load(parseConst(baseEnabled(func(c *GuardConfig) {
		c.Whitelist = []string{"198.51.100.1"}
	})))

	rec := serve(t, h, http.MethodGet, "/articles", "198.51.100.1", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("whitelisted IP must pass, got %d", rec.Code)
	}
	rec = serve(t, h, http.MethodGet, "/articles", "203.0.113.7", "")
	if rec.Code == http.StatusOK {
		t.Fatalf("non-whitelisted IP must be denied with a whitelist configured, got %d", rec.Code)
	}
}

func TestRateLimitReturns429(t *testing.T) {
	h := Load(parseConst(baseEnabled(func(c *GuardConfig) {
		c.RateLimit = 2
		c.RateLimitWindow = 60
	})))
	ip := "203.0.113.9"
	for i := 0; i < 2; i++ {
		if rec := serve(t, h, http.MethodGet, "/articles", ip, ""); rec.Code != http.StatusOK {
			t.Fatalf("request %d must pass under limit, got %d", i+1, rec.Code)
		}
	}
	if rec := serve(t, h, http.MethodGet, "/articles", ip, ""); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("over-limit request must get 429, got %d", rec.Code)
	}
	if rec := serve(t, h, http.MethodGet, "/articles", "203.0.113.10", ""); rec.Code != http.StatusOK {
		t.Fatalf("a different client must keep its own budget, got %d", rec.Code)
	}
}

func TestHealthEndpointsExcludedByDefault(t *testing.T) {
	h := Load(parseConst(baseEnabled(func(c *GuardConfig) {
		c.RateLimit = 1
		c.RateLimitWindow = 60
		c.Blacklist = []string{"203.0.113.0/24"}
	})))
	ip := "203.0.113.7" // blacklisted; exclusion short-circuits before the engine

	for _, path := range []string{"/_health", "/_health/ready", "/_ready", "/docs", "/static/app.js"} {
		if rec := serve(t, h, http.MethodGet, path, ip, ""); rec.Code != http.StatusOK {
			t.Fatalf("%s must be excluded by default even for a blacklisted IP, got %d", path, rec.Code)
		}
	}
	if rec := serve(t, h, http.MethodGet, "/healthcheck", ip, ""); rec.Code == http.StatusOK {
		t.Fatal("/healthcheck is not a default exclusion and must still be guarded")
	}
}

func TestOperatorExclusionsMergeWithDefaults(t *testing.T) {
	h := Load(parseConst(baseEnabled(func(c *GuardConfig) {
		c.RateLimit = 1
		c.ExcludePaths = []string{"/reports"}
	})))
	ip := "203.0.113.9"
	if rec := serve(t, h, http.MethodGet, "/reports/daily", ip, ""); rec.Code != http.StatusOK {
		t.Fatalf("operator exclusion must skip checks, got %d", rec.Code)
	}
	if rec := serve(t, h, http.MethodGet, "/_health", ip, ""); rec.Code != http.StatusOK {
		t.Fatalf("defaults must survive operator exclusions, got %d", rec.Code)
	}
	// Rate limits key per client and endpoint: the first /articles request
	// passes, the second trips the limit of 1.
	if rec := serve(t, h, http.MethodGet, "/articles", ip, ""); rec.Code != http.StatusOK {
		t.Fatalf("non-excluded path first request must pass, got %d", rec.Code)
	}
	if rec := serve(t, h, http.MethodGet, "/articles", ip, ""); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("non-excluded path must stay rate limited, got %d", rec.Code)
	}
}

// realisticPRESTRequests are the shapes pREST sends every day: query filter
// operators, ordering, grouping, and JSON bodies carrying ordinary prose.
func realisticPRESTRequests() []struct{ method, target, body string } {
	return []struct{ method, target, body string }{
		{http.MethodGet, "/articles?_select=*&_page=1&_page_size=20", ""},
		{http.MethodGet, "/articles?_order=-created_at,title", ""},
		{http.MethodGet, "/authors?_where=(name.eq.John,age.gt.30)&_groupby=city", ""},
		{http.MethodGet, "/products?_count=*&_distinct=true", ""},
		{http.MethodGet, "/authors?_where=email.like.%25gmail.com", ""},
		{http.MethodGet, "/books?title=do%20as%20or%20the", ""},
		{http.MethodPost, "/articles/", `{"title":"My first post","body":"Books, commas, and (parentheses); also 2+2=4 and 1<2. The quick brown fox, do as or the, jumps over lazy dogs."}`},
		{http.MethodPost, "/reviews/", `{"comment":"Great product!!! Would buy again :) (verified purchase)","rating":5}`},
		{http.MethodPost, "/notes/", `{"text":"SELECT your favorite color in the poll below; results update hourly."}`},
	}
}

func TestInspectionOffByDefaultLetsAttackQueriesPass(t *testing.T) {
	h := Load(parseConst(baseEnabled(nil)))
	rec := serve(t, h, http.MethodGet, "/articles?search=<script>alert(1)</script>", "203.0.113.11", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("with inspection off (default) the guard must not reject payloads, got %d", rec.Code)
	}
}

func TestInspectionOnRealisticPRESTTrafficPasses(t *testing.T) {
	h := Load(parseConst(baseEnabled(func(c *GuardConfig) {
		c.InspectPayloads = true
	})))
	for i, rr := range realisticPRESTRequests() {
		rec := serve(t, h, rr.method, rr.target, "203.0.113.12", rr.body)
		if rec.Code != http.StatusOK {
			t.Fatalf("realistic request %d (%s %s) must pass with inspection on, got %d: %s",
				i, rr.method, rr.target, rec.Code, rec.Body.String())
		}
	}
}

func TestInspectionOnBlocksAttacks(t *testing.T) {
	h := Load(parseConst(baseEnabled(func(c *GuardConfig) {
		c.InspectPayloads = true
	})))
	attacks := []struct{ name, method, target, body string }{
		{"xss query param", http.MethodGet, "/articles?search=<script>alert(1)</script>", ""},
		// Corpus-verified query_param sqli case (sqli_union_password,
		// guard-core-spec 4.1.0).
		{"sqli query param", http.MethodGet, "/articles?_where=" + url.QueryEscape("' AND 1=2 UNION SELECT username, password FROM users--"), ""},
		{"xss in json body", http.MethodPost, "/comments/", `{"text":"<script>alert(1)</script>"}`},
		{"sqli in json body", http.MethodPost, "/comments/", `{"name":"a","filter":"x' OR '1'='1"}`},
	}
	for _, a := range attacks {
		rec := serve(t, h, a.method, a.target, "203.0.113.13", a.body)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s must be blocked with 400 under inspection, got %d", a.name, rec.Code)
		}
	}
}

func TestPassiveModeLogsButPasses(t *testing.T) {
	h := Load(parseConst(baseEnabled(func(c *GuardConfig) {
		c.InspectPayloads = true
		c.Passive = true
	})))
	rec := serve(t, h, http.MethodGet, "/articles?search=<script>alert(1)</script>", "203.0.113.14", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("passive mode must log and pass, got %d", rec.Code)
	}
}

func TestAutoBanOffByDefaultSurvivesRepeatedAttacks(t *testing.T) {
	h := Load(parseConst(baseEnabled(func(c *GuardConfig) {
		c.InspectPayloads = true
	})))
	ip := "203.0.113.15"
	for i := 0; i < 12; i++ {
		serve(t, h, http.MethodGet, "/articles?search=<script>alert(1)</script>", ip, "")
	}
	rec := serve(t, h, http.MethodGet, "/articles", ip, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("auto ban is opt-in; clean traffic must pass after repeated attacks, got %d", rec.Code)
	}
}

func TestAutoBanOptInBansAfterThreshold(t *testing.T) {
	h := Load(parseConst(baseEnabled(func(c *GuardConfig) {
		c.InspectPayloads = true
		c.AutoBan = true
		c.AutoBanThreshold = 2
	})))
	ip := "203.0.113.16"
	// The violation that reaches the threshold applies the ban within the
	// same request, so the second attack answers 400 or 403.
	if rec := serve(t, h, http.MethodGet, "/articles?search=<script>alert(1)</script>", ip, ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("first attack must be blocked with 400, got %d", rec.Code)
	}
	serve(t, h, http.MethodGet, "/articles?search=<script>alert(1)</script>", ip, "")
	rec := serve(t, h, http.MethodGet, "/articles", ip, "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("after threshold the IP must be banned (403), got %d", rec.Code)
	}
}

// NOTE: there is no rate-limit auto-ban test: the engine's route-aware
// CheckRateLimit path never feeds its autoban counter, so the knob would be
// inert through any middleware (see the note in buildEngineConfig). It gets
// a test the moment the engine wires it.

func TestRedisFailOpenDefaultServesWithoutRedis(t *testing.T) {
	// Routed through ParseGuardConfig so the fail-open default applies, the
	// same path production takes.
	h := Load(parseEnv(map[string]string{
		"PREST_GUARD_ENABLED":    "true",
		"PREST_GUARD_RATE_LIMIT": "100",
		"PREST_GUARD_REDIS_URL":  "redis://127.0.0.1:1", // nothing listens here
	}))
	if rec := serve(t, h, http.MethodGet, "/articles", "203.0.113.18", ""); rec.Code != http.StatusOK {
		t.Fatalf("fail-open default must serve from per-instance memory when Redis is down, got %d", rec.Code)
	}
}

func TestRedisFailClosedAnswers503AndKeepsHealthAlive(t *testing.T) {
	h := Load(parseEnv(map[string]string{
		"PREST_GUARD_ENABLED":         "true",
		"PREST_GUARD_REDIS_URL":       "redis://127.0.0.1:1",
		"PREST_GUARD_REDIS_FAIL_OPEN": "false",
	}))
	if rec := serve(t, h, http.MethodGet, "/articles", "203.0.113.19", ""); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("fail-closed init failure must answer 503, got %d", rec.Code)
	}
	rec := serve(t, h, http.MethodGet, "/articles", "203.0.113.19", "")
	var parsed map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("503 body must be valid JSON, got %q", rec.Body.String())
	}
	if rec = serve(t, h, http.MethodGet, "/_health", "203.0.113.19", ""); rec.Code != http.StatusOK {
		t.Fatalf("health probes must keep flowing while the guard is not ready, got %d", rec.Code)
	}
}

func TestDeferredHandlerSwapsLiveGuardIn(t *testing.T) {
	exclusions := newPathExclusions(mergeExclusions(nil))
	attempts := 0
	factory := func() (negroni.Handler, error) {
		attempts++
		if attempts == 1 {
			return nil, fmt.Errorf("engine not ready")
		}
		return negroni.HandlerFunc(func(w http.ResponseWriter, _ *http.Request, next http.HandlerFunc) {
			w.WriteHeader(http.StatusTeapot)
		}), nil
	}
	h := deferredHandler(factory, exclusions, 1) // ms-scale backoff for the test

	if rec := serve(t, h, http.MethodGet, "/articles", "203.0.113.20", ""); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("before the retry succeeds the request must get 503, got %d", rec.Code)
	}
	deadline := 100 // bounded wait for the background retry
	for i := 0; i < deadline; i++ {
		rec := serve(t, h, http.MethodGet, "/articles", "203.0.113.20", "")
		if rec.Code == http.StatusTeapot {
			return // live guard swapped in
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatal("deferred handler never swapped the live guard in")
}

package main

import (
	"net/http"
	"testing"

	"github.com/rennf93/guard-core-go/v4/guardcore"
)

// GuardConfig.Configure is the Go-level escape hatch for a plugin built from
// a fork: it receives the fully built engine config (func and interface
// fields such as CustomResponseModifier, AgentHandler or GeoIPHandler
// included) after the env subset and the engine config JSON were applied.
func TestConfigureHookEditsTheEngineConfigLast(t *testing.T) {
	cfg := ParseGuardConfig(lookupMap(map[string]string{
		"PREST_GUARD_ENABLED":       "true",
		"PREST_GUARD_ENGINE_CONFIG": `{"BlockedUserAgents": ["badbot"]}`,
	}))
	cfg.Configure = func(c *guardcore.SecurityConfig) {
		c.BlockedUserAgents = []string{"worsebot"}
	}
	h := Load(parseConst(cfg))
	if code, _, _ := serveWithHeaders(t, h, "/articles", "198.51.100.2", map[string]string{"User-Agent": "badbot/1.0"}); code != http.StatusOK {
		t.Fatalf("the hook runs after the JSON and replaces it, got %d", code)
	}
	if code, _, _ := serveWithHeaders(t, h, "/articles", "198.51.100.2", map[string]string{"User-Agent": "worsebot/1.0"}); code != http.StatusForbidden {
		t.Fatalf("the hook's own setting must apply, got %d", code)
	}
}

func TestConfigureHookRunsBeforeValidation(t *testing.T) {
	cfg := baseEnabled(nil)
	cfg.Configure = func(c *guardcore.SecurityConfig) { c.Blacklist = []string{"not-an-ip"} }
	h := Load(parseConst(cfg))
	if code, _, _ := serveWithHeaders(t, h, "/articles", "198.51.100.2", nil); code != http.StatusInternalServerError {
		t.Fatalf("an invalid hook result must fail closed, got %d", code)
	}
}

package main

import (
	"strings"
	"testing"
)

func lookupMap(env map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		v, ok := env[key]
		return v, ok
	}
}

func TestParseGuardConfigDefaults(t *testing.T) {
	cfg := ParseGuardConfig(lookupMap(nil))

	if cfg.Enabled || cfg.active() {
		t.Fatal("guard must be disabled by default")
	}
	if cfg.Passive || cfg.InspectPayloads || cfg.AutoBan {
		t.Fatal("passive, inspection, and bans must be off by default")
	}
	if cfg.RateLimit != 0 {
		t.Fatalf("rate limit must default to 0, got %d", cfg.RateLimit)
	}
	if cfg.RateLimitWindow != 60 {
		t.Fatalf("rate limit window must default to 60, got %d", cfg.RateLimitWindow)
	}
	if cfg.MaxBodyBytes != 8<<10 {
		t.Fatalf("max body bytes must default to 8 KiB, got %d", cfg.MaxBodyBytes)
	}
	if !cfg.RedisFailOpen {
		t.Fatal("redis fail open must default to true")
	}
	if cfg.RedisPrefix != "prest_guard" {
		t.Fatalf("redis prefix must default to prest_guard, got %q", cfg.RedisPrefix)
	}
	if cfg.AutoBanThreshold != 10 || cfg.AutoBanDuration != 3600 {
		t.Fatalf("auto ban defaults wrong: threshold %d duration %d", cfg.AutoBanThreshold, cfg.AutoBanDuration)
	}
	if cfg.Invalid != nil {
		t.Fatalf("defaults must parse clean, got %v", cfg.Invalid)
	}
}

func TestParseGuardConfigFullValid(t *testing.T) {
	env := map[string]string{
		"PREST_GUARD_ENABLED":               "true",
		"PREST_GUARD_PASSIVE":               "false",
		"PREST_GUARD_INSPECT_PAYLOADS":      "true",
		"PREST_GUARD_INSPECT_CATEGORIES":    "sqli, xss,cmd_injection",
		"PREST_GUARD_AUTO_BAN":              "1",
		"PREST_GUARD_RATE_LIMIT_AUTO_BAN":   "true",
		"PREST_GUARD_RATE_LIMIT":            "100",
		"PREST_GUARD_RATE_LIMIT_WINDOW":     "30",
		"PREST_GUARD_AUTO_BAN_THRESHOLD":    "5",
		"PREST_GUARD_AUTO_BAN_DURATION":     "600",
		"PREST_GUARD_MAX_BODY_BYTES":        "131072",
		"PREST_GUARD_BLACKLIST":             "203.0.113.0/24, 198.51.100.9",
		"PREST_GUARD_WHITELIST":             "10.0.0.1,10.0.0.2",
		"PREST_GUARD_EXEMPT_IPS":            "192.0.2.1",
		"PREST_GUARD_EXCLUDE_PATHS":         "/reports,/auth",
		"PREST_GUARD_TRUSTED_PROXIES":       "10.0.0.9,172.16.0.0/12",
		"PREST_GUARD_BLOCK_CLOUD_PROVIDERS": "AWS,GCP",
		"PREST_GUARD_REDIS_URL":             "redis://redis:6379",
		"PREST_GUARD_REDIS_PREFIX":          "custom_prefix",
		"PREST_GUARD_REDIS_FAIL_OPEN":       "false",
	}
	cfg := ParseGuardConfig(lookupMap(env))

	if !cfg.Enabled || !cfg.InspectPayloads || !cfg.AutoBan {
		t.Fatalf("boolean switches wrong: %+v", cfg)
	}
	if cfg.Passive {
		t.Fatal("passive must be false")
	}
	if cfg.RateLimit != 100 || cfg.RateLimitWindow != 30 {
		t.Fatalf("rate limit wrong: %d per %d", cfg.RateLimit, cfg.RateLimitWindow)
	}
	if cfg.AutoBanThreshold != 5 || cfg.AutoBanDuration != 600 {
		t.Fatalf("auto ban wrong: %d for %d", cfg.AutoBanThreshold, cfg.AutoBanDuration)
	}
	if cfg.MaxBodyBytes != 131072 {
		t.Fatalf("max body bytes wrong: %d", cfg.MaxBodyBytes)
	}
	if len(cfg.InspectCategories) != 3 || cfg.InspectCategories[1] != "xss" {
		t.Fatalf("categories wrong: %v", cfg.InspectCategories)
	}
	if len(cfg.Blacklist) != 2 || cfg.Blacklist[1] != "198.51.100.9" {
		t.Fatalf("blacklist must split on commas and trim, got %v", cfg.Blacklist)
	}
	if len(cfg.Whitelist) != 2 || len(cfg.ExemptIPs) != 1 {
		t.Fatalf("ip lists wrong: %v %v", cfg.Whitelist, cfg.ExemptIPs)
	}
	if len(cfg.TrustedProxies) != 2 || cfg.TrustedProxies[1] != "172.16.0.0/12" {
		t.Fatalf("trusted proxies wrong: %v", cfg.TrustedProxies)
	}
	if cfg.RedisURL != "redis://redis:6379" || cfg.RedisPrefix != "custom_prefix" || cfg.RedisFailOpen {
		t.Fatalf("redis settings wrong: %s %s %v", cfg.RedisURL, cfg.RedisPrefix, cfg.RedisFailOpen)
	}
	if cfg.Invalid != nil {
		t.Fatalf("valid config must not carry errors: %v", cfg.Invalid)
	}
}

func TestParseGuardConfigInvalidValuesFailLoud(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{
			name: "invalid enabled bool",
			env:  map[string]string{"PREST_GUARD_ENABLED": "yes"},
			want: "PREST_GUARD_ENABLED",
		},
		{
			name: "numeric typo does not silently disable rate limiting",
			env:  map[string]string{"PREST_GUARD_RATE_LIMIT": "100/min"},
			want: "PREST_GUARD_RATE_LIMIT",
		},
		{
			name: "negative rate limit",
			env:  map[string]string{"PREST_GUARD_RATE_LIMIT": "-5"},
			want: "negative",
		},
		{
			name: "zero window",
			env:  map[string]string{"PREST_GUARD_RATE_LIMIT_WINDOW": "0"},
			want: "PREST_GUARD_RATE_LIMIT_WINDOW",
		},
		{
			name: "negative body bytes",
			env:  map[string]string{"PREST_GUARD_MAX_BODY_BYTES": "-1"},
			want: "PREST_GUARD_MAX_BODY_BYTES",
		},
		{
			name: "invalid passive bool",
			env:  map[string]string{"PREST_GUARD_PASSIVE": "on"},
			want: "PREST_GUARD_PASSIVE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ParseGuardConfig(lookupMap(tt.env))
			if cfg.Invalid == nil {
				t.Fatal("expected a config error")
			}
			if !strings.Contains(cfg.Invalid.Error(), tt.want) {
				t.Fatalf("error %q must name the offending variable", cfg.Invalid)
			}
		})
	}
}

func TestParseGuardConfigEnabledInvalidMeansActive(t *testing.T) {
	cfg := ParseGuardConfig(lookupMap(map[string]string{"PREST_GUARD_ENABLED": "yes"}))
	if cfg.Enabled {
		t.Fatal("invalid value must not enable the guard")
	}
	if !cfg.active() {
		t.Fatal("an invalid enable switch must count as active so errors fail closed")
	}
}

func TestParseGuardConfigMultipleErrorsJoined(t *testing.T) {
	cfg := ParseGuardConfig(lookupMap(map[string]string{
		"PREST_GUARD_ENABLED":    "maybe",
		"PREST_GUARD_RATE_LIMIT": "soon",
	}))
	if cfg.Invalid == nil || strings.Count(cfg.Invalid.Error(), ";") != 1 {
		t.Fatalf("both errors must be reported, got %v", cfg.Invalid)
	}
}

func TestParseGuardConfigFromEnv(t *testing.T) {
	t.Setenv("PREST_GUARD_ENABLED", "true")
	t.Setenv("PREST_GUARD_RATE_LIMIT", "42")
	cfg := ParseGuardConfigFromEnv()
	if !cfg.Enabled || cfg.RateLimit != 42 {
		t.Fatalf("env parse wrong: %+v", cfg)
	}
}

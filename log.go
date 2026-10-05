package main

import (
	"log"
	"log/slog"
	"strings"

	"github.com/rennf93/guard-core-go/v4/guardcore"
)

// onBlockLogger adapts the engine's OnBlock hook to slog. The engine-provided
// payload carries only metadata (check name, reason, client IP, method,
// path); request secrets such as headers, query strings, or body content
// never reach this logger.
func onBlockLogger() func(guardcore.Request, map[string]any) {
	return func(_ guardcore.Request, payload map[string]any) {
		slog.Warn("prest-guard: request blocked",
			"passive", boolVal(payload["passive_mode"]),
			"check", strVal(payload["check_name"]),
			"reason", strVal(payload["reason"]),
			"trigger", strVal(payload["trigger_info"]),
			"client_ip", strVal(payload["client_ip"]),
			"method", strVal(payload["method"]),
			"path", strVal(payload["path"]),
		)
	}
}

func strVal(v any) string {
	s, _ := v.(string)
	return s
}

func boolVal(v any) bool {
	b, _ := v.(bool)
	return b
}

// slogBridge routes the nethttp adapter's stdlib logger into slog so the
// plugin emits one log stream.
type slogBridge struct {
	logger *slog.Logger
}

func (b slogBridge) Write(p []byte) (int, error) {
	b.logger.Info(strings.TrimSpace(string(p)))
	return len(p), nil
}

// bridgeLogger returns a stdlib logger backed by the default slog logger.
func bridgeLogger() *log.Logger {
	return log.New(slogBridge{logger: slog.Default()}, "", 0)
}

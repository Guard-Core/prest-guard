package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/urfave/negroni/v3"
)

// silenceLogs keeps block-event warnings (part of the measured reject path)
// from corrupting benchmark output.
func silenceLogs(b *testing.B) {
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	b.Cleanup(func() { slog.SetDefault(old) })
}

func serveB(tb testing.TB, h negroni.Handler, method, target, clientIP, body string) *httptest.ResponseRecorder {
	return serve(tb, h, method, target, clientIP, body)
}

func httptestRequest() *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/articles", nil)
	req.RemoteAddr = "192.0.2.1:1"
	return req
}

func httptestRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// bodyKinds are the workloads the inspection layer can face:
//
//   - prose: ordinary JSON payloads, measured on the pass path. This is the
//     cost every inspected request pays.
//   - encoded: percent-encoding chains that decode to benign text. Measures
//     the preprocessor's decode work on the pass path.
//   - threat: an outright injection payload. Measures the reject path, which
//     is not cheaper than the pass path (cost is body-size driven).
type bodyKind string

const (
	kindProse   bodyKind = "prose"
	kindEncoded bodyKind = "encoded"
	kindThreat  bodyKind = "threat"
)

func generateBody(kind bodyKind, size int) string {
	var b strings.Builder
	for b.Len() < size {
		switch kind {
		case kindProse:
			b.WriteString("The quick brown fox jumps over the lazy dog. " +
				"Books, commas, and (parentheses); numbers like 2+2=4, dates like 2026-10-05, " +
				"and the words do, as, or, the appear here for no reason at all. ")
		case kindEncoded:
			b.WriteString("a%2531b%2532c%2533d%2534e%2535f")
		case kindThreat:
			b.WriteString("1' OR '1'='1 -- ")
		}
	}
	return b.String()[:size]
}

func benchmarkGuard(b *testing.B, kind bodyKind, size int) {
	silenceLogs(b)
	cfg := baseEnabled(func(c *GuardConfig) {
		c.InspectPayloads = true
		c.MaxBodyBytes = 1 << 20 // scan everything the benchmark sends
	})
	h := Load(parseConst(cfg))

	body := generateBody(kind, size)
	wantCode := http.StatusOK
	if kind == kindThreat {
		wantCode = http.StatusBadRequest // rejected by design
	}
	// The workload must behave as labeled before timing starts, or the
	// number measures a path it claims not to.
	if rec := serve(b, h, http.MethodPost, "/comments/", "192.0.2.1", body); rec.Code != wantCode {
		b.Fatalf("%s %d bytes got %d, expected %d", kind, size, rec.Code, wantCode)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := serveB(b, h, http.MethodPost, "/comments/", "192.0.2.1", body)
		if rec.Code != wantCode {
			b.Fatalf("%s body must answer %d, got %d", kind, wantCode, rec.Code)
		}
	}
}

func BenchmarkGuardInspection8KiB(b *testing.B) {
	b.Run("prose", func(b *testing.B) { benchmarkGuard(b, kindProse, 8<<10) })
	b.Run("encoded", func(b *testing.B) { benchmarkGuard(b, kindEncoded, 8<<10) })
	b.Run("threat", func(b *testing.B) { benchmarkGuard(b, kindThreat, 8<<10) })
}

func BenchmarkGuardInspection16KiB(b *testing.B) {
	b.Run("prose", func(b *testing.B) { benchmarkGuard(b, kindProse, 16<<10) })
	b.Run("encoded", func(b *testing.B) { benchmarkGuard(b, kindEncoded, 16<<10) })
	b.Run("threat", func(b *testing.B) { benchmarkGuard(b, kindThreat, 16<<10) })
}

func BenchmarkGuardInspection64KiB(b *testing.B) {
	b.Run("prose", func(b *testing.B) { benchmarkGuard(b, kindProse, 64<<10) })
	b.Run("encoded", func(b *testing.B) { benchmarkGuard(b, kindEncoded, 64<<10) })
	b.Run("threat", func(b *testing.B) { benchmarkGuard(b, kindThreat, 64<<10) })
}

func BenchmarkGuardInspection256KiB(b *testing.B) {
	b.Run("prose", func(b *testing.B) { benchmarkGuard(b, kindProse, 256<<10) })
	b.Run("encoded", func(b *testing.B) { benchmarkGuard(b, kindEncoded, 256<<10) })
	b.Run("threat", func(b *testing.B) { benchmarkGuard(b, kindThreat, 256<<10) })
}

func BenchmarkGuardInspection1MiB(b *testing.B) {
	b.Run("prose", func(b *testing.B) { benchmarkGuard(b, kindProse, 1<<20) })
	b.Run("encoded", func(b *testing.B) { benchmarkGuard(b, kindEncoded, 1<<20) })
	b.Run("threat", func(b *testing.B) { benchmarkGuard(b, kindThreat, 1<<20) })
}

func BenchmarkGuardRateLimitOnly(b *testing.B) {
	silenceLogs(b)
	cfg := baseEnabled(func(c *GuardConfig) {
		c.RateLimit = 1_000_000 // effectively unlimited; measures the check itself
	})
	h := Load(parseConst(cfg))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptestRequest()
		rec := httptestRecorder()
		h.ServeHTTP(rec, req, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected %d", rec.Code)
		}
	}
}

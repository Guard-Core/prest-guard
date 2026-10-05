package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func clientAddrAfterProxy(req *http.Request) string {
	return requestClientHost(req)
}

func TestTrustedProxyResolvesRightmostUntrusted(t *testing.T) {
	trusted, err := newTrustedProxySet([]string{"10.0.0.9"})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.9:54321"
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.9")

	applyTrustedProxy(req, trusted)

	if got := clientAddrAfterProxy(req); got != "203.0.113.5" {
		t.Fatalf("client must resolve to the rightmost untrusted hop, got %q", got)
	}
}

func TestUntrustedPeerKeepsForwardedHeadersIgnored(t *testing.T) {
	trusted, err := newTrustedProxySet([]string{"10.0.0.9"})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.99:54321"
	req.Header.Set("X-Forwarded-For", "198.51.100.1")

	applyTrustedProxy(req, trusted)

	if got := clientAddrAfterProxy(req); got != "203.0.113.99" {
		t.Fatalf("forwarded headers from an untrusted peer must be ignored, got %q", got)
	}
}

func TestTrustedProxyJoinsMultipleHeaderLines(t *testing.T) {
	trusted, err := newTrustedProxySet([]string{"10.0.0.9"})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.9:54321"
	// Some proxies append a second header line instead of joining the hop
	// into the existing value.
	req.Header.Add("X-Forwarded-For", "198.51.100.7")
	req.Header.Add("X-Forwarded-For", "10.0.0.9")

	applyTrustedProxy(req, trusted)

	if got := clientAddrAfterProxy(req); got != "198.51.100.7" {
		t.Fatalf("every header line must take part in the walk, got %q", got)
	}
}

func TestTrustedProxyAllHopsTrustedFallsBackToLeftmost(t *testing.T) {
	trusted, err := newTrustedProxySet([]string{"10.0.0.9", "10.0.0.8"})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.9:54321"
	req.Header.Set("X-Forwarded-For", "10.0.0.8, 10.0.0.9")

	applyTrustedProxy(req, trusted)

	if got := clientAddrAfterProxy(req); got != "10.0.0.8" {
		t.Fatalf("when every hop is trusted the leftmost entry is used, got %q", got)
	}
}

func TestTrustedProxyMalformedHopKeepsPeer(t *testing.T) {
	trusted, err := newTrustedProxySet([]string{"10.0.0.9"})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.9:54321"
	req.Header.Set("X-Forwarded-For", "not-an-ip, 10.0.0.9")

	applyTrustedProxy(req, trusted)

	if got := clientAddrAfterProxy(req); got != "10.0.0.9" {
		t.Fatalf("a malformed chain must not rewrite identity, got %q", got)
	}
}

func TestTrustedProxyCIDRRange(t *testing.T) {
	trusted, err := newTrustedProxySet([]string{"172.16.0.0/12"})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "172.16.9.9:54321"
	req.Header.Set("X-Forwarded-For", "203.0.113.8")

	applyTrustedProxy(req, trusted)

	if got := clientAddrAfterProxy(req); got != "203.0.113.8" {
		t.Fatalf("CIDR entries must be honored, got %q", got)
	}
}

func TestNewTrustedProxySetRejectsGarbage(t *testing.T) {
	if _, err := newTrustedProxySet([]string{"obviously-not-ip"}); err == nil {
		t.Fatal("invalid entries must fail at load, not mis-resolve at request time")
	}
	if _, err := newTrustedProxySet(nil); err != nil {
		t.Fatalf("empty set is valid: %v", err)
	}
}

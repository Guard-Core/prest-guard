package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

// trustedProxySet holds pre-parsed trusted proxy IPs and CIDRs.
type trustedProxySet struct {
	ips  map[string]bool
	nets []*net.IPNet
}

// newTrustedProxySet parses trusted proxy entries. Any entry that is neither
// an IP nor a CIDR is a config error, so the guard fails closed at load
// rather than mis-resolving clients at request time.
func newTrustedProxySet(entries []string) (*trustedProxySet, error) {
	if len(entries) == 0 {
		return nil, nil
	}
	set := &trustedProxySet{ips: make(map[string]bool, len(entries))}
	for i, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if ip := net.ParseIP(entry); ip != nil {
			set.ips[ip.String()] = true
			continue
		}
		_, cidr, err := net.ParseCIDR(entry)
		if err != nil {
			return nil, fmt.Errorf("trusted_proxies[%d]: %q is not an IP or CIDR", i, entry)
		}
		set.nets = append(set.nets, cidr)
	}
	return set, nil
}

// contains reports whether ip belongs to the trusted set.
func (t *trustedProxySet) contains(ip net.IP) bool {
	if t == nil || ip == nil {
		return false
	}
	if t.ips[ip.String()] {
		return true
	}
	for _, n := range t.nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// applyTrustedProxy rewrites r.RemoteAddr to the real client IP when the
// direct peer is a trusted proxy and X-Forwarded-For is present. Resolution
// walks the hop list from right to left and picks the rightmost entry that
// is not itself a trusted proxy (the standard rightmost-untrusted rule);
// when every hop is trusted the leftmost entry is used. When the peer is not
// trusted the forwarded header is attacker-controlled and ignored entirely,
// leaving RemoteAddr untouched. The rewritten address keeps a port suffix
// (net.JoinHostPort) so downstream net.SplitHostPort keeps working.
//
// The engine deliberately leaves the chain walk to adapters; it only arms
// spoof detection, so this walk is the plugin's half of that contract.
func applyTrustedProxy(r *http.Request, trusted *trustedProxySet) {
	if trusted == nil {
		return
	}
	peer := net.ParseIP(requestClientHost(r))
	if !trusted.contains(peer) {
		return
	}
	// Header.Values joins every X-Forwarded-For line: some proxies append a
	// new header line instead of joining their hop into the existing value,
	// and taking only the first line would trust client-supplied data.
	xff := strings.Join(r.Header.Values("X-Forwarded-For"), ",")
	if strings.TrimSpace(xff) == "" {
		return
	}
	hops := strings.Split(xff, ",")
	for i := len(hops) - 1; i >= 0; i-- {
		hop := net.ParseIP(strings.TrimSpace(hops[i]))
		if hop == nil {
			// Malformed hop: the chain cannot be trusted, keep the peer.
			return
		}
		if trusted.contains(hop) {
			continue
		}
		r.RemoteAddr = net.JoinHostPort(hop.String(), "0")
		return
	}
	// All hops are trusted proxies: fall back to the leftmost entry.
	leftmost := net.ParseIP(strings.TrimSpace(hops[0]))
	if leftmost == nil {
		return
	}
	r.RemoteAddr = net.JoinHostPort(leftmost.String(), "0")
}

// requestClientHost extracts the bare host from RemoteAddr, tolerating
// addresses without a port.
func requestClientHost(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

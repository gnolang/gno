package state

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestLimiter(cfg RateLimitConfig) *IPLimiter {
	l := NewIPLimiter(cfg)
	return l
}

func TestRateLimitAllowsBurst(t *testing.T) {
	now := time.Unix(0, 0)
	l := newTestLimiter(RateLimitConfig{
		PerMinute: 100,
		Burst:     100,
		NowFunc:   func() time.Time { return now },
	})
	for i := range 100 {
		if !l.Allow("1.2.3.4") {
			t.Fatalf("request %d: expected allowed, got rejected", i+1)
		}
	}
}

func TestRateLimitRejectsOverLimit(t *testing.T) {
	now := time.Unix(0, 0)
	l := newTestLimiter(RateLimitConfig{
		PerMinute: 100,
		Burst:     100,
		NowFunc:   func() time.Time { return now },
	})
	for i := range 100 {
		if !l.Allow("1.2.3.4") {
			t.Fatalf("warmup %d: unexpected reject", i+1)
		}
	}
	if l.Allow("1.2.3.4") {
		t.Fatal("101st request: expected reject, got allow")
	}
}

func TestRateLimitRefillsOverTime(t *testing.T) {
	current := time.Unix(0, 0)
	l := newTestLimiter(RateLimitConfig{
		PerMinute: 100,
		Burst:     100,
		NowFunc:   func() time.Time { return current },
	})
	for i := range 100 {
		if !l.Allow("1.2.3.4") {
			t.Fatalf("warmup %d: unexpected reject", i+1)
		}
	}
	if l.Allow("1.2.3.4") {
		t.Fatal("expected reject immediately after burst")
	}
	current = current.Add(60 * time.Second)
	for i := range 100 {
		if !l.Allow("1.2.3.4") {
			t.Fatalf("after refill, request %d: unexpected reject", i+1)
		}
	}
	if l.Allow("1.2.3.4") {
		t.Fatal("expected reject after refilled-bucket burst")
	}
}

func TestRateLimitEvictsLRU(t *testing.T) {
	now := time.Unix(0, 0)
	l := newTestLimiter(RateLimitConfig{
		PerMinute: 100,
		Burst:     100,
		MaxIPs:    3,
		NowFunc:   func() time.Time { return now },
	})
	for _, ip := range []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"} {
		if !l.Allow(ip) {
			t.Fatalf("warmup %s: unexpected reject", ip)
		}
	}
	if got := l.size(); got != 3 {
		t.Fatalf("size after warmup = %d, want 3", got)
	}
	if !l.Allow("4.4.4.4") {
		t.Fatal("new ip: unexpected reject")
	}
	if got := l.size(); got != 3 {
		t.Fatalf("size after eviction = %d, want 3", got)
	}
	if l.has("1.1.1.1") {
		t.Fatal("oldest IP 1.1.1.1 should have been evicted")
	}
	if !l.has("4.4.4.4") {
		t.Fatal("new IP 4.4.4.4 should be tracked")
	}
}

func mustCIDRs(t *testing.T, cidrs ...string) []*net.IPNet {
	t.Helper()
	nets, err := ParseTrustedProxies(cidrs)
	if err != nil {
		t.Fatalf("ParseTrustedProxies(%v) error: %v", cidrs, err)
	}
	if len(nets) != len(cidrs) {
		t.Fatalf("ParseTrustedProxies(%v) = %d nets, want %d", cidrs, len(nets), len(cidrs))
	}
	return nets
}

// TestExtractIPTrustedProxy: X-Real-IP is honored only when RemoteAddr is
// inside a trusted CIDR, is ignored otherwise, and garbage header values
// always fall back to RemoteAddr.
func TestExtractIPTrustedProxy(t *testing.T) {
	trusted := mustCIDRs(t, "10.0.0.0/8")

	// (a) trusted RemoteAddr → header honored.
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:5555"
	r.Header.Set("X-Real-IP", "1.2.3.4")
	if ip := extractIP(r, trusted); ip != "1.2.3.4" {
		t.Fatalf("trusted proxy: extractIP = %q, want 1.2.3.4", ip)
	}

	// (b) untrusted RemoteAddr → header ignored, RemoteAddr used.
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "203.0.113.7:5555"
	r.Header.Set("X-Real-IP", "1.2.3.4")
	if ip := extractIP(r, trusted); ip != "203.0.113.7" {
		t.Fatalf("untrusted proxy: extractIP = %q, want 203.0.113.7", ip)
	}

	// (c) trusted RemoteAddr but garbage header → fall back to RemoteAddr.
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:5555"
	r.Header.Set("X-Real-IP", "not-an-ip")
	if ip := extractIP(r, trusted); ip != "10.0.0.1" {
		t.Fatalf("garbage header: extractIP = %q, want 10.0.0.1", ip)
	}

	// no trusted proxies configured → header always ignored.
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:5555"
	r.Header.Set("X-Real-IP", "1.2.3.4")
	if ip := extractIP(r, nil); ip != "10.0.0.1" {
		t.Fatalf("no trusted proxies: extractIP = %q, want 10.0.0.1", ip)
	}
}

func TestParseTrustedProxies(t *testing.T) {
	t.Run("valid entries", func(t *testing.T) {
		nets, err := ParseTrustedProxies([]string{"10.0.0.0/8", "192.168.1.5", "", "::1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(nets) != 3 {
			t.Fatalf("ParseTrustedProxies = %d nets, want 3 (CIDR + bare IPv4 + bare IPv6)", len(nets))
		}
		if !nets[1].Contains(net.ParseIP("192.168.1.5")) {
			t.Fatal("bare IP 192.168.1.5 should be matched as /32")
		}
		if nets[1].Contains(net.ParseIP("192.168.1.6")) {
			t.Fatal("bare IP 192.168.1.5 should NOT match a neighbor")
		}
	})

	t.Run("invalid entries error", func(t *testing.T) {
		nets, err := ParseTrustedProxies([]string{"10.0.0.0/8", "garbage"})
		if err == nil {
			t.Fatal("expected error for invalid entry, got nil")
		}
		if nets != nil {
			t.Fatalf("expected nil nets on error, got %v", nets)
		}
	})
}

func TestRateLimitDisabledWhenPerMinuteZero(t *testing.T) {
	l := NewIPLimiter(RateLimitConfig{PerMinute: 0})
	if l != nil {
		t.Fatalf("expected nil limiter when PerMinute <= 0, got %v", l)
	}
}

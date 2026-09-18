package state

import (
	"container/list"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimitConfig configures the per-IP token bucket.
//
// Budget math under the lazy-preview model: each pretty-view page-load
// debits 1 token for the SSR request + ~1 token per ref scrolled into
// view (fragment GETs). A page with N above-fold refs consumes 1+N
// tokens. With PerMinute=Burst=100 (default), a viewport-heavy page
// can consume a third of the budget in one paint — acceptable for a
// transitional defense-in-depth bucket. Primary HTTP rate-limit belongs
// to nginx; this limiter is the fallback when gnoweb is deployed alone.
//
// Zero value is the "disabled" mode: NewIPLimiter returns nil and the
// limiter check is a no-op, allowing Deps to default-zero without touching it.
type RateLimitConfig struct {
	PerMinute int // <=0 disables the limiter
	Burst     int // default = PerMinute
	MaxIPs    int // default 10_000
	// TrustedProxies is the set of trusted reverse-proxy networks. X-Real-IP
	// is honored ONLY when the connecting RemoteAddr falls inside one of
	// these CIDRs; empty = trust nothing (the safe default).
	TrustedProxies []*net.IPNet
	NowFunc        func() time.Time // injectable clock; default time.Now
}

const defaultMaxIPs = 10_000

// ParseTrustedProxies parses CIDR strings into networks for TrustedProxies.
// A bare IP (no mask) is accepted and treated as a /32 or /128. Empty
// entries are skipped, but any unparseable entry returns an error so a
// misconfigured proxy list fails loudly at startup rather than silently
// dropping entries (and quietly narrowing or widening trust).
func ParseTrustedProxies(cidrs []string) ([]*net.IPNet, error) {
	var (
		nets    []*net.IPNet
		invalid []string
	)
	for _, c := range cidrs {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if _, n, err := net.ParseCIDR(c); err == nil {
			nets = append(nets, n)
			continue
		}
		if ip := net.ParseIP(c); ip != nil {
			bits := 32
			if ip.To4() == nil {
				bits = 128
			}
			nets = append(nets, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
			continue
		}
		invalid = append(invalid, c)
	}
	if len(invalid) > 0 {
		return nil, fmt.Errorf("invalid trusted proxy entries: %s", strings.Join(invalid, ", "))
	}
	return nets, nil
}

// ipBucket is a single-IP token bucket; refills based on wall-clock delta.
type ipBucket struct {
	tokens   float64
	lastFill time.Time
	capacity float64
	rate     float64 // tokens per second
}

func (b *ipBucket) take(now time.Time) bool {
	elapsed := now.Sub(b.lastFill).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * b.rate
		if b.tokens > b.capacity {
			b.tokens = b.capacity
		}
		b.lastFill = now
	}
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// IPLimiter is a per-IP token-bucket map with LRU bounded eviction.
// Safe for concurrent use. Construct with NewIPLimiter.
type IPLimiter struct {
	now     func() time.Time
	rate    float64 // tokens per second derived from PerMinute
	burst   float64
	maxIPs  int
	mu      sync.Mutex
	buckets map[string]*ipBucket
	order   *list.List // front = most recently used, back = oldest
	elems   map[string]*list.Element
}

// NewIPLimiter constructs a limiter from cfg. Returns nil when cfg.PerMinute <= 0
// (caller treats nil as "open mode" — the limiter check is skipped).
func NewIPLimiter(cfg RateLimitConfig) *IPLimiter {
	if cfg.PerMinute <= 0 {
		return nil
	}
	if cfg.Burst <= 0 {
		cfg.Burst = cfg.PerMinute
	}
	if cfg.MaxIPs <= 0 {
		cfg.MaxIPs = defaultMaxIPs
	}
	now := cfg.NowFunc
	if now == nil {
		now = time.Now
	}
	return &IPLimiter{
		now:     now,
		rate:    float64(cfg.PerMinute) / 60.0,
		burst:   float64(cfg.Burst),
		maxIPs:  cfg.MaxIPs,
		buckets: make(map[string]*ipBucket),
		order:   list.New(),
		elems:   make(map[string]*list.Element),
	}
}

// Allow returns true if a request from ip is permitted right now.
// Side effect: refills then debits one token; touches LRU order; may evict.
func (l *IPLimiter) Allow(ip string) bool {
	if l == nil {
		return true
	}
	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()

	if b, ok := l.buckets[ip]; ok {
		if e, ok := l.elems[ip]; ok {
			l.order.MoveToFront(e)
		}
		return b.take(now)
	}

	if len(l.buckets) >= l.maxIPs {
		back := l.order.Back()
		if back != nil {
			oldIP, _ := back.Value.(string)
			l.order.Remove(back)
			delete(l.elems, oldIP)
			delete(l.buckets, oldIP)
		}
	}

	b := &ipBucket{
		tokens:   l.burst,
		lastFill: now,
		capacity: l.burst,
		rate:     l.rate,
	}
	l.buckets[ip] = b
	l.elems[ip] = l.order.PushFront(ip)
	return b.take(now)
}

// size returns the number of tracked IPs (test-only).
func (l *IPLimiter) size() int {
	if l == nil {
		return 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.buckets)
}

// has reports whether ip is currently tracked (test-only).
func (l *IPLimiter) has(ip string) bool {
	if l == nil {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	_, ok := l.buckets[ip]
	return ok
}

// extractIP returns the bucket key for r. X-Real-IP is honored only when the
// connecting RemoteAddr is inside a trusted-proxy network AND the header value
// is a valid IP; otherwise the RemoteAddr host is used. Falls back to the raw
// RemoteAddr string when SplitHostPort fails (e.g. unix sockets in tests).
func extractIP(r *http.Request, trusted []*net.IPNet) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if len(trusted) > 0 {
		if remoteIP := net.ParseIP(host); remoteIP != nil && ipInNets(remoteIP, trusted) {
			if v := net.ParseIP(strings.TrimSpace(r.Header.Get("X-Real-IP"))); v != nil {
				return v.String()
			}
		}
	}
	return host
}

func ipInNets(ip net.IP, nets []*net.IPNet) bool {
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

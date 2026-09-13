package playground

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Rate-limit defaults for the eval API: 10 tokens of burst, +1 token
// every 3s ≈ 20 req/min sustained per IP. Matches the legacy
// handler_playground.go behavior.
const (
	evalBurstSize      = 10
	evalRefillInterval = 3 * time.Second
)

// rateLimiter is a simple per-IP token bucket. Each IP gets burstSize
// tokens; one token is added every refillInterval.
type rateLimiter struct {
	mu             sync.Mutex
	buckets        map[string]*rateBucket
	burstSize      int
	refillInterval time.Duration
}

type rateBucket struct {
	tokens   int
	lastSeen time.Time
}

func newRateLimiter(burstSize int, refillInterval time.Duration) *rateLimiter {
	rl := &rateLimiter{
		buckets:        make(map[string]*rateBucket),
		burstSize:      burstSize,
		refillInterval: refillInterval,
	}
	// Prune stale buckets every minute.
	go rl.pruneLoop()
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[ip]
	if !ok {
		rl.buckets[ip] = &rateBucket{tokens: rl.burstSize - 1, lastSeen: now}
		return true
	}

	// Refill tokens based on elapsed time.
	elapsed := now.Sub(b.lastSeen)
	refill := int(elapsed / rl.refillInterval)
	if refill > 0 {
		b.tokens = min(rl.burstSize, b.tokens+refill)
		b.lastSeen = now
	}

	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

func (rl *rateLimiter) pruneLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-5 * time.Minute)
		for ip, b := range rl.buckets {
			if b.lastSeen.Before(cutoff) {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// clientIP extracts the real client IP, respecting X-Forwarded-For when
// present (first hop wins; downstream proxies should overwrite it).
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if ip, _, err := net.SplitHostPort(strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])); err == nil {
			return ip
		}
		return strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

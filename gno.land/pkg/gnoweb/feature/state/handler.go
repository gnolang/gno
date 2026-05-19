package state

import (
	"context"
	"net/http"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// pageTimeout bounds the page and JSON dispatch paths. The page path fans
// out up to ~17 RPC calls; without a per-request ceiling it would inherit
// only the gnoweb-global timeout (which may be unset). Fragments keep their
// own tighter fragmentTimeout (2 s) — see fragments.go.
const pageTimeout = 10 * time.Second

// Handle is the main entry for ?state* URLs. A nil view return means
// "body already written". The per-IP token-bucket runs first: on reject,
// htmx clients see a fragment-error (HTTP 200 + visible body); non-htmx
// clients get the standard 429 + Retry-After.
func (h *Handler) Handle(ctx context.Context, w http.ResponseWriter, r *http.Request, u *weburl.GnoURL) (int, *components.View) {
	if h.limiter != nil {
		ip := extractIP(r, h.deps.RateLimit.TrustedProxies)
		if !h.limiter.Allow(ip) {
			return writeRateLimited(w, r), nil
		}
	}
	switch {
	case u.WebQuery.Has("json"):
		ctx, cancel := context.WithTimeout(ctx, pageTimeout)
		defer cancel()
		return h.serveJSON(ctx, w, u), nil
	case u.WebQuery.Has("frag"):
		// serveFragment applies its own (tighter) fragmentTimeout.
		return h.serveFragment(ctx, w, u), nil
	default:
		ctx, cancel := context.WithTimeout(ctx, pageTimeout)
		defer cancel()
		return h.servePage(ctx, w, r, u)
	}
}

// writeRateLimited writes the rate-limit response: htmx clients get HTTP 200
// + a visible fragment-error body; plain clients get HTTP 429 + Retry-After.
// Body is written directly; the caller wraps the int status with a nil
// view so the wire-in skips chrome wrapping.
func writeRateLimited(w http.ResponseWriter, r *http.Request) int {
	if r.Header.Get("HX-Request") != "" {
		return writeFragError(w, "Rate limit exceeded", "Please slow down and retry in a moment.")
	}
	w.Header().Set("Retry-After", "60")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusTooManyRequests)
	_, _ = w.Write([]byte("rate limit exceeded\n"))
	return http.StatusTooManyRequests
}

// serveFragment lives in fragments.go; servePage lives in page.go.

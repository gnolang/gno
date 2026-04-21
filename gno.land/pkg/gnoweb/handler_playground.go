package gnoweb

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// playgroundRateLimiter is a simple per-IP token bucket limiter.
// Each IP gets burstSize tokens; one token is added every refillInterval.
type playgroundRateLimiter struct {
	mu             sync.Mutex
	buckets        map[string]*rateBucket
	burstSize      int
	refillInterval time.Duration
}

type rateBucket struct {
	tokens   int
	lastSeen time.Time
}

func newRateLimiter(burstSize int, refillInterval time.Duration) *playgroundRateLimiter {
	rl := &playgroundRateLimiter{
		buckets:        make(map[string]*rateBucket),
		burstSize:      burstSize,
		refillInterval: refillInterval,
	}
	// Prune stale buckets every minute.
	go rl.pruneLoop()
	return rl
}

func (rl *playgroundRateLimiter) allow(ip string) bool {
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

func (rl *playgroundRateLimiter) pruneLoop() {
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

// playgroundAPIHandler handles playground JSON API endpoints.
type playgroundAPIHandler struct {
	logger  *slog.Logger
	client  ClientAdapter
	domain  string
	remote  string
	limiter *playgroundRateLimiter
}

// evalRequest is the JSON request body for the eval endpoint.
type evalRequest struct {
	PkgPath    string `json:"pkg_path"`
	Expression string `json:"expression"`
}

// evalResponse is the JSON response for the eval endpoint.
type evalResponse struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

// funcsResponse represents function info returned by the funcs endpoint.
type funcsResponse struct {
	Functions []funcInfo `json:"functions"`
	PkgDoc    string     `json:"pkg_doc,omitempty"`
}

type funcInfo struct {
	Name      string      `json:"name"`
	Doc       string      `json:"doc,omitempty"`
	Signature string      `json:"signature"`
	Params    []paramInfo `json:"params,omitempty"`
	Crossing  bool        `json:"crossing"`
}

type paramInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// handlerPlaygroundEval creates an HTTP handler for expression evaluation.
// burstSize: max requests per IP before throttling; refillInterval: one token per interval.
func handlerPlaygroundEval(logger *slog.Logger, cli ClientAdapter, domain, remote string) http.Handler {
	h := &playgroundAPIHandler{
		logger:  logger,
		client:  cli,
		domain:  domain,
		remote:  remote,
		limiter: newRateLimiter(10, 3*time.Second), // 10 burst, +1 token every 3s ≈ 20 req/min
	}
	return http.HandlerFunc(h.serveEval)
}

// handlerPlaygroundFuncs creates an HTTP handler for listing functions.
func handlerPlaygroundFuncs(logger *slog.Logger, cli ClientAdapter) http.Handler {
	h := &playgroundAPIHandler{
		logger: logger,
		client: cli,
	}
	return http.HandlerFunc(h.serveFuncs)
}

// clientIP extracts the real client IP, respecting X-Forwarded-For if present.
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

func (h *playgroundAPIHandler) serveEval(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.limiter != nil && !h.limiter.allow(clientIP(r)) {
		writeJSON(w, http.StatusTooManyRequests, evalResponse{Error: "rate limit exceeded, please slow down"})
		return
	}

	var req evalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, evalResponse{Error: "invalid request body"})
		return
	}

	if req.PkgPath == "" || req.Expression == "" {
		writeJSON(w, http.StatusBadRequest, evalResponse{Error: "pkg_path and expression are required"})
		return
	}

	// Clean the pkg path.
	pkgPath := strings.TrimPrefix(req.PkgPath, h.domain+"/")
	pkgPath = strings.TrimPrefix(pkgPath, h.domain)
	pkgPath = strings.TrimPrefix(pkgPath, "/")

	// Build the qeval data string: "gno.land/r/demo/boards.Render("")"
	data := fmt.Sprintf("%s/%s.%s", h.domain, pkgPath, req.Expression)

	h.logger.Debug("playground eval", "data", data)

	start := time.Now()
	result, err := h.client.Eval(r.Context(), data)
	took := time.Since(start)

	h.logger.Debug("playground eval result", "took", took, "error", err)

	if err != nil {
		writeJSON(w, http.StatusOK, evalResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, evalResponse{Result: string(result)})
}

func (h *playgroundAPIHandler) serveFuncs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pkgPath := r.URL.Query().Get("path")
	if pkgPath == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "path parameter required"})
		return
	}

	jdoc, err := h.client.Doc(r.Context(), pkgPath)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"error": err.Error()})
		return
	}

	resp := funcsResponse{
		PkgDoc:    jdoc.PackageDoc,
		Functions: make([]funcInfo, 0, len(jdoc.Funcs)),
	}

	for _, fn := range jdoc.Funcs {
		if fn.Type != "" { // Skip methods
			continue
		}

		fi := funcInfo{
			Name:      fn.Name,
			Doc:       fn.Doc,
			Signature: fn.Signature,
			Crossing:  fn.Crossing,
			Params:    make([]paramInfo, 0, len(fn.Params)),
		}

		for _, p := range fn.Params {
			fi.Params = append(fi.Params, paramInfo{
				Name: p.Name,
				Type: p.Type,
			})
		}

		resp.Functions = append(resp.Functions, fi)
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

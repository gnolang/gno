package playground

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"time"
	"unicode/utf8"
)

const (
	// maxEvalBodyBytes caps the eval request body to guard against memory
	// exhaustion from oversized JSON payloads.
	maxEvalBodyBytes = 1 << 20 // 1 MiB

	// maxEvalPkgPathLen and maxEvalExpressionLen bound the eval request fields
	// well below maxEvalBodyBytes, rejecting oversized requests before they are
	// forwarded to the RPC node. Legitimate values are far smaller.
	maxEvalPkgPathLen    = 1024
	maxEvalExpressionLen = 64 * 1024

	// maxEvalResultBytes caps the result returned to the reader, far above
	// what a preview shows; the rest is cut with a visible marker.
	maxEvalResultBytes = 64 << 10 // 64 KiB

	// evalCacheTTL is how long an answer is reused. The chain moves one block
	// at a time, so a few seconds keeps previews current while every reader
	// of a popular Actions page costs the node one query per function.
	evalCacheTTL = 3 * time.Second

	// evalCacheSize bounds the cached answers; with the result and
	// expression caps above, it bounds the cache to a few tens of MiB.
	evalCacheSize = 256

	// defaultEvalTimeout bounds an eval when Deps.Timeout is not set; it is
	// gnoweb's default page deadline.
	defaultEvalTimeout = 30 * time.Second
)

// The errors a ClientAdapter wraps when an eval failed for a reason that
// says nothing about the expression. serveEval answers them with a generic
// message, never the transport detail.
var (
	ErrEvalUnavailable = errors.New("the node is unavailable, try again later")
	ErrEvalTimeout     = errors.New("the node did not answer the evaluation in time")
)

// EvalHandler returns the http.Handler for POST /_/api/eval.
func (h *Handler) EvalHandler() http.Handler {
	return http.HandlerFunc(h.serveEval)
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

// serveEval answers POST /_/api/eval. An expression that fails on the chain
// is a 200 with {error}: the request was served, its answer is the error.
// Every other status says gnoweb or the node could not answer.
func (h *Handler) serveEval(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, evalResponse{Error: "method not allowed"})
		return
	}

	// A JSON body makes a cross-site fetch need a preflight gnoweb never
	// approves. Origin is not read: a proxy may rewrite the Host it would match.
	if mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type")); mt != "application/json" {
		writeJSON(w, http.StatusUnsupportedMediaType, evalResponse{Error: "the body must be application/json"})
		return
	}
	switch r.Header.Get("Sec-Fetch-Site") {
	case "", "same-origin", "none":
	default:
		writeJSON(w, http.StatusForbidden, evalResponse{Error: "cross-origin request refused"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxEvalBodyBytes)

	var req evalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if errors.As(err, new(*http.MaxBytesError)) {
			writeJSON(w, http.StatusRequestEntityTooLarge, evalResponse{Error: "request body is too large"})
			return
		}
		writeJSON(w, http.StatusBadRequest, evalResponse{Error: "invalid request body"})
		return
	}

	if req.PkgPath == "" || req.Expression == "" {
		writeJSON(w, http.StatusBadRequest, evalResponse{Error: "pkg_path and expression are required"})
		return
	}

	if len(req.PkgPath) > maxEvalPkgPathLen || len(req.Expression) > maxEvalExpressionLen {
		writeJSON(w, http.StatusBadRequest, evalResponse{Error: "pkg_path or expression is too long"})
		return
	}

	pkgPath := RelPath(h.deps.Domain, req.PkgPath)
	key := evalKey(pkgPath, req.Expression)

	// Only a query that reaches the node is charged: a cached answer costs
	// nothing, so readers behind one address (a proxy gnoweb is not told
	// to trust) do not drain their shared budget on a popular page.
	if _, ok := h.cachedEval(key); !ok && h.deps.Limiter != nil && !h.deps.Limiter.AllowRequest(r) {
		writeJSON(w, http.StatusTooManyRequests, evalResponse{Error: "rate limit exceeded, please slow down"})
		return
	}

	h.deps.Logger.Debug("playground eval", "pkg_path", pkgPath)

	ctx, cancel := context.WithTimeout(r.Context(), h.deps.Timeout)
	defer cancel()

	start := time.Now()
	result, err := h.eval(ctx, key, pkgPath, req.Expression)
	took := time.Since(start)

	h.deps.Logger.Debug("playground eval result", "took", took, "error", err)

	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, evalResponse{Result: result})
	case errors.Is(err, ErrEvalTimeout):
		writeJSON(w, http.StatusGatewayTimeout, evalResponse{Error: ErrEvalTimeout.Error()})
	case errors.Is(err, ErrEvalUnavailable):
		writeJSON(w, http.StatusBadGateway, evalResponse{Error: ErrEvalUnavailable.Error()})
	default:
		writeJSON(w, http.StatusOK, evalResponse{Error: err.Error()})
	}
}

// evalAnswer is what the node answered for one expression.
type evalAnswer struct {
	result  string
	err     error
	expires time.Time
}

// eval answers expr in pkgPath, reusing an answer younger than evalCacheTTL
// and sharing one node query between the readers asking at the same time.
// The query runs detached from the reader who started it, under its own
// deadline, so a closed tab does not fail the readers waiting on it.
func (h *Handler) eval(ctx context.Context, key, pkgPath, expr string) (string, error) {
	if a, ok := h.cachedEval(key); ok {
		return a.result, a.err
	}

	ch := h.evalGroup.DoChan(key, func() (v any, _ error) {
		// DoChan runs this on a goroutine of its own, where a panic would
		// stop gnoweb: it is an answer the node could not give instead.
		defer func() {
			if p := recover(); p != nil {
				h.deps.Logger.Error("playground eval panicked", "pkg_path", pkgPath, "panic", p)
				v = evalAnswer{err: fmt.Errorf("%w: eval panicked", ErrEvalUnavailable)}
			}
		}()
		qctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), h.deps.Timeout)
		defer cancel()
		// The client adds the chain domain and joins pkgPath and the
		// expression as vm/qeval expects: "gno.land/r/demo/boards.Render("")".
		res, err := h.deps.Client.Eval(qctx, pkgPath, expr)
		a := evalAnswer{result: capEvalResult(res), err: err, expires: time.Now().Add(evalCacheTTL)}
		// Only the chain's answer is kept: a node that could not answer
		// may answer the next reader.
		if !errors.Is(err, ErrEvalUnavailable) && !errors.Is(err, ErrEvalTimeout) {
			h.evals.Add(key, a)
		}
		return a, nil
	})

	select {
	case res := <-ch:
		a := res.Val.(evalAnswer)
		return a.result, a.err
	case <-ctx.Done():
		return "", fmt.Errorf("%w: %w", ErrEvalTimeout, ctx.Err())
	}
}

// evalKey is the cache key of expr in pkgPath. The length prefix keeps two
// different pairs from sharing a key.
func evalKey(pkgPath, expr string) string {
	return fmt.Sprintf("%d:%s%s", len(pkgPath), pkgPath, expr)
}

// cachedEval is the answer for key younger than evalCacheTTL, if any.
func (h *Handler) cachedEval(key string) (evalAnswer, bool) {
	a, ok := h.evals.Get(key)
	return a, ok && time.Now().Before(a.expires)
}

// capEvalResult cuts a result longer than maxEvalResultBytes on a rune
// boundary and says how much was cut.
func capEvalResult(b []byte) string {
	if len(b) <= maxEvalResultBytes {
		return string(b)
	}
	n := maxEvalResultBytes
	for n > 0 && !utf8.RuneStart(b[n]) {
		n--
	}
	return fmt.Sprintf("%s\n… (truncated: first %d of %d bytes shown)", b[:n], n, len(b))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

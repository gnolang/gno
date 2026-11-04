package gnoweb

import (
	"context"
	"crypto/rand"
	"encoding/base64"
)

// cspNonceCtxKey is the context key under which the per-request Content-Security-Policy
// nonce is stored. It is unexported so only this package can set or read it.
type cspNonceCtxKey struct{}

// NewCSPNonce returns a fresh, cryptographically random nonce suitable for use in a
// Content-Security-Policy header and in the matching <meta name="csp-nonce"> tag.
// A new value must be generated for every response; reusing a nonce defeats its purpose.
func NewCSPNonce() string {
	var b [16]byte
	// crypto/rand.Read never returns an error on the platforms we support.
	_, _ = rand.Read(b[:])
	return base64.StdEncoding.EncodeToString(b[:])
}

// WithCSPNonce returns a copy of ctx carrying the given CSP nonce so that downstream
// handlers can embed it in the rendered HTML.
func WithCSPNonce(ctx context.Context, nonce string) context.Context {
	return context.WithValue(ctx, cspNonceCtxKey{}, nonce)
}

// CSPNonceFromContext returns the CSP nonce stored in ctx, or "" when none is set
// (for example when the server runs without strict security headers).
func CSPNonceFromContext(ctx context.Context) string {
	nonce, _ := ctx.Value(cspNonceCtxKey{}).(string)
	return nonce
}

package playground

import (
	"context"
	"log/slog"
	"net/http"
)

// ClientAdapter is the subset of the gnoweb chain-client interface that
// the playground feature consumes. Declared locally so feature/playground
// does not import the gnoweb package. The signatures match the
// corresponding methods on gnoweb.ClientAdapter.
type ClientAdapter interface {
	// ListFiles lists all source files available in a specified package path.
	ListFiles(ctx context.Context, path string) ([]string, error)

	// File fetches the source file from a given package path and filename.
	File(ctx context.Context, path, filename string) ([]byte, error)

	// Eval evaluates expr inside pkgPath through vm/qeval. pkgPath carries
	// no domain: the client adds its own.
	Eval(ctx context.Context, pkgPath, expr string) ([]byte, error)
}

// Limiter is the admission check applied to each eval request. It takes the
// request, not an address: which address to trust depends on which proxies
// to trust, and that rule belongs with the limiter.
type Limiter interface {
	AllowRequest(r *http.Request) bool
}

// Deps gathers the dependencies the playground Handler needs.
type Deps struct {
	Client ClientAdapter
	Logger *slog.Logger

	// Domain is the chain domain (e.g. "gno.land").
	Domain string

	// Limiter bounds the per-client eval rate. Optional; nil disables the
	// check.
	Limiter Limiter
}

// Handler owns the playground feature state.
type Handler struct {
	deps Deps
}

// New validates required deps and returns a Handler.
// If empty Domain defaults to "gno.land" and Logger defaults the
// standard Go library's logger.
// It panics if Client is not specified.
func New(deps Deps) *Handler {
	if deps.Client == nil {
		panic("playground.New: Client is required")
	}

	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}

	if deps.Domain == "" {
		deps.Domain = "gno.land"
	}

	return &Handler{deps: deps}
}

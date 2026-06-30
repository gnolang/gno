package state

import (
	"context"
	"log/slog"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

// ClientAdapter is the subset of gnoweb.ClientAdapter the state handler
// consumes. Declared locally so feature/state does not import the
// gnoweb package (handler_http.go in gnoweb imports state — a back
// import would create a cycle). Method set is a subset of
// gnoweb.ClientAdapter so a *gnoweb.MockClient or *gnoweb.rpcClient
// satisfies this contract.
type ClientAdapter interface {
	Realm(ctx context.Context, path, args string) ([]byte, error)
	ListPaths(ctx context.Context, prefix string, limit int) ([]string, error)
	Doc(ctx context.Context, path string, height int64) (*doc.JSONDocumentation, error)
	StatePkg(ctx context.Context, path string, height int64) ([]byte, error)
	StateObject(ctx context.Context, oid string, height int64) ([]byte, error)
	StateType(ctx context.Context, typeId string, height int64) ([]byte, error)
}

// Deps is a struct of interfaces so each field is independently mockable.
type Deps struct {
	Client      ClientAdapter
	Highlighter components.SnippetHighlighter
	// FileFetcher reads one source file by (pkgPath, fileName). Optional:
	// when nil, frag=source returns a fragment-error pointing at the
	// permanent ?source link. Wrapped per-request because the local
	// ClientAdapter cannot carry FileMeta without an import cycle.
	FileFetcher components.FileFetcher
	Logger      *slog.Logger
	// RateLimit configures the per-IP token bucket. Zero value disables it;
	// the limiter check in Handle becomes a no-op.
	RateLimit RateLimitConfig
}

type Handler struct {
	deps    Deps
	limiter *IPLimiter
}

// New validates required deps and returns a Handler.
// Panics if Client or Highlighter is nil; Logger falls back to slog.Default().
func New(deps Deps) *Handler {
	if deps.Client == nil {
		panic("state.New: Client is required")
	}
	if deps.Highlighter == nil {
		panic("state.New: Highlighter is required")
	}
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}
	return &Handler{deps: deps, limiter: NewIPLimiter(deps.RateLimit)}
}

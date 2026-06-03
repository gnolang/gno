package playground

import (
	"context"
	"log/slog"

	"github.com/gnolang/gno/gnovm/pkg/doc"
)

// ClientAdapter is the subset of the gnoweb chain-client interface that
// the playground feature consumes. Declared locally so feature/playground
// does not import the gnoweb package. The signatures match the
// corresponding methods on gnoweb.ClientAdapter so a *gnoweb.MockClient
// or *gnoweb.rpcClient satisfies this contract through a thin adapter
// wired in at construction time.
type ClientAdapter interface {
	// ListFiles is used by the fork view to enumerate package sources.
	ListFiles(ctx context.Context, path string) ([]string, error)

	// File is used by the fork view to read each source file.
	File(ctx context.Context, path, filename string) ([]byte, error)

	// Doc is used by the funcs API to enumerate exported functions.
	Doc(ctx context.Context, path string) (*doc.JSONDocumentation, error)

	// Eval is used by the eval API to run an expression against a
	// realm via vm/qeval.
	Eval(ctx context.Context, data string) ([]byte, error)
}

// Deps gathers the dependencies the playground Handler needs.
type Deps struct {
	Client ClientAdapter

	// Logger falls back to slog.Default().
	Logger *slog.Logger

	// Domain is the chain domain (e.g. "gno.land").
	Domain string

	// Remote is the RPC endpoint surfaced to the playground UI so it
	// can show the user which node it is talking to.
	Remote string

	// ChainId is the active chain id surfaced to the playground UI.
	ChainId string
}

// Handler owns the playground feature state.
type Handler struct {
	deps    Deps
	limiter *rateLimiter
}

// New validates required deps and returns a Handler.
func New(deps Deps) *Handler {
	if deps.Client == nil {
		panic("playground.New: Client is required")
	}

	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}

	return &Handler{
		deps:    deps,
		limiter: newRateLimiter(evalBurstSize, evalRefillInterval),
	}
}

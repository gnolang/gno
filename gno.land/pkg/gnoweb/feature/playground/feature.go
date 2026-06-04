package playground

import (
	"context"
	"log/slog"

	"github.com/gnolang/gno/gnovm/pkg/doc"
)

// ClientAdapter is the subset of the gnoweb chain-client interface that
// the playground feature consumes. Declared locally so feature/playground
// does not import the gnoweb package. The signatures match the
// corresponding methods on gnoweb.ClientAdapter.
type ClientAdapter interface {
	// ListFiles lists all source files available in a specified package path.
	ListFiles(ctx context.Context, path string) ([]string, error)

	// File fetche the source file from a given package path and filename.
	File(ctx context.Context, path, filename string) ([]byte, error)

	// Doc retrieves the JSON doc suitable for printing from a
	// specified package path.
	Doc(ctx context.Context, path string) (*doc.JSONDocumentation, error)

	// Eval evaluates a Gno expression via vm/qeval query.
	// The data string should be in the format "gno.land/r/pkg.Expression(args)".
	Eval(ctx context.Context, data string) ([]byte, error)
}

// Deps gathers the dependencies the playground Handler needs.
type Deps struct {
	Client ClientAdapter
	Logger *slog.Logger

	// Domain is the chain domain (e.g. "gno.land").
	Domain string

	// Remote is the RPC endpoint.
	Remote string

	// ChainId is the active chain ID.
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

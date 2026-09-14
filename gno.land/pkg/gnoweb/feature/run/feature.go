package run

import "log/slog"

// Deps gathers the dependencies the run Handler needs.
type Deps struct {
	Logger *slog.Logger

	// Domain is the chain domain (e.g. "gno.land"), used to build the
	// fully-qualified package import path stamped on the rendered page.
	Domain string

	// Remote is the RPC endpoint.
	Remote string

	// ChainId is the active chain ID.
	ChainId string
}

// Handler owns the run feature state.
type Handler struct {
	deps Deps
}

// New returns a Run handler.
// If empty Domain defaults to "gno.land" and Logger defaults the
// standard Go library's logger.
// It panics if Remote or ChainId are not specified.
func New(deps Deps) *Handler {
	if deps.Remote == "" {
		panic("run.New: Remote RPC endpoint is required")
	}

	if deps.ChainId == "" {
		panic("run.New: Chain ID is required")
	}

	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}

	if deps.Domain == "" {
		deps.Domain = "gno.land"
	}

	return &Handler{deps: deps}
}

package run

import "log/slog"

// Deps gathers the dependencies the run Handler needs.
type Deps struct {
	// Logger falls back to slog.Default().
	Logger *slog.Logger

	// Domain is the chain domain (e.g. "gno.land"), used to build the
	// fully-qualified package import path stamped on the rendered page.
	Domain string

	// Remote is the RPC endpoint surfaced to the run UI so the CLI
	// command builder can default it.
	Remote string

	// ChainId is the active chain id surfaced to the run UI.
	ChainId string
}

// Handler owns the run feature state.
type Handler struct {
	deps Deps
}

// New returns a Run handler.
func New(deps Deps) *Handler {
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}
	return &Handler{deps: deps}
}

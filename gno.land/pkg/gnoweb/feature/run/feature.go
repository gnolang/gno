package run

import "html/template"

// Highlighter renders source code as highlighted HTML.
type Highlighter interface {
	Render(fileName string, source []byte) (template.HTML, error)
}

// Deps gathers the dependencies the run Handler needs.
type Deps struct {
	// Domain is the chain domain (e.g. "gno.land"), used to build the
	// fully-qualified package import path stamped on the rendered page.
	Domain string

	// Remote is the RPC endpoint the printed commands talk to.
	Remote string

	// ChainId is the chain ID the printed commands sign for.
	ChainId string

	// Highlighter renders the script as served, before the editor loads.
	// Optional: without it the script is served as plain text.
	Highlighter Highlighter
}

// Handler owns the run feature state.
type Handler struct {
	deps Deps
}

// New returns a Run handler. An empty Domain defaults to "gno.land".
func New(deps Deps) *Handler {
	if deps.Domain == "" {
		deps.Domain = "gno.land"
	}

	return &Handler{deps: deps}
}

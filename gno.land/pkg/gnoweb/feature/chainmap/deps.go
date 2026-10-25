package chainmap

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

// depsWait is how long the dependencies page waits for the importers before
// saying they are still being looked up. The lookup carries on detached, so a
// reload finds it done.
const depsWait = 12 * time.Second

// DepsData is the render payload for templates/deps.html.
type DepsData struct {
	// Title names the package in the heading, "ufmt/v0" for a versioned one.
	Title string
	Graph components.DepGraph
	// Indexer is the provenance footer, set when the indexer answered.
	Indexer *components.IndexerStatus
}

// DepsView renders the dependency graph of pkgPath, a gnoweb-relative path:
// its imports from the chain and, with an indexer, its importers. It is a page
// of its own because the importers cost a whole-chain scan plus a node read
// per candidate; the overview only links to it.
//
// The error is the chain's, from reading pkgPath's imports, for the caller to
// render as it renders any other read failure.
func (h *Handler) DepsView(r *http.Request, pkgPath, title string) (int, *components.View, error) {
	ctx := r.Context()
	full := h.deps.Domain + pkgPath

	// The package is read first: a path with no live package must not cost a
	// whole-chain scan, nor wait for one before its 404.
	imports, err := h.imports(ctx, full)
	if err != nil {
		return 0, nil, err
	}

	data := DepsData{
		Title: title,
		Graph: components.DepGraph{Name: pkgPath, Imports: components.ImportLinks(imports, h.deps.Domain)},
	}
	status := http.StatusOK
	if h.HasIndexer() {
		wctx, cancel := context.WithTimeout(ctx, depsWait)
		defer cancel()
		importers, err := h.Importers(wctx, r, full)
		data.Graph.Importers, status = h.importersSide(pkgPath, importers, err)
		if err == nil {
			data.Indexer = &components.IndexerStatus{URL: h.deps.Indexer.URL(), LastBlock: importers.AsOf}
		}
	}

	view := &components.View{Type: components.DepsViewType, Component: &pageComponent{name: "renderDeps", data: data}}
	view.SkipTargetInBody = true // on the content header
	return status, view, nil
}

// importersSide turns a lookup into what the graph shows, saying why when it
// shows none.
func (h *Handler) importersSide(pkgPath string, imp *Importers, err error) (*components.Importers, int) {
	switch {
	case err == nil:
		return &components.Importers{
			Links:  components.ImportLinks(imp.Paths, h.deps.Domain),
			Capped: imp.Capped,
			Retry:  imp.Unread,
		}, http.StatusOK
	case errors.Is(err, ErrRateLimited):
		return &components.Importers{Unavailable: "Too many lookups from your address. Try again in a minute."}, http.StatusTooManyRequests
	case errors.Is(err, ErrPending):
		return &components.Importers{Unavailable: "Still looking them up. Reload in a moment."}, http.StatusOK
	case errors.Is(err, ErrBusy):
		return &components.Importers{Unavailable: "Other lookups are running. Reload in a moment."}, http.StatusServiceUnavailable
	default:
		h.deps.Logger.Warn("deps: importers unavailable", "path", pkgPath, "error", err)
		return &components.Importers{Unavailable: "The indexer could not be read."}, http.StatusOK
	}
}

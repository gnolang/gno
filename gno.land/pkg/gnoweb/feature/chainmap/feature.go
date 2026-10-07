package chainmap

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

// Indexer is the subset of *indexer.Client this feature consumes. A nil
// Indexer is the feature switch: the map then shows no activity and the
// dependency graph no importers, and neither says anything about them.
type Indexer interface {
	LatestBlockHeight(ctx context.Context) (int, error)
	Block(ctx context.Context, height int) (*indexer.Block, error)
	CallsBetween(ctx context.Context, lower, upper int) ([]indexer.Tx, error)
	// DeploysQuoting returns the deploys in the heights (lower, upper] whose
	// source holds pkgPath as a Go string literal, delimiters included, so
	// that a path never matches inside a longer one.
	DeploysQuoting(ctx context.Context, pkgPath string, lower, upper int) ([]indexer.Tx, error)
	URL() string
}

// ImportReader reads a live package's non-test imports from the chain. It
// returns ErrNotLive when no live package sits at the path.
type ImportReader interface {
	Imports(ctx context.Context, pkgPath string) ([]string, error)
}

// Limiter is the admission check applied before indexer-backed work a reader
// asked for. Nil disables the check.
type Limiter interface {
	AllowRequest(r *http.Request) bool
}

// Deps is a struct of interfaces so each field is independently mockable.
type Deps struct {
	// Indexer answers activity and proposes importers. Optional: nil means
	// no indexer is configured.
	Indexer Indexer

	// Imports reads a package's imports from the chain: the graph's left
	// side, and the check on every importer the indexer proposes. Required.
	Imports ImportReader

	// Domain is the chain's domain (e.g. "gno.land"), which turns a gnoweb
	// path into the package path the indexer keys on.
	Domain string

	// Limiter bounds the per-IP rate of importer lookups. Optional.
	Limiter Limiter

	Logger *slog.Logger
}

// Handler renders maps and answers the indexer-backed queries behind them.
type Handler struct {
	deps     Deps
	activity *flight[*Activity]
	// closedBands keeps the activity of whole bands across refreshes.
	closedBands *closedBands
	importers   importerFlights
}

// New validates required deps and returns a Handler. The caches exist only
// with an indexer, so none can be reached without one.
func New(deps Deps) *Handler {
	if deps.Imports == nil {
		panic("chainmap.New: Imports is required")
	}
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}
	h := &Handler{deps: deps}
	if deps.Indexer != nil {
		h.activity = newActivityFlight()
		h.closedBands = new(closedBands)
		h.importers = newImporterFlights()
	}
	return h
}

// HasIndexer reports whether the indexer-backed answers exist on this
// deployment.
func (h *Handler) HasIndexer() bool { return h.deps.Indexer != nil }

// DepsURL is the dependencies page of pkgPath, a gnoweb-relative path, or ""
// on a deployment that cannot list importers: without an indexer the
// overview's graph already shows all there is.
func (h *Handler) DepsURL(pkgPath string) string {
	if !h.HasIndexer() {
		return ""
	}
	return pkgPath + "$deps"
}

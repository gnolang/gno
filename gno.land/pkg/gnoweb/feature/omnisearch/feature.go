package omnisearch

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

// ClientAdapter is the subset of gnoweb.ClientAdapter this feature consumes,
// declared locally to avoid a back import (gnoweb imports omnisearch).
type ClientAdapter interface {
	Realm(ctx context.Context, path, args string) ([]byte, error)
	Doc(ctx context.Context, path string, height int64) (*doc.JSONDocumentation, error)
	ListFiles(ctx context.Context, path string, height int64) ([]string, error)
}

// Directory lists the chain's paths. Shaped after gnoweb's RealmDirectory so
// the wire-in passes the instance /search.json already uses, singleflight and
// semaphore included — qpaths is the most expensive RPC gnoweb makes.
type Directory interface {
	// Paths returns the realm and package paths, and whether the node capped
	// the listing. Truncation travels with the answer because a silent cap
	// is how a search comes to say "no such realm" about one that exists.
	Paths(ctx context.Context) (realms, packages []string, truncated bool, err error)
}

// Indexer is the subset of *indexer.Client this feature consumes. A nil
// Indexer is the feature switch: New registers the indexer-backed selectors
// only when one is present.
type Indexer interface {
	LatestBlockHeight(ctx context.Context) (int, error)
	TxByHash(ctx context.Context, hash string) (*indexer.Tx, error)
	RecentByPackage(ctx context.Context, pkgPath string, limit int) ([]indexer.Tx, error)
	RecentByAddress(ctx context.Context, addr string, limit int) ([]indexer.Tx, error)
	Deploys(ctx context.Context, pkgPath string, limit int) ([]indexer.Tx, error)
	SourceContains(ctx context.Context, text string, limit int) ([]indexer.Tx, error)
	Block(ctx context.Context, height int) (*indexer.Block, error)
	URL() string
}

// Limiter is the admission check applied before any fetch. It takes the
// request, not an address: which address to trust depends on which proxies
// to trust, and that rule belongs with the limiter. Nil disables the check.
type Limiter interface {
	AllowRequest(r *http.Request) bool
}

// Deps is a struct of interfaces so each field is independently mockable.
type Deps struct {
	// Client answers the chain-backed selectors. Required.
	Client ClientAdapter

	// Directory lists realm and package paths for the discovery search.
	// Required.
	Directory Directory

	// Indexer answers the indexer-backed selectors. Optional: nil means no
	// indexer is configured and those selectors do not exist.
	Indexer Indexer

	// Domain is the chain's domain (e.g. "gno.land"), used to turn a gnoweb
	// path into the fully qualified package path the chain indexes under.
	Domain string

	// Limiter bounds the per-IP request rate. Optional; nil disables the
	// check.
	Limiter Limiter

	Logger *slog.Logger
}

// Handler owns every `$search` URL.
type Handler struct {
	deps      Deps
	selectors []*Selector
	byName    map[string]*Selector

	// docGroup coalesces qdoc: func:/type:/imports all read it, once per
	// debounced keystroke, so N readers on one realm would otherwise mean N
	// identical queries.
	docGroup singleflight.Group
}

// doc coalesces concurrent callers. The shared fetch is detached from
// whichever request started it, so a closed tab cannot cancel the others.
func (h *Handler) doc(ctx context.Context, pkgPath string) (*doc.JSONDocumentation, error) {
	ch := h.docGroup.DoChan(pkgPath, func() (any, error) {
		fetchCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), docTimeout)
		defer cancel()
		return h.deps.Client.Doc(fetchCtx, pkgPath, 0)
	})

	select {
	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		return res.Val.(*doc.JSONDocumentation), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// docTimeout bounds the shared fetch, independent of its starter.
const docTimeout = 5 * time.Second

// New validates required deps and returns a Handler. The selector table is
// built once: which selectors exist is a property of the deployment.
func New(deps Deps) *Handler {
	if deps.Client == nil {
		panic("omnisearch.New: Client is required")
	}
	if deps.Directory == nil {
		panic("omnisearch.New: Directory is required")
	}
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}

	h := &Handler{deps: deps, byName: make(map[string]*Selector)}
	h.register(SourceChain, chainSelectors()...)
	if deps.Indexer != nil {
		h.register(SourceIndexer, indexerSelectors()...)
	}
	return h
}

// register stamps provenance from the table a selector is in, so no literal
// can claim the wrong one by copy-paste.
func (h *Handler) register(src Source, sels ...*Selector) {
	for _, s := range sels {
		s.Source = src
		h.selectors = append(h.selectors, s)
		h.byName[s.Name] = s
	}
}

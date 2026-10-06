package chainmap

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

const (
	// importersTTL is how long a package's importers are served. A new
	// importer needs a deploy, and the page states the block it is as of.
	importersTTL = 10 * time.Minute

	// importersTimeout bounds one lookup: the whole-chain source scan took
	// about 5 s on gnoland-1 in October 2026, and the checks run after it.
	importersTimeout = 30 * time.Second

	// maxCandidates bounds the chain reads one lookup makes. Past it the
	// answer is "at least": gno.land/p/nt/avl/v0 had 265 candidates.
	maxCandidates = 400

	// checkConcurrency bounds the qdoc reads in flight for one lookup.
	checkConcurrency = 8

	// maxImporterEntries bounds the cache, so a crawler walking every
	// package cannot grow it without limit.
	maxImporterEntries = 512
)

var (
	// ErrRateLimited is returned when a reader asks for more lookups than
	// the limiter allows.
	ErrRateLimited = errors.New("too many requests")

	// ErrNotLive is what an ImportReader returns for a path with no live
	// package.
	ErrNotLive = errors.New("package is not live")
)

// Importers are the live packages whose non-test source imports a package.
//
// The indexer only proposes candidates: it can say which deploys quote the
// import path, not which package in a batched deploy did, nor whether the
// quote is an import, a test import or a string. Each candidate is then
// checked against the chain's own import list, so every path here is an
// import as of the chain's current state.
type Importers struct {
	// Paths are fully qualified and sorted.
	Paths []string
	// AtLeast is set when some candidates went unchecked: the indexer capped
	// its answer, there were more than maxCandidates, or time ran out.
	AtLeast bool
	// AsOf is the indexer's last block when the lookup ran.
	AsOf int
}

type importersEntry struct {
	imp *Importers
	at  time.Time
}

// importersCache memoizes lookups per package, bounded in size and age. Like
// activityCache it runs nothing in the background.
type importersCache struct {
	idx  Indexer
	docs ImportReader

	mu      sync.Mutex
	entries map[string]importersEntry

	group singleflight.Group
}

func newImportersCache(idx Indexer, docs ImportReader) *importersCache {
	return &importersCache{idx: idx, docs: docs, entries: make(map[string]importersEntry)}
}

// Importers answers which live packages import pkgPath, a fully qualified
// package path. The limiter is consulted only on a cache miss: a cached
// answer costs neither the indexer nor the node anything.
func (h *Handler) Importers(ctx context.Context, r *http.Request, pkgPath string) (*Importers, error) {
	if h.importers == nil {
		return nil, errors.New("no indexer configured")
	}
	if imp := h.importers.cached(pkgPath); imp != nil {
		return imp, nil
	}
	if h.deps.Limiter != nil && !h.deps.Limiter.AllowRequest(r) {
		return nil, ErrRateLimited
	}
	return h.importers.lookup(ctx, pkgPath)
}

func (c *importersCache) cached(pkgPath string) *Importers {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.entries[pkgPath]; ok && time.Since(e.at) < importersTTL {
		return e.imp
	}
	return nil
}

// lookup runs one shared, detached lookup per package and waits for it as
// long as ctx allows.
func (c *importersCache) lookup(ctx context.Context, pkgPath string) (*Importers, error) {
	ch := c.group.DoChan(pkgPath, func() (any, error) {
		lctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), importersTimeout)
		defer cancel()
		imp, err := c.find(lctx, pkgPath)
		if err != nil {
			return nil, err
		}
		c.store(pkgPath, imp)
		return imp, nil
	})
	select {
	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		return res.Val.(*Importers), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// store keeps imp, evicting the oldest entry when the cache is full.
func (c *importersCache) store(pkgPath string, imp *Importers) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.entries[pkgPath]; !ok && len(c.entries) >= maxImporterEntries {
		var oldest string
		var oldestAt time.Time
		for k, e := range c.entries {
			if oldest == "" || e.at.Before(oldestAt) {
				oldest, oldestAt = k, e.at
			}
		}
		delete(c.entries, oldest)
	}
	c.entries[pkgPath] = importersEntry{imp: imp, at: time.Now()}
}

// find gathers candidates from the indexer, then keeps those whose current
// import list on chain names pkgPath.
func (c *importersCache) find(ctx context.Context, pkgPath string) (*Importers, error) {
	asOf, err := c.idx.LatestBlockHeight(ctx)
	if err != nil {
		return nil, err
	}
	// The quotes keep gno.land/p/nt/avl/v0 from matching inside
	// gno.land/p/nt/avl/v0/rotree.
	txs, err := c.idx.DeploysQuoting(ctx, `"`+pkgPath+`"`)
	capped := errors.Is(err, indexer.ErrTooLarge)
	if err != nil && !capped {
		return nil, err
	}

	candidates := candidatesFrom(txs, pkgPath)
	imp := &Importers{AsOf: asOf, AtLeast: capped}
	if len(candidates) > maxCandidates {
		candidates, imp.AtLeast = candidates[:maxCandidates], true
	}

	var mu sync.Mutex
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(checkConcurrency)
	for _, cand := range candidates {
		g.Go(func() error {
			imports, err := c.docs.Imports(gctx, cand)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				if slices.Contains(imports, pkgPath) {
					imp.Paths = append(imp.Paths, cand)
				}
			case errors.Is(err, ErrNotLive):
				// Not live any more, or never was (a parked deploy): it
				// imports nothing today.
			default:
				// Unread, whether the node failed or time ran out: what was
				// checked stands, and the answer says there may be more.
				imp.AtLeast = true
			}
			return nil
		})
	}
	_ = g.Wait()
	slices.Sort(imp.Paths)
	return imp, nil
}

// candidatesFrom lists, once each and in a stable order, the packages the
// matched deploys added, other than pkgPath itself.
func candidatesFrom(txs []indexer.Tx, pkgPath string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, tx := range txs {
		for _, m := range tx.Messages {
			p := m.Path()
			if m.Type() != "MsgAddPackage" || p == "" || p == pkgPath {
				continue
			}
			if _, dup := seen[p]; dup {
				continue
			}
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}
	slices.Sort(out)
	return out
}

package chainmap

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

const (
	// importersTTL is how long an answer, and a candidate's imports, are
	// served. A new importer needs a deploy, and the page states the block
	// it is as of.
	importersTTL = 10 * time.Minute

	// importersTimeout bounds one lookup. On gnoland-1 in October 2026 the
	// scan took about 5 s and checking gno.land/p/nt/avl/v0's 265 candidates
	// another 38 s (qdoc averaged 1.1 s on the public RPC). A reader waits
	// depsWait at most; the lookup finishes for the next one.
	importersTimeout = 60 * time.Second

	// scanBand is the height span of one candidate query. A whole-chain scan
	// took 4.9 s on gnoland-1 in October 2026, past the indexer client's 4 s
	// request timeout; bands of this width took 0.6 to 1.6 s unloaded. The
	// margin matters: that client's breaker counts timeouts, and an open
	// breaker turns off search for every reader too.
	scanBand = 75_000

	// scanConcurrency bounds the band queries in flight for one lookup.
	scanConcurrency = 2

	// maxLookups bounds the lookups running at once across all readers. Each
	// is a whole-chain scan plus up to maxCandidates node reads; the per-IP
	// limiter alone lets many addresses run them side by side.
	maxLookups = 2

	// maxCandidates bounds the chain reads one lookup makes. Past it the
	// answer is "at least": gno.land/p/nt/avl/v0 had 265 candidates.
	maxCandidates = 400

	// checkConcurrency bounds the qdoc reads in flight for one lookup.
	checkConcurrency = 8

	// importReadTimeout bounds one candidate's qdoc read.
	importReadTimeout = 10 * time.Second

	// maxImporterEntries and maxImportEntries bound the caches, so a crawler
	// walking every package cannot grow them without limit.
	maxImporterEntries = 512
	maxImportEntries   = 4096
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
	// a band, there were more than maxCandidates, or a read failed.
	AtLeast bool
	// unread is set when a candidate read failed or ran out of time, which a
	// retry can complete, unlike a cap.
	unread bool
	// AsOf is the indexer's last block when the lookup ran.
	AsOf int
}

// importerFlights hold the answers per package, and each candidate's imports
// so that lookups for packages sharing importers read them once.
type importerFlights struct {
	answers *flight[*Importers]
	imports *flight[[]string]
	// slots admits maxLookups lookups at a time.
	slots chan struct{}
}

func newImporterFlights() importerFlights {
	return importerFlights{
		slots: make(chan struct{}, maxLookups),
		answers: &flight[*Importers]{
			ttl: importersTTL, errTTL: failureTTL, timeout: importersTimeout, max: maxImporterEntries,
			partial: func(imp *Importers) bool { return imp != nil && imp.unread },
		},
		// A failure is not remembered: ErrNotLive turns into imports the
		// moment a parked deploy is approved.
		imports: &flight[[]string]{ttl: importersTTL, timeout: importReadTimeout, max: maxImportEntries},
	}
}

// Importers answers which live packages import pkgPath, a fully qualified
// package path. The limiter is consulted only when the answer is not held: a
// held answer costs neither the indexer nor the node anything.
func (h *Handler) Importers(ctx context.Context, r *http.Request, pkgPath string) (*Importers, error) {
	if h.importers.answers == nil {
		return nil, errors.New("no indexer configured")
	}
	if e, ok := h.importers.answers.fresh(pkgPath); ok {
		return e.val, e.err
	}
	if h.deps.Limiter != nil && !h.deps.Limiter.AllowRequest(r) {
		return nil, ErrRateLimited
	}
	return h.importers.answers.get(ctx, pkgPath, func(ctx context.Context) (*Importers, error) {
		return h.findImporters(ctx, pkgPath)
	})
}

// findImporters gathers candidates from the indexer, then keeps those whose
// current import list on chain names pkgPath.
func (h *Handler) findImporters(ctx context.Context, pkgPath string) (*Importers, error) {
	select {
	case h.importers.slots <- struct{}{}:
		defer func() { <-h.importers.slots }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	asOf, err := h.deps.Indexer.LatestBlockHeight(ctx)
	if err != nil {
		return nil, err
	}
	candidates, capped, err := h.scanCandidates(ctx, pkgPath, asOf)
	if err != nil {
		return nil, err
	}

	imp := &Importers{AsOf: asOf, AtLeast: capped}
	if len(candidates) > maxCandidates {
		candidates, imp.AtLeast = candidates[:maxCandidates], true
	}

	var (
		mu sync.Mutex
		g  errgroup.Group
	)
	g.SetLimit(checkConcurrency)
	for _, cand := range candidates {
		if ctx.Err() != nil {
			// Out of time: the rest stay unread, and the answer says so.
			mu.Lock()
			imp.AtLeast, imp.unread = true, true
			mu.Unlock()
			break
		}
		g.Go(func() error {
			imports, err := h.importers.imports.get(ctx, cand, func(ctx context.Context) ([]string, error) {
				return h.deps.Imports.Imports(ctx, cand)
			})
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
				imp.AtLeast, imp.unread = true, true
			}
			return nil
		})
	}
	_ = g.Wait()
	slices.Sort(imp.Paths)
	return imp, nil
}

// scanCandidates lists, once each and sorted, the packages added by every
// deploy quoting pkgPath as a string literal, band by band up to tip. capped
// reports a band the indexer cut short. The first failing band cancels the
// others: the answer is an error either way.
func (h *Handler) scanCandidates(ctx context.Context, pkgPath string, tip int) (paths []string, capped bool, err error) {
	var mu sync.Mutex
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(scanConcurrency)
	// The bottom band starts below 0: genesis packages live at height 0.
	for lower := -1; lower < tip; lower += scanBand {
		upper := min(lower+scanBand, tip)
		g.Go(func() error {
			txs, err := h.deps.Indexer.DeploysQuoting(gctx, pkgPath, lower, upper)
			bandCapped := errors.Is(err, indexer.ErrTooLarge)
			if err != nil && !bandCapped {
				return err
			}
			mu.Lock()
			defer mu.Unlock()
			capped = capped || bandCapped
			for _, tx := range txs {
				for _, m := range tx.Messages {
					if p := m.Path(); m.Type() == "MsgAddPackage" && p != "" && p != pkgPath {
						paths = append(paths, p)
					}
				}
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, false, err
	}
	slices.Sort(paths)
	return slices.Compact(paths), capped, nil
}

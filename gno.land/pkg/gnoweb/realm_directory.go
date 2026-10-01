package gnoweb

import (
	"context"
	"path"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// pathLister is the subset of ClientAdapter the directory depends on.
type pathLister interface {
	ListPaths(ctx context.Context, prefix string, limit int) ([]string, error)
}

// PathsResult is a directory listing, plus whether the node capped it.
type PathsResult struct {
	Realms   []string
	Packages []string

	// Truncated reports that a listing came back at the cap, so the chain
	// holds paths this result does not name. It is part of the answer: a
	// silent cap is how a search comes to say "no such realm" about a realm
	// that exists, and the lexicographic order of the underlying iterator
	// means the paths dropped are always the same ones.
	Truncated bool
}

// RealmDirectory exposes realm and package paths for discovery. It is the seam
// behind which the source can evolve (live RPC today, a dedicated search index
// later) without touching callers.
type RealmDirectory interface {
	// Paths returns the realm (/r/) and package (/p/) paths known to the chain.
	Paths(ctx context.Context) (PathsResult, error)
}

var _ RealmDirectory = (*rpcRealmDirectory)(nil)

// searchPathLimit is the requested per-prefix page size, and it is now
// actually forwarded to the node (it was not: `limit` was a dead parameter,
// so the node's 1000 default governed silently).
//
// 10000 is the node's own ceiling — pathsLimit clamps to it — so this asks
// for everything a single qpaths call can return. Past that the answer is
// truncated and says so; going further needs cursor pagination on qpaths.
const searchPathLimit = 10_000

// pathsTTL: the omnibar asks `$search&json` on every debounced keystroke, and
// each answer needs the listing. singleflight only merges calls in flight at
// the same moment, so without reuse a reader typing `author:de` costs two
// 10000-path qpaths per keystroke. /search.json shares the same cache.
const pathsTTL = 5 * time.Second

// rpcRealmDirectory serves paths from the chain, reusing a listing for
// pathsTTL. The semaphore bounds concurrent outbound RPC queries; the group
// coalesces concurrent misses so a cold cache cannot amplify a burst of
// clients into a burst of RPC calls.
type rpcRealmDirectory struct {
	client pathLister
	domain string
	sem    chan struct{}
	sf     singleflight.Group

	mu       sync.Mutex
	cached   PathsResult
	cachedAt time.Time
}

func newRPCRealmDirectory(client pathLister, domain string, maxConcurrent int) *rpcRealmDirectory {
	return &rpcRealmDirectory{
		client: client,
		domain: domain,
		sem:    make(chan struct{}, maxConcurrent),
	}
}

// fetchTimeout bounds the shared fetch. /search.json is mounted straight on
// the mux, so it carries no handler timeout of its own; without this, a
// follower with a three-second budget could hang on a leader bounded only by
// the RPC client's one-minute ceiling.
const fetchTimeout = 10 * time.Second

// Paths fans out one query per kind (r, p). Concurrent callers share a single
// in-flight fetch via singleflight.
//
// The leader's fetch is detached from its own request and each caller waits
// on its own context: singleflight carries no context, so a leader that
// disconnects would otherwise cancel the answer every follower was waiting
// on — one reader closing a tab failing everyone else's search.
func (d *rpcRealmDirectory) Paths(ctx context.Context) (PathsResult, error) {
	if res, ok := d.fresh(); ok {
		return res, nil
	}

	ch := d.sf.DoChan("paths", func() (any, error) {
		// Re-check under the group: a caller that just finished a fetch has
		// stored it by the time a late one gets here.
		if res, ok := d.fresh(); ok {
			return res, nil
		}
		fetchCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), fetchTimeout)
		defer cancel()
		res, err := d.fetchPaths(fetchCtx)
		if err != nil {
			return PathsResult{}, err
		}
		d.mu.Lock()
		d.cached, d.cachedAt = res, time.Now()
		d.mu.Unlock()
		return res, nil
	})

	select {
	case res := <-ch:
		if res.Err != nil {
			return PathsResult{}, res.Err
		}
		return res.Val.(PathsResult), nil
	case <-ctx.Done():
		return PathsResult{}, ctx.Err()
	}
}

// fresh returns the cached listing while it is within pathsTTL. Callers
// only read the slices, so sharing them is safe.
func (d *rpcRealmDirectory) fresh() (PathsResult, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.cachedAt.IsZero() && time.Since(d.cachedAt) < pathsTTL {
		return d.cached, true
	}
	return PathsResult{}, false
}

func (d *rpcRealmDirectory) fetchPaths(ctx context.Context) (PathsResult, error) {
	var (
		wg         sync.WaitGroup
		res        PathsResult
		rErr, pErr error
	)
	wg.Add(2)
	go func() { defer wg.Done(); res.Realms, rErr = d.list(ctx, path.Join(d.domain, "r")) }()
	go func() { defer wg.Done(); res.Packages, pErr = d.list(ctx, path.Join(d.domain, "p")) }()
	wg.Wait()
	if rErr != nil {
		return PathsResult{}, rErr
	}
	if pErr != nil {
		return PathsResult{}, pErr
	}
	// A listing that comes back at the cap may have more behind it. There is
	// no cursor to ask with, so "at least this many" is the honest reading —
	// and over-reporting truncation on an exactly-full chain is the harmless
	// direction to be wrong in.
	res.Truncated = len(res.Realms) >= searchPathLimit || len(res.Packages) >= searchPathLimit
	return res, nil
}

// list fetches paths under prefix, bounded by the semaphore, dropping the empty
// entries the RPC layer yields for an empty result.
func (d *rpcRealmDirectory) list(ctx context.Context, prefix string) ([]string, error) {
	select {
	case d.sem <- struct{}{}:
		defer func() { <-d.sem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	paths, err := d.client.ListPaths(ctx, prefix, searchPathLimit)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if p != "" {
			out = append(out, p)
		}
	}
	return out, nil
}

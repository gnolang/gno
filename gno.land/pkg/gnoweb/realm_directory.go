package gnoweb

import (
	"context"
	"path"
	"sync"

	"golang.org/x/sync/singleflight"
)

// pathLister is the subset of ClientAdapter the directory depends on.
type pathLister interface {
	ListPaths(ctx context.Context, prefix string, limit int) ([]string, error)
}

// RealmDirectory exposes realm and package paths for discovery. It is the seam
// behind which the source can evolve (live RPC today, a dedicated search index
// later) without touching callers.
type RealmDirectory interface {
	// Paths returns the realm (/r/) and package (/p/) paths known to the chain.
	Paths(ctx context.Context) (realms, packages []string, err error)
}

var _ RealmDirectory = (*rpcRealmDirectory)(nil)

// searchPathLimit is the requested per-prefix page size. Note: the current RPC
// client does not forward it to the node, so the node's default cap (1000)
// effectively governs; raising the real cap needs limit forwarding + qpaths
// cursor pagination beyond 10000.
const searchPathLimit = 1000

// rpcRealmDirectory serves paths straight from the chain. It holds no state
// beyond a semaphore and a singleflight group: the semaphore bounds concurrent
// outbound RPC queries; the group coalesces concurrent /search.json hits so a
// cold edge cache cannot amplify a burst of clients into a burst of RPC calls.
type rpcRealmDirectory struct {
	client pathLister
	domain string
	sem    chan struct{}
	sf     singleflight.Group
}

func newRPCRealmDirectory(client pathLister, domain string, maxConcurrent int) *rpcRealmDirectory {
	return &rpcRealmDirectory{
		client: client,
		domain: domain,
		sem:    make(chan struct{}, maxConcurrent),
	}
}

type pathsResult struct {
	realms   []string
	packages []string
}

// Paths fans out one query per kind (r, p). Concurrent callers share a single
// in-flight fetch via singleflight; the leader's context governs cancellation,
// which is acceptable here as the result is short-lived and identical for all.
func (d *rpcRealmDirectory) Paths(ctx context.Context) (realms, packages []string, err error) {
	v, err, _ := d.sf.Do("paths", func() (any, error) {
		return d.fetchPaths(ctx)
	})
	if err != nil {
		return nil, nil, err
	}
	res := v.(pathsResult)
	return res.realms, res.packages, nil
}

func (d *rpcRealmDirectory) fetchPaths(ctx context.Context) (pathsResult, error) {
	var (
		wg         sync.WaitGroup
		res        pathsResult
		rErr, pErr error
	)
	wg.Add(2)
	go func() { defer wg.Done(); res.realms, rErr = d.list(ctx, path.Join(d.domain, "r")) }()
	go func() { defer wg.Done(); res.packages, pErr = d.list(ctx, path.Join(d.domain, "p")) }()
	wg.Wait()
	if rErr != nil {
		return pathsResult{}, rErr
	}
	if pErr != nil {
		return pathsResult{}, pErr
	}
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

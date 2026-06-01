package gnoweb

import (
	"context"
	"path"
	"sync"
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

// rpcRealmDirectory serves paths straight from the chain. It holds no state:
// each call fetches the current paths. A semaphore bounds concurrent outbound
// queries so a burst of callers cannot amplify load against the node.
type rpcRealmDirectory struct {
	client pathLister
	domain string
	sem    chan struct{}
}

func newRPCRealmDirectory(client pathLister, domain string, maxConcurrent int) *rpcRealmDirectory {
	return &rpcRealmDirectory{
		client: client,
		domain: domain,
		sem:    make(chan struct{}, maxConcurrent),
	}
}

func (d *rpcRealmDirectory) Paths(ctx context.Context) (realms, packages []string, err error) {
	var wg sync.WaitGroup
	var rErr, pErr error
	wg.Add(2)
	go func() { defer wg.Done(); realms, rErr = d.list(ctx, path.Join(d.domain, "r")) }()
	go func() { defer wg.Done(); packages, pErr = d.list(ctx, path.Join(d.domain, "p")) }()
	wg.Wait()
	if rErr != nil {
		return nil, nil, rErr
	}
	if pErr != nil {
		return nil, nil, pErr
	}
	return realms, packages, nil
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

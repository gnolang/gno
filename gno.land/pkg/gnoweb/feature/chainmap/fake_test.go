package chainmap

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

// fakeIndexer answers from fixed data. Blocks are one second apart from t0.
type fakeIndexer struct {
	tip int
	t0  time.Time

	// calls are MsgCall transactions by height.
	calls map[int][]indexer.Tx
	// capOver makes CallsBetween refuse any band wider than this many
	// blocks, as the indexer's element cap does; 0 never refuses.
	capOver int
	// failBands makes CallsBetween fail outright.
	failBands bool
	// bandDelay makes each CallsBetween take this long, or until its ctx ends.
	bandDelay time.Duration
	// block makes DeploysQuoting wait until it is closed.
	block chan struct{}

	recent       []indexer.Tx
	recentCalls  int
	deploys      []indexer.Tx
	deploysErr   error
	quoted       []string
	mu           sync.Mutex
	bandsQueried [][2]int
}

func (f *fakeIndexer) LatestBlockHeight(context.Context) (int, error) { return f.tip, nil }

func (f *fakeIndexer) Block(_ context.Context, height int) (*indexer.Block, error) {
	if height < 1 || height > f.tip {
		return nil, fmt.Errorf("block %w: %d", indexer.ErrNotFound, height)
	}
	return &indexer.Block{Height: height, Time: f.t0.Add(time.Duration(height) * time.Second)}, nil
}

func (f *fakeIndexer) CallsBetween(ctx context.Context, lower, upper int) ([]indexer.Tx, error) {
	f.mu.Lock()
	f.bandsQueried = append(f.bandsQueried, [2]int{lower, upper})
	f.mu.Unlock()
	if f.bandDelay > 0 {
		select {
		case <-time.After(f.bandDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if f.failBands {
		return nil, errors.New("indexer down")
	}
	if f.capOver > 0 && upper-lower > f.capOver {
		return nil, fmt.Errorf("%w: max elements per query", indexer.ErrTooLarge)
	}
	var out []indexer.Tx
	for h := lower + 1; h <= upper; h++ {
		out = append(out, f.calls[h]...)
	}
	return out, nil
}

// DeploysQuoting answers the deploys whose height falls in the band; a
// negative lower bound reaches height 0, as on the indexer.
func (f *fakeIndexer) DeploysQuoting(_ context.Context, pkgPath string, lower, upper int) ([]indexer.Tx, error) {
	f.mu.Lock()
	f.quoted = append(f.quoted, pkgPath)
	f.mu.Unlock()
	if f.block != nil {
		<-f.block
	}
	var out []indexer.Tx
	for _, tx := range f.deploys {
		if tx.Height > lower && tx.Height <= upper {
			out = append(out, tx)
		}
	}
	return out, f.deploysErr
}

func (f *fakeIndexer) URL() string { return "https://indexer.test/graphql/query" }

// call builds a MsgCall transaction.
func call(height int, ok bool, caller, pkg string) indexer.Tx {
	tx := indexer.Tx{Height: height, Success: ok, Messages: make([]indexer.Message, 1)}
	tx.Messages[0].Value.Type = "MsgCall"
	tx.Messages[0].Value.Caller = caller
	tx.Messages[0].Value.PkgPath = pkg
	return tx
}

// deploy builds a MsgAddPackage transaction adding every path given, as a
// batched deploy does.
func deploy(paths ...string) indexer.Tx {
	tx := indexer.Tx{Messages: make([]indexer.Message, len(paths))}
	for i, p := range paths {
		tx.Messages[i].Value.Type = "MsgAddPackage"
		tx.Messages[i].Value.Package = &struct {
			Path string `json:"path"`
		}{Path: p}
	}
	return tx
}

// fakeImports answers a package's imports from a map; a path absent from it
// is not live.
type fakeImports struct {
	imports map[string][]string
	fail    map[string]bool
}

func (f fakeImports) Imports(_ context.Context, pkgPath string) ([]string, error) {
	if f.fail[pkgPath] {
		return nil, errors.New("node unreachable")
	}
	imps, ok := f.imports[pkgPath]
	if !ok {
		return nil, ErrNotLive
	}
	return imps, nil
}

func (f *fakeIndexer) RecentByPackage(_ context.Context, _ string, _ int) ([]indexer.Tx, error) {
	f.mu.Lock()
	f.recentCalls++
	f.mu.Unlock()
	return f.recent, nil
}

func (f *fakeIndexer) BlockTimes(_ context.Context, heights []int) (map[int]time.Time, error) {
	out := make(map[int]time.Time, len(heights))
	for _, h := range heights {
		out[h] = f.t0.Add(time.Duration(h) * time.Second)
	}
	return out, nil
}

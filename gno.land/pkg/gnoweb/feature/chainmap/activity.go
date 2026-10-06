package chainmap

import (
	"context"
	"errors"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

// Activity windows and budgets.
const (
	// activityWindow is how far back calls are counted.
	activityWindow = 7 * 24 * time.Hour

	// activityTTL is how long a computed aggregate is served before a request
	// refreshes it. The map is a picture of a week; a few minutes of lag is
	// not a different picture.
	activityTTL = 5 * time.Minute

	// activityTimeout bounds one refresh, independent of the request that
	// started it.
	activityTimeout = 30 * time.Second

	// rateProbe is how many blocks back the block rate is measured over.
	rateProbe = 20_000

	// bandWidth is the first height span read per query. On gnoland-1 a
	// span this wide held about 600 calls in October 2026, well under the
	// indexer's element cap; denser chains split bands as they need to.
	bandWidth = 25_000

	// minBandWidth is the narrowest band split towards. A band this narrow
	// still over the cap is left uncounted and the aggregate marked partial.
	minBandWidth = 250

	// bandConcurrency bounds the queries one refresh keeps in flight.
	bandConcurrency = 3
)

// Activity is the call count of every realm over the window.
type Activity struct {
	// Calls and Callers are keyed by fully qualified package path.
	Calls   map[string]int
	Callers map[string]int

	// From and To are the heights counted, both inclusive.
	From, To int
	// Since is when block From was produced.
	Since time.Time

	// Partial is set when part of the window could not be read. A count is
	// then a lower bound, and a zero says nothing.
	Partial bool
}

// activityCache holds the latest aggregate and refreshes it on demand. It runs
// nothing in the background: a refresh is started by a request, detached from
// it so a closed tab does not waste the work, and shared by every request that
// arrives while it runs.
type activityCache struct {
	idx Indexer

	mu  sync.Mutex
	cur *Activity
	at  time.Time

	group singleflight.Group
}

// get returns the aggregate, waiting for a refresh only as long as ctx allows.
// A stale aggregate is returned at once while a fresh one is computed behind
// it. With nothing to serve and no time left, it returns ctx's error, and the
// refresh it started keeps running for the next reader.
func (c *activityCache) get(ctx context.Context) (*Activity, error) {
	c.mu.Lock()
	cur, fresh := c.cur, time.Since(c.at) < activityTTL
	c.mu.Unlock()
	if cur != nil && fresh {
		return cur, nil
	}

	ch := c.group.DoChan("activity", func() (any, error) {
		rctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), activityTimeout)
		defer cancel()
		a, err := computeActivity(rctx, c.idx)
		if err != nil {
			return nil, err
		}
		c.mu.Lock()
		c.cur, c.at = a, time.Now()
		c.mu.Unlock()
		return a, nil
	})
	if cur != nil {
		return cur, nil
	}

	select {
	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		return res.Val.(*Activity), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// computeActivity counts the window's calls band by band.
func computeActivity(ctx context.Context, idx Indexer) (*Activity, error) {
	tip, err := idx.LatestBlockHeight(ctx)
	if err != nil {
		return nil, err
	}
	from, since, err := windowStart(ctx, idx, tip)
	if err != nil {
		return nil, err
	}

	a := &Activity{
		Calls:   make(map[string]int),
		Callers: make(map[string]int),
		From:    from,
		To:      tip,
		Since:   since,
	}
	callers := make(map[string]map[string]struct{})

	var mu sync.Mutex
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(bandConcurrency)
	for lower := from - 1; lower < tip; lower += bandWidth {
		upper := min(lower+bandWidth, tip)
		g.Go(func() error {
			txs, complete, err := readBand(gctx, idx, lower, upper)
			if err != nil {
				return err
			}
			mu.Lock()
			defer mu.Unlock()
			a.Partial = a.Partial || !complete
			for _, tx := range txs {
				if !tx.Success {
					continue
				}
				for _, m := range tx.Messages {
					p := m.Path()
					if m.Type() != "MsgCall" || p == "" {
						continue
					}
					a.Calls[p]++
					if callers[p] == nil {
						callers[p] = make(map[string]struct{})
					}
					callers[p][m.Value.Caller] = struct{}{}
				}
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	for p, set := range callers {
		a.Callers[p] = len(set)
	}
	return a, nil
}

// readBand reads the heights (lower, upper], halving any band the indexer
// caps. complete is false when some band stayed over the cap at the narrowest
// width; its calls are then missing from the result.
func readBand(ctx context.Context, idx Indexer, lower, upper int) ([]indexer.Tx, bool, error) {
	txs, err := idx.CallsBetween(ctx, lower, upper)
	switch {
	case err == nil:
		return txs, true, nil
	case !errors.Is(err, indexer.ErrTooLarge):
		return nil, false, err
	case upper-lower <= minBandWidth:
		return nil, false, nil
	}
	mid := lower + (upper-lower)/2
	low, lowOK, err := readBand(ctx, idx, lower, mid)
	if err != nil {
		return nil, false, err
	}
	high, highOK, err := readBand(ctx, idx, mid, upper)
	if err != nil {
		return nil, false, err
	}
	return append(low, high...), lowOK && highOK, nil
}

// windowStart estimates the first height of the window from the block rate
// over the last rateProbe blocks, then reads that block's time so the map can
// say exactly where its window starts.
func windowStart(ctx context.Context, idx Indexer, tip int) (int, time.Time, error) {
	if tip <= rateProbe {
		b, err := idx.Block(ctx, 1)
		if err != nil {
			return 0, time.Time{}, err
		}
		return 1, b.Time, nil
	}
	head, err := idx.Block(ctx, tip)
	if err != nil {
		return 0, time.Time{}, err
	}
	probe, err := idx.Block(ctx, tip-rateProbe)
	if err != nil {
		return 0, time.Time{}, err
	}
	elapsed := head.Time.Sub(probe.Time)
	if elapsed <= 0 {
		return 0, time.Time{}, errors.New("indexer block times do not increase")
	}
	blocks := int(float64(rateProbe) * float64(activityWindow) / float64(elapsed))
	from := max(tip-blocks, 1)

	start, err := idx.Block(ctx, from)
	if err != nil {
		return 0, time.Time{}, err
	}
	return from, start.Time, nil
}

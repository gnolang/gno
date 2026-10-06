package chainmap

import (
	"context"
	"errors"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

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

	// failureTTL is how long a failed indexer answer is remembered, so an
	// indexer that is down is not asked again by every reader.
	failureTTL = time.Minute

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

// newActivityFlight holds the one aggregate. A stale one is served at once
// while a request refreshes it, so after the first map nobody waits on the
// indexer for activity.
func newActivityFlight() *flight[*Activity] {
	return &flight[*Activity]{ttl: activityTTL, errTTL: failureTTL, timeout: activityTimeout, max: 1, stale: true}
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
	// partial marks the window as not fully read. A refresh that runs out of
	// time keeps what it counted and says the rest is unknown, rather than
	// throwing a week of counts away for its last band.
	partial := func() {
		mu.Lock()
		a.Partial = true
		mu.Unlock()
	}
	for lower := from - 1; lower < tip; lower += bandWidth {
		// A band started after the refresh gave up would only fail, and each
		// failure counts against the indexer client's breaker.
		if gctx.Err() != nil {
			partial()
			break
		}
		upper := min(lower+bandWidth, tip)
		g.Go(func() error {
			txs, complete, err := readBand(gctx, idx, lower, upper)
			if err != nil {
				if ctx.Err() != nil {
					partial()
					return nil
				}
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
	// The halves are read one after the other: splitting them into goroutines
	// at every level would put up to 2^depth queries in flight per band, past
	// bandConcurrency and onto the indexer client shared with search.
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
// say exactly where its window starts. A chain younger than the probe is
// counted from its first block.
func windowStart(ctx context.Context, idx Indexer, tip int) (int, time.Time, error) {
	from := 1
	if tip > rateProbe {
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
		from = max(tip-blocks, 1)
	}
	start, err := idx.Block(ctx, from)
	if err != nil {
		return 0, time.Time{}, err
	}
	return from, start.Time, nil
}

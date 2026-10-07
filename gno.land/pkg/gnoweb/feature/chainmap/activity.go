package chainmap

import (
	"context"
	"errors"
	"maps"
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

	// bandWidth is the first height span read per query. Bands are aligned
	// on multiples of it, so one that has closed reads the same on every
	// refresh and is read once. On gnoland-1 the busiest week of October 2026
	// held about 2,800 calls in such a band, well under the indexer's element
	// cap; denser chains split bands as they need to.
	bandWidth = 25_000

	// minBandWidth is the narrowest band split towards. A band this narrow
	// still over the cap is left uncounted and the aggregate marked partial.
	minBandWidth = 250

	// bandConcurrency bounds the queries one refresh keeps in flight.
	bandConcurrency = 3
)

// Activity is the call count of every realm over the window.
type Activity struct {
	// Calls, Callers and Gas are keyed by fully qualified package path. A
	// transaction's gas is shared evenly between the calls it makes, the
	// indexer reporting gas per transaction only.
	Calls   map[string]int
	Callers map[string]int
	Gas     map[string]int64

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

// bandCounts are the successful calls of one band, per package.
type bandCounts struct {
	calls   map[string]int
	callers map[string]map[string]struct{}
	gas     map[string]int64
	// complete is false when part of the band stayed over the indexer's cap.
	complete bool
}

func countBand(txs []indexer.Tx, complete bool) *bandCounts {
	b := &bandCounts{calls: make(map[string]int), callers: make(map[string]map[string]struct{}), gas: make(map[string]int64), complete: complete}
	for _, tx := range txs {
		if !tx.Success {
			continue
		}
		// The gas is shared between the messages that run code: a call
		// batched with a run or a deploy gets its share, not the whole.
		var calls []indexer.Message
		workers := 0
		for _, m := range tx.Messages {
			switch m.Type() {
			case "MsgCall":
				workers++
				if m.Path() != "" {
					calls = append(calls, m)
				}
			case "MsgRun", "MsgAddPackage":
				workers++
			}
		}
		for _, m := range calls {
			p := m.Path()
			b.gas[p] += int64(tx.GasUsed) / int64(workers)
			b.calls[p]++
			if b.callers[p] == nil {
				b.callers[p] = make(map[string]struct{})
			}
			b.callers[p][m.Value.Caller] = struct{}{}
		}
	}
	return b
}

// closedBands keeps the counts of whole bands below the indexer's tip, keyed
// by their lower bound: their blocks are final, so a refresh reads only the
// band the window starts in, the one it ends in, and any band closed since.
// Bands that fall out of the window are dropped.
type closedBands struct {
	mu sync.Mutex
	m  map[int]*bandCounts
}

// get returns the counts of the band starting at lower; nil-safe.
func (c *closedBands) get(lower int) *bandCounts {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m[lower]
}

func (c *closedBands) put(lower int, b *bandCounts) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.m == nil {
		c.m = make(map[int]*bandCounts)
	}
	c.m[lower] = b
}

// prune drops the bands starting below first.
func (c *closedBands) prune(first int) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for lower := range c.m {
		if lower < first {
			delete(c.m, lower)
		}
	}
}

// computeActivity counts the window's calls band by band. Whole bands come
// from closed when it holds them, and go into it once read; closed may be
// nil.
func computeActivity(ctx context.Context, idx Indexer, closed *closedBands) (*Activity, error) {
	tip, err := idx.LatestBlockHeight(ctx)
	if err != nil {
		return nil, err
	}
	from, since, err := windowStart(ctx, idx, tip)
	if err != nil {
		return nil, err
	}

	// Heights (from-1, tip], cut on multiples of bandWidth.
	first := (from - 1) - (from-1)%bandWidth
	closed.prune(first)
	var bands []*bandCounts
	for aligned := first; aligned < tip; aligned += bandWidth {
		// The band the window starts in is cached whole, but counts only
		// from the window's start: read it clipped unless it starts there.
		if aligned < from-1 {
			bands = append(bands, nil)
			continue
		}
		bands = append(bands, closed.get(aligned))
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(bandConcurrency)
	for i := range bands {
		if bands[i] != nil {
			continue
		}
		aligned := first + i*bandWidth
		lower, upper := max(aligned, from-1), min(aligned+bandWidth, tip)
		whole := lower == aligned && upper == aligned+bandWidth
		g.Go(func() error {
			// Checked here, not before g.Go, which may block on the limit
			// past the moment the refresh gave up: a band started then would
			// only fail, against the indexer client's breaker. A refresh out
			// of time fails as a whole, and the last good week stays served.
			if err := gctx.Err(); err != nil {
				return err
			}
			txs, complete, err := readBand(gctx, idx, lower, upper)
			if err != nil {
				return err
			}
			// Each goroutine writes its own slot: no lock needed.
			bands[i] = countBand(txs, complete)
			// Only a band read in full is final: a partial one keeps its gap
			// until it leaves the window otherwise.
			if whole && bands[i].complete {
				closed.put(aligned, bands[i])
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	a := &Activity{
		Calls:   make(map[string]int),
		Callers: make(map[string]int),
		Gas:     make(map[string]int64),
		From:    from,
		To:      tip,
		Since:   since,
	}
	callers := make(map[string]map[string]struct{})
	for _, b := range bands {
		a.Partial = a.Partial || !b.complete
		for p, n := range b.calls {
			a.Calls[p] += n
		}
		for p, g := range b.gas {
			a.Gas[p] += g
		}
		for p, set := range b.callers {
			if callers[p] == nil {
				callers[p] = make(map[string]struct{}, len(set))
			}
			maps.Copy(callers[p], set)
		}
	}
	for p, set := range callers {
		a.Callers[p] = len(set)
	}
	return a, nil
}

// readBand reads the heights (lower, upper], halving any band over the
// indexer's element cap or the client's response size cap. complete is false
// when some band stayed over a cap at the narrowest width; its calls are then
// missing from the result.
func readBand(ctx context.Context, idx Indexer, lower, upper int) ([]indexer.Tx, bool, error) {
	txs, err := idx.CallsBetween(ctx, lower, upper)
	switch {
	case err == nil:
		return txs, true, nil
	case !errors.Is(err, indexer.ErrTooLarge) && !errors.Is(err, indexer.ErrResponseTooLarge):
		// Over the element cap or over the client's size cap, a narrower
		// band may fit; anything else is a failure.
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

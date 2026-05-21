package fork

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	coretypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
)

// blockData bundles a block and its execution results for a single height.
type blockData struct {
	block   *coretypes.ResultBlock
	results *coretypes.ResultBlockResults
}

// blockResult is one item delivered by pooledFetcher.FetchRange. err is set
// when the fetcher gave up on this height after exhausting retries; err
// terminates the stream.
type blockResult struct {
	height int64
	data   *blockData
	err    error
}

// pooledFetcher fetches blocks across one or more RPC endpoints in parallel.
//
// It owns no client itself: fetch is the per-endpoint call, indexed by
// endpoint id in [0, numEndpoints). Workers (numEndpoints * workersPerEndpoint
// total) pull heights from a shared queue, round-robin through endpoints
// starting at their assigned offset, and rotate to the next endpoint on
// failure. A per-endpoint semaphore caps in-flight calls at workersPerEndpoint.
// After all endpoints fail for a given height in one cycle, it sleeps backoff
// and retries up to maxCycles times before yielding a terminal error.
//
// Memory ceiling: the reorder buffer holds at most ~3*numEndpoints*
// workersPerEndpoint blocks in-flight (heights channel + worker results
// channel + reorder map). Operators sizing endpoints/workers should bound
// peak memory by (3 * numEndpoints * workersPerEndpoint) * avg-block-size.
type pooledFetcher struct {
	numEndpoints       int
	workersPerEndpoint int
	maxCycles          int
	backoff            time.Duration
	fetch              func(ctx context.Context, endpoint int, h int64) (*blockData, error)
}

// FetchRange returns a channel emitting [from, to] in monotonic height order.
// The channel closes once all heights have been delivered, on context
// cancellation, or after a terminal error.
func (p *pooledFetcher) FetchRange(ctx context.Context, from, to int64) <-chan blockResult {
	if from > to {
		out := make(chan blockResult)
		close(out)
		return out
	}
	if p.numEndpoints <= 0 || p.workersPerEndpoint <= 0 {
		out := make(chan blockResult, 1)
		out <- blockResult{
			height: from,
			err: fmt.Errorf("pooledFetcher: numEndpoints=%d workersPerEndpoint=%d, both must be >= 1",
				p.numEndpoints, p.workersPerEndpoint),
		}
		close(out)
		return out
	}
	out := make(chan blockResult)
	go p.run(ctx, from, to, out)
	return out
}

func (p *pooledFetcher) run(ctx context.Context, from, to int64, out chan<- blockResult) {
	defer close(out)

	// Own a child context so any early return (terminal error, ctx.Done in
	// the reorder loop) propagates to producer + workers + fetchWithFailover.
	// Without this, a terminal-error path closes `out` but leaves background
	// goroutines fetching and trying to push to workerResults until the
	// caller's parent ctx eventually cancels.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	totalWorkers := p.numEndpoints * p.workersPerEndpoint
	heights := make(chan int64, totalWorkers)
	workerResults := make(chan blockResult, totalWorkers)
	sems := make([]chan struct{}, p.numEndpoints)
	for i := range sems {
		sems[i] = make(chan struct{}, p.workersPerEndpoint)
	}

	// Heights producer.
	go func() {
		defer close(heights)
		for h := from; h <= to; h++ {
			select {
			case <-ctx.Done():
				return
			case heights <- h:
			}
		}
	}()

	// Worker pool.
	var wg sync.WaitGroup
	for w := 0; w < totalWorkers; w++ {
		wg.Go(func() {
			startEp := w % p.numEndpoints
			for h := range heights {
				data, err := p.fetchWithFailover(ctx, h, startEp, sems)
				select {
				case <-ctx.Done():
					return
				case workerResults <- blockResult{height: h, data: data, err: err}:
				}
			}
		})
	}

	go func() {
		wg.Wait()
		close(workerResults)
	}()

	// Reorder + emit in monotonic order.
	buffer := make(map[int64]blockResult, totalWorkers)
	next := from
	for r := range workerResults {
		if r.err != nil {
			select {
			case <-ctx.Done():
			case out <- r:
			}
			return
		}
		buffer[r.height] = r
		for {
			br, ok := buffer[next]
			if !ok {
				break
			}
			delete(buffer, next)
			select {
			case <-ctx.Done():
				return
			case out <- br:
			}
			next++
		}
	}
}

// fetchWithFailover tries each endpoint in round-robin order starting at
// startEp. On per-endpoint failure it rotates to the next endpoint; once all
// endpoints have been tried in a cycle, it backs off and retries the cycle up
// to maxCycles times. Reuses a single timer across cycles to avoid per-retry
// allocations.
func (p *pooledFetcher) fetchWithFailover(
	ctx context.Context, h int64, startEp int, sems []chan struct{},
) (*blockData, error) {
	var (
		lastErr      error
		backoffTimer *time.Timer
	)
	defer func() {
		if backoffTimer != nil {
			backoffTimer.Stop()
		}
	}()

	for cycle := 0; cycle < p.maxCycles; cycle++ {
		for k := 0; k < p.numEndpoints; k++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			ep := (startEp + k) % p.numEndpoints
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case sems[ep] <- struct{}{}:
			}

			data, err := p.fetch(ctx, ep, h)
			<-sems[ep]

			if err == nil {
				return data, nil
			}
			lastErr = err
		}

		if cycle < p.maxCycles-1 {
			if backoffTimer == nil {
				backoffTimer = time.NewTimer(p.backoff)
			} else {
				backoffTimer.Reset(p.backoff)
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-backoffTimer.C:
			}
		}
	}
	return nil, fmt.Errorf("height %d: all %d endpoint(s) failed after %d cycles: %w",
		h, p.numEndpoints, p.maxCycles, lastErr)
}

// errNoEndpoints is returned by tryEndpoints when given an empty client slice.
var errNoEndpoints = errors.New("no RPC clients configured")

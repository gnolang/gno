package fork

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	coretypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func okBlock(h int64) *blockData {
	return &blockData{
		block:   &coretypes.ResultBlock{},
		results: &coretypes.ResultBlockResults{Height: h},
	}
}

func TestPooledFetcher_InOrder_SingleEndpoint(t *testing.T) {
	t.Parallel()

	f := &pooledFetcher{
		numEndpoints:       1,
		workersPerEndpoint: 4,
		maxCycles:          3,
		backoff:            time.Millisecond,
		fetch: func(_ context.Context, _ int, h int64) (*blockData, error) {
			return &blockData{
				block:   &coretypes.ResultBlock{},
				results: &coretypes.ResultBlockResults{Height: h},
			}, nil
		},
	}

	out := f.FetchRange(context.Background(), 1, 100)
	var heights []int64
	for r := range out {
		require.NoError(t, r.err)
		heights = append(heights, r.height)
	}

	require.Len(t, heights, 100)
	for i, h := range heights {
		assert.Equal(t, int64(i+1), h)
	}
}

func TestPooledFetcher_InOrder_MultiEndpointWithJitter(t *testing.T) {
	t.Parallel()

	f := &pooledFetcher{
		numEndpoints:       3,
		workersPerEndpoint: 4,
		maxCycles:          3,
		backoff:            time.Millisecond,
		fetch: func(_ context.Context, _ int, h int64) (*blockData, error) {
			// Reverse-proportional delay: lower heights take longer so
			// later heights are likely to finish first if no reorder buffer.
			delay := time.Duration(200-h%200) * time.Microsecond
			time.Sleep(delay)
			return &blockData{
				block:   &coretypes.ResultBlock{},
				results: &coretypes.ResultBlockResults{Height: h},
			}, nil
		},
	}

	out := f.FetchRange(context.Background(), 1, 200)
	var heights []int64
	for r := range out {
		require.NoError(t, r.err)
		heights = append(heights, r.height)
	}

	require.Len(t, heights, 200)
	for i, h := range heights {
		assert.Equal(t, int64(i+1), h, "out of order at index %d", i)
	}
}

func TestPooledFetcher_FailoverOnEndpointError(t *testing.T) {
	t.Parallel()

	var ep0, ep1 atomic.Int64
	f := &pooledFetcher{
		numEndpoints:       2,
		workersPerEndpoint: 4,
		maxCycles:          3,
		backoff:            time.Millisecond,
		fetch: func(_ context.Context, endpoint int, h int64) (*blockData, error) {
			if endpoint == 0 {
				ep0.Add(1)
				return nil, errors.New("endpoint 0 down")
			}
			ep1.Add(1)
			return okBlock(h), nil
		},
	}

	out := f.FetchRange(context.Background(), 1, 50)
	var heights []int64
	for r := range out {
		require.NoError(t, r.err)
		heights = append(heights, r.height)
	}

	require.Len(t, heights, 50)
	for i, h := range heights {
		assert.Equal(t, int64(i+1), h)
	}
	assert.Equal(t, int64(50), ep1.Load(), "endpoint 1 should have served all 50 heights")
	assert.Greater(t, ep0.Load(), int64(0), "endpoint 0 should have been tried first by half the workers")
}

func TestPooledFetcher_AllEndpointsFailHardError(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64
	f := &pooledFetcher{
		numEndpoints:       2,
		workersPerEndpoint: 2,
		maxCycles:          3,
		backoff:            time.Millisecond,
		fetch: func(_ context.Context, _ int, _ int64) (*blockData, error) {
			calls.Add(1)
			return nil, errors.New("always fail")
		},
	}

	out := f.FetchRange(context.Background(), 1, 10)
	var heights []int64
	var terminalErr error
	for r := range out {
		if r.err != nil {
			terminalErr = r.err
			break
		}
		heights = append(heights, r.height)
	}

	require.Error(t, terminalErr, "expected a terminal error result")
	assert.Contains(t, terminalErr.Error(), "all 2 endpoint(s) failed after 3 cycles")
	// Channel must be drained until close.
	for range out {
	}
}

func TestPooledFetcher_RecoversAfterTransientFailure(t *testing.T) {
	t.Parallel()

	// Per-height retry: track failures keyed by height so each height fails
	// once then succeeds. This exercises the cycle-retry-after-backoff path
	// for the same endpoint (numEndpoints=1).
	var (
		mu     sync.Mutex
		failed = map[int64]int{}
	)
	f := &pooledFetcher{
		numEndpoints:       1,
		workersPerEndpoint: 2,
		maxCycles:          3,
		backoff:            time.Millisecond,
		fetch: func(_ context.Context, _ int, h int64) (*blockData, error) {
			mu.Lock()
			n := failed[h]
			failed[h] = n + 1
			mu.Unlock()
			if n == 0 {
				return nil, errors.New("transient")
			}
			return okBlock(h), nil
		},
	}

	out := f.FetchRange(context.Background(), 1, 20)
	var heights []int64
	for r := range out {
		require.NoError(t, r.err)
		heights = append(heights, r.height)
	}

	require.Len(t, heights, 20)
	for i, h := range heights {
		assert.Equal(t, int64(i+1), h)
	}
}

func TestPooledFetcher_SemaphoreCapsConcurrencyUnderFailover(t *testing.T) {
	t.Parallel()

	const workersPerEp = 2
	// Total workers = 4. Endpoint 0 always fails, so all 4 funnel onto
	// endpoint 1. Semaphore must cap in-flight on endpoint 1 at 2.
	var (
		ep1Cur, ep1Max atomic.Int64
	)
	f := &pooledFetcher{
		numEndpoints:       2,
		workersPerEndpoint: workersPerEp,
		maxCycles:          3,
		backoff:            time.Millisecond,
		fetch: func(_ context.Context, endpoint int, h int64) (*blockData, error) {
			if endpoint == 0 {
				return nil, errors.New("endpoint 0 down")
			}
			cur := ep1Cur.Add(1)
			for {
				m := ep1Max.Load()
				if cur <= m || ep1Max.CompareAndSwap(m, cur) {
					break
				}
			}
			// Hold long enough that other workers queue on the semaphore.
			time.Sleep(2 * time.Millisecond)
			ep1Cur.Add(-1)
			return okBlock(h), nil
		},
	}

	out := f.FetchRange(context.Background(), 1, 30)
	for r := range out {
		require.NoError(t, r.err)
	}

	maxConcurrent := ep1Max.Load()
	assert.LessOrEqual(t, maxConcurrent, int64(workersPerEp),
		"endpoint 1 in-flight exceeded semaphore: max=%d, cap=%d", maxConcurrent, workersPerEp)
	assert.Equal(t, int64(workersPerEp), maxConcurrent,
		"expected to saturate semaphore; max=%d", maxConcurrent)
}

func TestPooledFetcher_ContextCancellationClosesChannel(t *testing.T) {
	t.Parallel()

	f := &pooledFetcher{
		numEndpoints:       1,
		workersPerEndpoint: 2,
		maxCycles:          3,
		backoff:            time.Millisecond,
		fetch: func(ctx context.Context, _ int, h int64) (*blockData, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(10 * time.Millisecond):
				return okBlock(h), nil
			}
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	out := f.FetchRange(ctx, 1, 100000)

	// Read a few results then cancel.
	received := 0
	for r := range out {
		if r.err != nil {
			break
		}
		received++
		if received == 3 {
			cancel()
			break
		}
	}

	// Channel must close within a reasonable bound after cancel; if any
	// goroutine forgets to honour ctx.Done() it would leak and this drain
	// would hang until the test timeout.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range out {
		}
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("FetchRange channel did not close within 2s after context cancellation")
	}
}

func TestPooledFetcher_EmptyRange(t *testing.T) {
	t.Parallel()

	f := &pooledFetcher{
		numEndpoints:       1,
		workersPerEndpoint: 4,
		maxCycles:          3,
		backoff:            time.Millisecond,
		fetch: func(_ context.Context, _ int, _ int64) (*blockData, error) {
			t.Fatal("fetch should not be called for empty range")
			return nil, nil
		},
	}

	out := f.FetchRange(context.Background(), 10, 9)
	for range out {
		t.Fatal("expected no results")
	}
}

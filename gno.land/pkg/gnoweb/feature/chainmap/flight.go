package chainmap

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// ErrPending is returned when a reader's own wait ran out before the answer
// came. The fetch carries on; a reader can say "still working" for this
// error, and only for this one, since a fetch that failed on a deadline is a
// failure.
var ErrPending = errors.New("answer still pending")

// errTransient marks a fetch error that says nothing about the answer, such
// as no capacity or no budget to start the fetch: it is returned, to the
// reader who started the fetch and to any that joined it, but never
// remembered.
var errTransient = errors.New("transient")

// errFetchTimeout is why a fetch that ran past its flight's timeout ended.
var errFetchTimeout = errors.New("fetch timed out")

// flight memoizes one answer per key and refreshes it on request. It runs
// nothing in the background: a fetch is started by a reader, detached from
// that reader so a closed tab does not waste the work, and shared by every
// reader asking for the same key while it runs.
type flight[T any] struct {
	// ttl is how long an answer is served as fresh.
	ttl time.Duration
	// errTTL is how long a failure is remembered, so a failing backend is not
	// asked again by every reader. Zero remembers none.
	errTTL time.Duration
	// timeout bounds one fetch, whoever started it. It ends the fetch by
	// cancelling it, not with a deadline: the indexer client's breaker counts
	// a deadline as the indexer failing, and an open breaker turns search off
	// for every reader. A fetch cut short this way is still a failure here.
	timeout time.Duration
	// max bounds the number of keys; the oldest answer goes first.
	max int
	// stale serves an expired answer at once while it is refreshed.
	stale bool
	// partial, when set, marks answers kept only for errTTL, like failures:
	// an answer some reads failed to complete is worth retrying soon.
	partial func(T) bool

	mu      sync.Mutex
	entries map[string]flightEntry[T]
	// stored numbers the answers in the order they were stored, which is the
	// eviction order: clocks too coarse to tell two stores apart would leave
	// "oldest" to map order.
	stored uint64
	group  singleflight.Group
}

type flightEntry[T any] struct {
	val T
	err error
	at  time.Time
	seq uint64
}

// fresh returns the answer for key if it is still fresh.
func (f *flight[T]) fresh(key string) (flightEntry[T], bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	e, ok := f.entries[key]
	if !ok {
		return e, false
	}
	ttl := f.ttl
	if e.err != nil || (f.partial != nil && f.partial(e.val)) {
		ttl = f.errTTL
	}
	return e, time.Since(e.at) < ttl
}

// get answers key, fetching it when no fresh answer is held, and waits for
// the fetch only as long as ctx allows; the fetch carries on for the next
// reader.
func (f *flight[T]) get(ctx context.Context, key string, fetch func(context.Context) (T, error)) (T, error) {
	if e, ok := f.fresh(key); ok {
		return e.val, e.err
	}
	// A reader already out of time starts nothing: a loop over keys under an
	// expired deadline would otherwise launch every fetch at once.
	if err := ctx.Err(); err != nil {
		var zero T
		return zero, fmt.Errorf("%w: %w", ErrPending, err)
	}

	ch := f.group.DoChan(key, func() (any, error) {
		// A reader arriving as the previous fetch lands must not start
		// another one.
		if e, ok := f.fresh(key); ok {
			return e, nil
		}
		fctx, cancel := context.WithCancelCause(context.WithoutCancel(ctx))
		defer cancel(nil)
		stop := time.AfterFunc(f.timeout, func() { cancel(errFetchTimeout) })
		defer stop.Stop()
		val, err := fetch(fctx)
		if err != nil && context.Cause(fctx) == errFetchTimeout {
			err = fmt.Errorf("%w: %w", errFetchTimeout, err)
		}
		return f.store(key, val, err), nil
	})

	if f.stale {
		f.mu.Lock()
		e, ok := f.entries[key]
		f.mu.Unlock()
		if ok && e.err == nil {
			return e.val, nil
		}
	}

	select {
	case res := <-ch:
		e := res.Val.(flightEntry[T])
		return e.val, e.err
	case <-ctx.Done():
		var zero T
		return zero, fmt.Errorf("%w: %w", ErrPending, ctx.Err())
	}
}

// store records an answer, evicting the oldest when full, and returns what
// readers get. A failure is remembered for errTTL; a stale flight keeps
// serving its last good answer instead, and asks again after errTTL.
func (f *flight[T]) store(key string, val T, err error) flightEntry[T] {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stored++
	e := flightEntry[T]{val: val, err: err, at: time.Now(), seq: f.stored}
	if err != nil {
		if f.errTTL == 0 || errors.Is(err, errTransient) {
			return e
		}
		if old, ok := f.entries[key]; ok && f.stale && old.err == nil {
			old.at = time.Now().Add(f.errTTL - f.ttl)
			f.entries[key] = old
			return old
		}
	}
	if f.entries == nil {
		f.entries = make(map[string]flightEntry[T])
	}
	if _, ok := f.entries[key]; !ok && len(f.entries) >= f.max {
		var oldest string
		for k, old := range f.entries {
			if oldest == "" || old.seq < f.entries[oldest].seq {
				oldest = k
			}
		}
		delete(f.entries, oldest)
	}
	f.entries[key] = e
	return e
}

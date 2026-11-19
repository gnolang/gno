// Package flight memoizes answers per key for gnoweb's features: each key
// is fetched once for every reader asking at the same time, and its answer,
// or its failure, is kept for a while.
package flight

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// ErrPending is returned when a reader's own wait ran out before the answer
// came. The fetch carries on; a reader can say "still working" for this
// error, and only for this one, since a fetch that failed on a deadline is a
// failure.
var ErrPending = errors.New("answer still pending")

// ErrTransient marks a fetch error that says nothing about the answer, such
// as no capacity or no budget to start the fetch: it is returned, to the
// reader who started the fetch and to any that joined it, but never
// remembered.
var ErrTransient = errors.New("transient")

// errFetchTimeout is why a fetch that ran past its flight's timeout ended.
var errFetchTimeout = errors.New("fetch timed out")

// Flight memoizes one answer per key and refreshes it on request. It runs
// nothing in the background: a fetch is started by a reader, detached from
// that reader so a closed tab does not waste the work, and shared by every
// reader asking for the same key while it runs. A fetch that panics on its
// own goroutine fails its key instead of the process, and is logged with
// its stack; a panic on a goroutine the fetch starts is the fetch's to
// recover.
type Flight[T any] struct {
	// TTL is how long an answer is served as fresh.
	TTL time.Duration
	// ErrTTL is how long a failure is remembered, so a failing backend is not
	// asked again by every reader. Zero remembers none.
	ErrTTL time.Duration
	// Timeout bounds one fetch, whoever started it. A fetch cut short is a
	// failure here.
	Timeout time.Duration
	// Max bounds the number of keys; the oldest answer goes first.
	Max int
	// Stale serves an expired answer at once while it is refreshed.
	Stale bool
	// Partial, when set, marks answers kept only for ErrTTL, like failures:
	// an answer some reads failed to complete is worth retrying soon.
	Partial func(T) bool
	// Logger receives a fetch's panic and stack; nil uses slog.Default().
	Logger *slog.Logger

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
func (f *Flight[T]) fresh(key string) (flightEntry[T], bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	e, ok := f.entries[key]
	if !ok {
		return e, false
	}
	ttl := f.TTL
	if e.err != nil || (f.Partial != nil && f.Partial(e.val)) {
		ttl = f.ErrTTL
	}
	return e, time.Since(e.at) < ttl
}

// Get answers key, fetching it when no fresh answer is held, and waits for
// the fetch only as long as ctx allows; the fetch carries on for the next
// reader.
func (f *Flight[T]) Get(ctx context.Context, key string, fetch func(context.Context) (T, error)) (T, error) {
	if e, ok := f.fresh(key); ok {
		return e.val, e.err
	}
	// A reader already out of time starts nothing: a loop over keys under an
	// expired deadline would otherwise launch every fetch at once.
	if err := ctx.Err(); err != nil {
		var zero T
		return zero, fmt.Errorf("%w: %w", ErrPending, err)
	}

	ch := f.start(ctx, key, fetch)

	if f.Stale {
		if e, ok := f.held(key); ok {
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

// Peek returns the answer at hand, fresh or stale, and never waits: with
// none fresh it starts the fetch, or joins the one running, for the readers
// that follow. It is for a page drawn with or without the answer.
func (f *Flight[T]) Peek(ctx context.Context, key string, fetch func(context.Context) (T, error)) (T, bool) {
	if e, ok := f.fresh(key); ok {
		return e.val, e.err == nil
	}
	if ctx.Err() == nil {
		f.start(ctx, key, fetch)
	}
	e, ok := f.held(key)
	return e.val, ok
}

// start runs the fetch for key, detached from ctx, unless one is running,
// and returns where its answer lands.
func (f *Flight[T]) start(ctx context.Context, key string, fetch func(context.Context) (T, error)) <-chan singleflight.Result {
	return f.group.DoChan(key, func() (any, error) {
		// A reader arriving as the previous fetch lands must not start
		// another one.
		if e, ok := f.fresh(key); ok {
			return e, nil
		}
		// A deadline, not a bare cancel: clients tell a timeout from a
		// refusal by it (DeadlineExceeded), and say so in their errors.
		fctx, cancel := context.WithTimeoutCause(context.WithoutCancel(ctx), f.Timeout, errFetchTimeout)
		defer cancel()
		val, err := f.recovered(fctx, key, fetch)
		if err != nil && errors.Is(context.Cause(fctx), errFetchTimeout) {
			err = fmt.Errorf("%w: %w", errFetchTimeout, err)
		}
		return f.store(key, val, err), nil
	})
}

// recovered runs fetch and turns a panic into its error, logged with its
// stack: singleflight re-raises a panic on a goroutine of its own, out of
// any recover's reach, which would end the process rather than fail one key.
func (f *Flight[T]) recovered(ctx context.Context, key string, fetch func(context.Context) (T, error)) (val T, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("flight: fetch panicked: %v", r)
			cmp.Or(f.Logger, slog.Default()).Error("flight: fetch panicked", "key", key, "panic", r, "stack", string(debug.Stack()))
		}
	}()
	return fetch(ctx)
}

// held returns the last good answer for key, whatever its age.
func (f *Flight[T]) held(key string) (flightEntry[T], bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	e, ok := f.entries[key]
	return e, ok && e.err == nil
}

// store records an answer, evicting the oldest when full, and returns what
// readers get. A failure is remembered for ErrTTL; a stale flight keeps
// serving its last good answer instead, and asks again after ErrTTL.
func (f *Flight[T]) store(key string, val T, err error) flightEntry[T] {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stored++
	e := flightEntry[T]{val: val, err: err, at: time.Now(), seq: f.stored}
	if err != nil {
		if f.ErrTTL == 0 || errors.Is(err, ErrTransient) {
			return e
		}
		if old, ok := f.entries[key]; ok && f.Stale && old.err == nil {
			old.at = time.Now().Add(f.ErrTTL - f.TTL)
			f.entries[key] = old
			return old
		}
	}
	if f.entries == nil {
		f.entries = make(map[string]flightEntry[T])
	}
	if _, ok := f.entries[key]; !ok && len(f.entries) >= f.Max {
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

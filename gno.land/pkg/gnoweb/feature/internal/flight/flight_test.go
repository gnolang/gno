package flight

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestFlightServesAFreshAnswerWithoutFetching(t *testing.T) {
	t.Parallel()

	f := &Flight[int]{TTL: time.Minute, Timeout: time.Second, Max: 4}
	var fetches atomic.Int32
	fetch := func(context.Context) (int, error) { fetches.Add(1); return 7, nil }
	for range 3 {
		if v, err := f.Get(context.Background(), "k", fetch); v != 7 || err != nil {
			t.Fatalf("get = %d, %v", v, err)
		}
	}
	if n := fetches.Load(); n != 1 {
		t.Fatalf("fetched %d times, want once", n)
	}
}

// A stale answer comes back at once, and the refresh it started lands for
// the next reader.
func TestFlightServesStaleWhileRefreshing(t *testing.T) {
	t.Parallel()

	f := &Flight[int]{TTL: time.Minute, Timeout: time.Second, Max: 1, Stale: true}
	f.store("k", 1, nil)
	f.entries["k"] = flightEntry[int]{val: 1, at: time.Now().Add(-2 * time.Minute)}

	release := make(chan struct{})
	v, err := f.Get(context.Background(), "k", func(context.Context) (int, error) { <-release; return 2, nil })
	if v != 1 || err != nil {
		t.Fatalf("get = %d, %v; want the stale answer at once", v, err)
	}
	close(release)
	deadline := time.Now().Add(5 * time.Second)
	for {
		if e, ok := f.fresh("k"); ok && e.val == 2 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("the refresh never landed")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// A failure is remembered for errTTL, so a backend that is down is not asked
// again by every reader; a stale flight keeps its last good answer instead.
func TestFlightRemembersFailures(t *testing.T) {
	t.Parallel()

	boom := errors.New("down")
	f := &Flight[int]{TTL: time.Minute, ErrTTL: time.Minute, Timeout: time.Second, Max: 4}
	var fetches atomic.Int32
	fail := func(context.Context) (int, error) { fetches.Add(1); return 0, boom }
	for range 3 {
		if _, err := f.Get(context.Background(), "k", fail); !errors.Is(err, boom) {
			t.Fatalf("err = %v, want the failure", err)
		}
	}
	if n := fetches.Load(); n != 1 {
		t.Fatalf("fetched %d times, want once: the failure is remembered", n)
	}

	s := &Flight[int]{TTL: time.Minute, ErrTTL: time.Minute, Timeout: time.Second, Max: 1, Stale: true}
	s.entries = map[string]flightEntry[int]{"k": {val: 5, at: time.Now().Add(-2 * time.Minute)}}
	if e := s.store("k", 0, boom); e.err != nil || e.val != 5 {
		t.Fatalf("store = %+v, want the last good answer kept", e)
	}
	if e, ok := s.fresh("k"); !ok || e.val != 5 {
		t.Fatal("after a failed refresh the last good answer must hold for errTTL")
	}

	n := &Flight[int]{TTL: time.Minute, Timeout: time.Second, Max: 4}
	n.store("k", 0, boom)
	if _, ok := n.fresh("k"); ok {
		t.Fatal("with no errTTL a failure must not be remembered")
	}
}

func TestFlightIsBounded(t *testing.T) {
	t.Parallel()

	f := &Flight[int]{TTL: time.Minute, Timeout: time.Second, Max: 8}
	for i := range 20 {
		f.store(fmt.Sprint(i), i, nil)
	}
	if len(f.entries) != 8 {
		t.Fatalf("entries = %d, want the cap 8", len(f.entries))
	}
	if _, ok := f.fresh("0"); ok {
		t.Error("the oldest answer must be the one evicted")
	}
}

func TestFlightExpires(t *testing.T) {
	t.Parallel()

	f := &Flight[int]{TTL: time.Minute, Timeout: time.Second, Max: 1}
	f.entries = map[string]flightEntry[int]{"k": {val: 1, at: time.Now().Add(-2 * time.Minute)}}
	if _, ok := f.fresh("k"); ok {
		t.Error("an expired answer must not be served as fresh")
	}
}

// A reader already out of time starts no fetch: under an expired deadline a
// loop over keys would otherwise launch every fetch at once.
func TestFlightStartsNothingForAnExpiredReader(t *testing.T) {
	t.Parallel()

	f := &Flight[int]{TTL: time.Minute, Timeout: time.Second, Max: 4}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var fetches atomic.Int32
	_, err := f.Get(ctx, "k", func(context.Context) (int, error) { fetches.Add(1); return 1, nil })
	if !errors.Is(err, ErrPending) {
		t.Fatalf("err = %v, want ErrPending", err)
	}
	time.Sleep(20 * time.Millisecond)
	if fetches.Load() != 0 {
		t.Fatal("an expired reader started a fetch")
	}
}

// Only the reader's own wait is pending. A fetch that failed on its own
// deadline is a failure, and must not read as "still working".
func TestFlightTellsPendingFromFailedOnADeadline(t *testing.T) {
	t.Parallel()

	f := &Flight[int]{TTL: time.Minute, ErrTTL: time.Minute, Timeout: time.Second, Max: 4}
	_, err := f.Get(context.Background(), "k", func(context.Context) (int, error) {
		return 0, context.DeadlineExceeded
	})
	if err == nil || errors.Is(err, ErrPending) {
		t.Fatalf("err = %v, want the failure, not ErrPending", err)
	}
	if _, err := f.Get(context.Background(), "k", nil); errors.Is(err, ErrPending) {
		t.Fatal("a remembered failure must not read as pending")
	}

	slow := &Flight[int]{TTL: time.Minute, Timeout: time.Second, Max: 4}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err = slow.Get(ctx, "k", func(c context.Context) (int, error) { <-c.Done(); return 0, c.Err() })
	if !errors.Is(err, ErrPending) {
		t.Fatalf("err = %v, want ErrPending for the reader's own wait", err)
	}
}

// An answer marked partial is kept only as long as a failure, so a reload
// soon retries the reads that failed.
func TestFlightKeepsPartialAnswersBriefly(t *testing.T) {
	t.Parallel()

	f := &Flight[int]{TTL: time.Hour, ErrTTL: time.Minute, Timeout: time.Second, Max: 4, Partial: func(v int) bool { return v < 0 }}
	f.entries = map[string]flightEntry[int]{
		"partial":  {val: -1, at: time.Now().Add(-2 * time.Minute)},
		"complete": {val: 1, at: time.Now().Add(-2 * time.Minute)},
	}
	if _, ok := f.fresh("partial"); ok {
		t.Error("a partial answer outlived errTTL")
	}
	if _, ok := f.fresh("complete"); !ok {
		t.Error("a complete answer expired before ttl")
	}
}

// A fetch past its timeout is ended by a deadline, which clients report as
// a timeout, and Get says why.
func TestFlightTimeoutCancelsTheFetch(t *testing.T) {
	t.Parallel()

	f := &Flight[int]{TTL: time.Minute, Timeout: 10 * time.Millisecond, Max: 1}
	var seen error
	_, err := f.Get(context.Background(), "k", func(ctx context.Context) (int, error) {
		<-ctx.Done()
		seen = ctx.Err()
		return 0, ctx.Err()
	})
	if !errors.Is(seen, context.DeadlineExceeded) {
		t.Errorf("fetch ended with %v, want context.DeadlineExceeded", seen)
	}
	if !errors.Is(err, errFetchTimeout) {
		t.Errorf("get = %v, want errFetchTimeout", err)
	}
}

// A reader leaving does not end the fetch it started: the answer lands for
// the next reader.
func TestFlightFetchOutlivesItsReader(t *testing.T) {
	t.Parallel()

	f := &Flight[int]{TTL: time.Minute, Timeout: time.Second, Max: 1}
	ctx, cancel := context.WithCancel(context.Background())
	started, release := make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := f.Get(ctx, "k", func(fctx context.Context) (int, error) {
			close(started)
			<-release
			return 7, fctx.Err()
		})
		done <- err
	}()
	<-started
	cancel()
	if err := <-done; !errors.Is(err, ErrPending) {
		t.Fatalf("get = %v, want ErrPending for the reader who left", err)
	}
	close(release)
	// Joins the running fetch or reads what it stored: never fetches again.
	v, err := f.Get(context.Background(), "k", func(context.Context) (int, error) {
		t.Error("fetched again: the first fetch stored nothing")
		return 0, nil
	})
	if v != 7 || err != nil {
		t.Fatalf("get = %d, %v; want the answer of the fetch whose reader left", v, err)
	}
}

// peek never waits: with no answer it starts the fetch and returns nothing,
// while one runs it starts no other, then it returns the answer, and a stale
// one while it is refreshed. A failure is nothing at hand.
func TestFlightPeekNeverWaits(t *testing.T) {
	t.Parallel()

	f := &Flight[int]{TTL: time.Minute, ErrTTL: time.Minute, Timeout: time.Minute, Max: 1, Stale: true}
	release := make(chan struct{})
	var fetches atomic.Int32
	fetch := func(context.Context) (int, error) { fetches.Add(1); <-release; return 7, nil }

	// The fetch is blocked: a peek that waited for it would never return.
	for range 3 {
		if v, ok := f.Peek(context.Background(), "k", fetch); ok {
			t.Fatalf("peek = %d before any answer", v)
		}
	}
	close(release)
	waitFor(t, func() bool { _, ok := f.Peek(context.Background(), "k", fetch); return ok })
	if v, _ := f.Peek(context.Background(), "k", fetch); v != 7 {
		t.Fatalf("peek = %d, want 7", v)
	}
	if n := fetches.Load(); n != 1 {
		t.Fatalf("fetched %d times, want once", n)
	}

	// Stale: the old answer at once, and one refresh for the next reader.
	f.entries["k"] = flightEntry[int]{val: 7, at: time.Now().Add(-2 * time.Minute)}
	refreshed := make(chan struct{})
	if v, ok := f.Peek(context.Background(), "k", func(context.Context) (int, error) { <-refreshed; return 8, nil }); !ok || v != 7 {
		t.Fatalf("peek = %d, %v; want the stale answer", v, ok)
	}
	close(refreshed)
	waitFor(t, func() bool { v, _ := f.Peek(context.Background(), "k", fetch); return v == 8 })

	// A failure with no answer before it is nothing at hand; an expired
	// reader starts nothing.
	g := &Flight[int]{TTL: time.Minute, ErrTTL: time.Minute, Timeout: time.Minute, Max: 1, Stale: true}
	g.store("k", 0, errors.New("down"))
	if _, ok := g.Peek(context.Background(), "k", fetch); ok {
		t.Fatal("peek returned a failure as an answer")
	}
	done, cancel := context.WithCancel(context.Background())
	cancel()
	var started atomic.Int32
	h := &Flight[int]{TTL: time.Minute, Timeout: time.Minute, Max: 1}
	h.Peek(done, "k", func(context.Context) (int, error) { started.Add(1); return 1, nil })
	time.Sleep(10 * time.Millisecond)
	if started.Load() != 0 {
		t.Fatal("an expired reader started a fetch")
	}
}

func waitFor(t *testing.T, ok func() bool) {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); !ok(); time.Sleep(time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatal("timed out")
		}
	}
}

// A fetch that panics fails its key, briefly, rather than the process.
func TestFlightFailsAPanickingFetch(t *testing.T) {
	t.Parallel()

	var logs strings.Builder
	f := &Flight[int]{TTL: time.Minute, ErrTTL: time.Minute, Timeout: time.Second, Max: 1, Logger: slog.New(slog.NewTextHandler(&logs, nil))}
	if _, err := f.Get(context.Background(), "k", func(context.Context) (int, error) { panic("boom") }); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err = %v, want the panic", err)
	}
	if _, err := f.Get(context.Background(), "k", func(context.Context) (int, error) { return 1, nil }); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err = %v, want the panic remembered as a failure", err)
	}
	if out := logs.String(); !strings.Contains(out, "panic=boom") || !strings.Contains(out, "TestFlightFailsAPanickingFetch") {
		t.Fatalf("log = %q, want the panic and its stack", out)
	}
}

// Concurrent readers of one key share a single fetch.
func TestFlightCoalescesConcurrentMisses(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		f := &Flight[int]{TTL: time.Minute, Timeout: time.Minute, Max: 1}
		release := make(chan struct{})
		var fetches atomic.Int32
		fetch := func(context.Context) (int, error) {
			fetches.Add(1)
			<-release
			return 1, nil
		}
		var wg sync.WaitGroup
		for range 8 {
			wg.Go(func() {
				if v, err := f.Get(context.Background(), "k", fetch); v != 1 || err != nil {
					t.Errorf("Get = %d, %v", v, err)
				}
			})
		}
		synctest.Wait() // every reader waits on the key
		close(release)
		wg.Wait()
		if n := fetches.Load(); n != 1 {
			t.Fatalf("fetched %d times, want once", n)
		}
	})
}

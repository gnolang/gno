package chainmap

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestFlightServesAFreshAnswerWithoutFetching(t *testing.T) {
	t.Parallel()

	f := &flight[int]{ttl: time.Minute, timeout: time.Second, max: 4}
	var fetches atomic.Int32
	fetch := func(context.Context) (int, error) { fetches.Add(1); return 7, nil }
	for range 3 {
		if v, err := f.get(context.Background(), "k", fetch); v != 7 || err != nil {
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

	f := &flight[int]{ttl: time.Minute, timeout: time.Second, max: 1, stale: true}
	f.store("k", 1, nil)
	f.entries["k"] = flightEntry[int]{val: 1, at: time.Now().Add(-2 * time.Minute)}

	release := make(chan struct{})
	v, err := f.get(context.Background(), "k", func(context.Context) (int, error) { <-release; return 2, nil })
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
	f := &flight[int]{ttl: time.Minute, errTTL: time.Minute, timeout: time.Second, max: 4}
	var fetches atomic.Int32
	fail := func(context.Context) (int, error) { fetches.Add(1); return 0, boom }
	for range 3 {
		if _, err := f.get(context.Background(), "k", fail); !errors.Is(err, boom) {
			t.Fatalf("err = %v, want the failure", err)
		}
	}
	if n := fetches.Load(); n != 1 {
		t.Fatalf("fetched %d times, want once: the failure is remembered", n)
	}

	s := &flight[int]{ttl: time.Minute, errTTL: time.Minute, timeout: time.Second, max: 1, stale: true}
	s.entries = map[string]flightEntry[int]{"k": {val: 5, at: time.Now().Add(-2 * time.Minute)}}
	if e := s.store("k", 0, boom); e.err != nil || e.val != 5 {
		t.Fatalf("store = %+v, want the last good answer kept", e)
	}
	if e, ok := s.fresh("k"); !ok || e.val != 5 {
		t.Fatal("after a failed refresh the last good answer must hold for errTTL")
	}

	n := &flight[int]{ttl: time.Minute, timeout: time.Second, max: 4}
	n.store("k", 0, boom)
	if _, ok := n.fresh("k"); ok {
		t.Fatal("with no errTTL a failure must not be remembered")
	}
}

func TestFlightIsBounded(t *testing.T) {
	t.Parallel()

	f := &flight[int]{ttl: time.Minute, timeout: time.Second, max: 8}
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

	f := &flight[int]{ttl: time.Minute, timeout: time.Second, max: 1}
	f.entries = map[string]flightEntry[int]{"k": {val: 1, at: time.Now().Add(-2 * time.Minute)}}
	if _, ok := f.fresh("k"); ok {
		t.Error("an expired answer must not be served as fresh")
	}
}

// A reader already out of time starts no fetch: under an expired deadline a
// loop over keys would otherwise launch every fetch at once.
func TestFlightStartsNothingForAnExpiredReader(t *testing.T) {
	t.Parallel()

	f := &flight[int]{ttl: time.Minute, timeout: time.Second, max: 4}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var fetches atomic.Int32
	_, err := f.get(ctx, "k", func(context.Context) (int, error) { fetches.Add(1); return 1, nil })
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

	f := &flight[int]{ttl: time.Minute, errTTL: time.Minute, timeout: time.Second, max: 4}
	_, err := f.get(context.Background(), "k", func(context.Context) (int, error) {
		return 0, context.DeadlineExceeded
	})
	if err == nil || errors.Is(err, ErrPending) {
		t.Fatalf("err = %v, want the failure, not ErrPending", err)
	}
	if _, err := f.get(context.Background(), "k", nil); errors.Is(err, ErrPending) {
		t.Fatal("a remembered failure must not read as pending")
	}

	slow := &flight[int]{ttl: time.Minute, timeout: time.Second, max: 4}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err = slow.get(ctx, "k", func(c context.Context) (int, error) { <-c.Done(); return 0, c.Err() })
	if !errors.Is(err, ErrPending) {
		t.Fatalf("err = %v, want ErrPending for the reader's own wait", err)
	}
}

// An answer marked partial is kept only as long as a failure, so a reload
// soon retries the reads that failed.
func TestFlightKeepsPartialAnswersBriefly(t *testing.T) {
	t.Parallel()

	f := &flight[int]{ttl: time.Hour, errTTL: time.Minute, timeout: time.Second, max: 4, partial: func(v int) bool { return v < 0 }}
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

// A fetch past its timeout is cancelled, not given a deadline: the indexer
// client's breaker counts a deadline against the indexer, and one slow
// aggregate must not turn search off.
func TestFlightTimeoutCancelsTheFetch(t *testing.T) {
	t.Parallel()

	f := &flight[int]{ttl: time.Minute, timeout: 10 * time.Millisecond, max: 1}
	var seen error
	_, err := f.get(context.Background(), "k", func(ctx context.Context) (int, error) {
		if _, ok := ctx.Deadline(); ok {
			t.Error("the fetch was given a deadline")
		}
		<-ctx.Done()
		seen = ctx.Err()
		return 0, ctx.Err()
	})
	if !errors.Is(seen, context.Canceled) {
		t.Errorf("fetch ended with %v, want context.Canceled", seen)
	}
	if !errors.Is(err, errFetchTimeout) {
		t.Errorf("get = %v, want errFetchTimeout", err)
	}
}

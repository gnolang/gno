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

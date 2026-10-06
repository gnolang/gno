package chainmap

import (
	"context"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

// A week of one-second blocks is 604 800 of them; a tip past that puts the
// window's start inside the chain.
const weekOfBlocks = int(activityWindow / time.Second)

func TestComputeActivityCountsTheWindow(t *testing.T) {
	t.Parallel()

	tip := weekOfBlocks + 50_000
	from := tip - weekOfBlocks
	f := &fakeIndexer{
		tip: tip,
		t0:  time.Unix(0, 0),
		calls: map[int][]indexer.Tx{
			from - 1: {call(from-1, true, "g1old", "gno.land/r/a")}, // before the window
			from:     {call(from, true, "g1x", "gno.land/r/a")},
			from + 9: {call(from+9, true, "g1x", "gno.land/r/a"), call(from+9, true, "g1y", "gno.land/r/a")},
			tip - 1:  {call(tip-1, false, "g1z", "gno.land/r/a")}, // failed: not a call that happened
			tip:      {call(tip, true, "g1z", "gno.land/r/b")},
		},
	}

	a, err := computeActivity(context.Background(), f, nil)
	if err != nil {
		t.Fatalf("computeActivity: %v", err)
	}
	if a.From != from || a.To != tip {
		t.Fatalf("window = [%d, %d], want [%d, %d]", a.From, a.To, from, tip)
	}
	if got := a.Calls["gno.land/r/a"]; got != 3 {
		t.Errorf("calls to r/a = %d, want 3 (in window, successful)", got)
	}
	if got := a.Callers["gno.land/r/a"]; got != 2 {
		t.Errorf("callers of r/a = %d, want 2 distinct", got)
	}
	if got := a.Calls["gno.land/r/b"]; got != 1 {
		t.Errorf("calls to r/b = %d, want 1: the tip itself is in the window", got)
	}
	if a.Partial {
		t.Error("a fully read window must not be partial")
	}
}

// Bands must tile the window exactly: a gap drops calls, an overlap counts
// them twice.
func TestComputeActivityBandsTileTheWindow(t *testing.T) {
	t.Parallel()

	f := &fakeIndexer{tip: weekOfBlocks + 1234, t0: time.Unix(0, 0)}
	a, err := computeActivity(context.Background(), f, nil)
	if err != nil {
		t.Fatalf("computeActivity: %v", err)
	}

	covered := make(map[int]int)
	for _, b := range f.bandsQueried {
		for h := b[0] + 1; h <= b[1]; h++ {
			covered[h]++
		}
	}
	for h := a.From; h <= a.To; h++ {
		if covered[h] != 1 {
			t.Fatalf("height %d read %d times, want once", h, covered[h])
		}
	}
}

// A refresh reads only what can have changed: the band the window now starts
// in, the one it ends in, and any band closed since. Whole bands below the
// tip are final and come from the cache, and the counts stay the same.
func TestComputeActivityReadsClosedBandsOnce(t *testing.T) {
	t.Parallel()

	tip := weekOfBlocks + 50_000
	from := tip - weekOfBlocks
	f := &fakeIndexer{
		tip: tip,
		t0:  time.Unix(0, 0),
		calls: map[int][]indexer.Tx{
			from + 100_000: {call(from+100_000, true, "g1x", "gno.land/r/a")},
			from + 300_000: {call(from+300_000, true, "g1y", "gno.land/r/a")},
		},
	}
	closed := new(closedBands)
	if _, err := computeActivity(context.Background(), f, closed); err != nil {
		t.Fatalf("first count: %v", err)
	}

	f.tip += 30_000 // one more band closes, a new one opens
	f.bandsQueried = nil
	a, err := computeActivity(context.Background(), f, closed)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if n := len(f.bandsQueried); n > 4 {
		t.Errorf("refresh read %d bands, want at most 4: %v", n, f.bandsQueried)
	}
	if a.Calls["gno.land/r/a"] != 2 || a.Callers["gno.land/r/a"] != 2 {
		t.Errorf("r/a = %d calls by %d callers, want 2 by 2", a.Calls["gno.land/r/a"], a.Callers["gno.land/r/a"])
	}
	for lower := range closed.m {
		if lower+bandWidth <= a.From-1 {
			t.Errorf("band at %d left the window but is still kept", lower)
		}
	}
}

// The band the window starts in was cached whole when it sat inside the
// window; once the window starts within it, the heights before the start
// must not be counted.
func TestComputeActivityClipsTheCachedFirstBand(t *testing.T) {
	t.Parallel()

	tip := weekOfBlocks + 50_000
	// Inside the first window, and before the next one, in the band the
	// next window starts in.
	early := tip + 30_000 - weekOfBlocks - 5
	f := &fakeIndexer{
		tip:   tip,
		t0:    time.Unix(0, 0),
		calls: map[int][]indexer.Tx{early: {call(early, true, "g1x", "gno.land/r/a")}},
	}
	closed := new(closedBands)
	first, err := computeActivity(context.Background(), f, closed)
	if err != nil || first.Calls["gno.land/r/a"] != 1 {
		t.Fatalf("first count = %v, %v; want the call counted", first, err)
	}

	f.tip += 30_000
	a, err := computeActivity(context.Background(), f, closed)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if n := a.Calls["gno.land/r/a"]; n != 0 {
		t.Errorf("a call before the window was counted: %d", n)
	}
}

// A band the indexer caps is split until it fits, so a dense chain still
// gets a complete count.
func TestComputeActivitySplitsCappedBands(t *testing.T) {
	t.Parallel()

	tip := weekOfBlocks + 10
	f := &fakeIndexer{
		tip:     tip,
		t0:      time.Unix(0, 0),
		capOver: 5_000,
		calls:   map[int][]indexer.Tx{tip - 3: {call(tip-3, true, "g1x", "gno.land/r/a")}},
	}
	a, err := computeActivity(context.Background(), f, nil)
	if err != nil {
		t.Fatalf("computeActivity: %v", err)
	}
	if a.Partial {
		t.Error("splitting under the cap must still give a complete count")
	}
	if a.Calls["gno.land/r/a"] != 1 {
		t.Errorf("calls = %d, want 1", a.Calls["gno.land/r/a"])
	}
}

// A band still over the cap at the narrowest width is left out and the
// aggregate says so, rather than passing a hole for zero calls.
func TestComputeActivityMarksUnreadBandsPartial(t *testing.T) {
	t.Parallel()

	f := &fakeIndexer{tip: weekOfBlocks + 10, t0: time.Unix(0, 0), capOver: minBandWidth / 2}
	a, err := computeActivity(context.Background(), f, nil)
	if err != nil {
		t.Fatalf("computeActivity: %v", err)
	}
	if !a.Partial {
		t.Error("bands the indexer would not answer must mark the aggregate partial")
	}
}

func TestComputeActivityFailsWhenTheIndexerDoes(t *testing.T) {
	t.Parallel()

	f := &fakeIndexer{tip: weekOfBlocks + 10, t0: time.Unix(0, 0), failBands: true}
	if _, err := computeActivity(context.Background(), f, nil); err == nil {
		t.Fatal("a failed band must fail the aggregate, not count as no calls")
	}
}

// A chain younger than the window is counted from its first block.
func TestComputeActivityOnAYoungChain(t *testing.T) {
	t.Parallel()

	f := &fakeIndexer{
		tip:   100,
		t0:    time.Unix(0, 0),
		calls: map[int][]indexer.Tx{1: {call(1, true, "g1x", "gno.land/r/a")}},
	}
	a, err := computeActivity(context.Background(), f, nil)
	if err != nil {
		t.Fatalf("computeActivity: %v", err)
	}
	if a.From != 1 || a.Calls["gno.land/r/a"] != 1 {
		t.Fatalf("from = %d, calls = %d; want the whole chain counted", a.From, a.Calls["gno.land/r/a"])
	}
}

// A refresh that runs out of time fails as a whole, so the last good week
// stays served, and no band starts after it gave up.
func TestComputeActivityOutOfTimeFails(t *testing.T) {
	t.Parallel()

	f := &fakeIndexer{tip: weekOfBlocks + 10, t0: time.Unix(0, 0), bandDelay: 40 * time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := computeActivity(ctx, f, nil); err == nil {
		t.Fatal("a refresh out of time must fail, not return a partial week as fresh")
	}
	total := (weekOfBlocks + bandWidth) / bandWidth
	if n := len(f.bandsQueried); n >= total {
		t.Errorf("queried %d of %d bands: none should start after the deadline", n, total)
	}
}

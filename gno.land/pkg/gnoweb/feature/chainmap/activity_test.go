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

	a, err := computeActivity(context.Background(), f)
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
	a, err := computeActivity(context.Background(), f)
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
	a, err := computeActivity(context.Background(), f)
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
	a, err := computeActivity(context.Background(), f)
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
	if _, err := computeActivity(context.Background(), f); err == nil {
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
	a, err := computeActivity(context.Background(), f)
	if err != nil {
		t.Fatalf("computeActivity: %v", err)
	}
	if a.From != 1 || a.Calls["gno.land/r/a"] != 1 {
		t.Fatalf("from = %d, calls = %d; want the whole chain counted", a.From, a.Calls["gno.land/r/a"])
	}
}

// A refresh that runs out of time keeps what it counted and says the rest is
// unknown, rather than failing a week of counts for its last bands.
func TestComputeActivityOutOfTimeIsPartial(t *testing.T) {
	t.Parallel()

	f := &fakeIndexer{tip: weekOfBlocks + 10, t0: time.Unix(0, 0), bandDelay: 40 * time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	a, err := computeActivity(ctx, f)
	if err != nil {
		t.Fatalf("computeActivity: %v, want a partial aggregate", err)
	}
	if !a.Partial {
		t.Fatal("bands left unread must mark the aggregate partial")
	}
	total := (a.To - a.From + bandWidth) / bandWidth
	if n := len(f.bandsQueried); n >= total {
		t.Errorf("queried %d of %d bands: none should start after the deadline", n, total)
	}
}

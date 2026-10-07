package chainmap

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

func callTx(h int, ok bool, gas int, pkgs ...string) indexer.Tx {
	tx := indexer.Tx{Height: h, Success: ok, GasUsed: gas}
	for _, p := range pkgs {
		var m indexer.Message
		m.Value.Type, m.Value.PkgPath, m.Value.Func, m.Value.Caller = "MsgCall", p, "F"+strconv.Itoa(h), "g1c"
		tx.Messages = append(tx.Messages, m)
	}
	return tx
}

// A band keeps each realm's newest calls, failed ones included, with the
// calls batched beside them; counts and gas still take successful ones only.
func TestCountBandKeepsRecentCalls(t *testing.T) {
	t.Parallel()

	var txs []indexer.Tx
	for h := 1; h <= recentCallsKept+3; h++ {
		txs = append(txs, callTx(h, true, 100, "gno.land/r/a"))
	}
	txs = append(txs, callTx(50, false, 7, "gno.land/r/a", "gno.land/r/b"))

	b := countBand(txs, true)
	got := b.recent["gno.land/r/a"]
	if len(got) != recentCallsKept || got[0].Height != 50 || !got[0].Failed || got[0].Others != 1 || got[0].Func != "F50" {
		t.Fatalf("recent = %+v, want %d newest, the failed batched one first", got, recentCallsKept)
	}
	if b.calls["gno.land/r/a"] != recentCallsKept+3 {
		t.Errorf("calls = %d: a failed call must not be counted", b.calls["gno.land/r/a"])
	}
}

// Across bands the aggregate keeps the newest calls overall, and gives each
// one a time placed between the window's first and last block.
func TestComputeActivityMergesRecentCalls(t *testing.T) {
	t.Parallel()

	tip := weekOfBlocks + 50_000
	f := &fakeIndexer{tip: tip, t0: time.Unix(0, 0), calls: map[int][]indexer.Tx{
		tip - 10:           {callTx(tip-10, true, 1, "gno.land/r/a")},
		tip - 100_000:      {callTx(tip-100_000, true, 1, "gno.land/r/a")},
		tip - weekOfBlocks: {callTx(tip-weekOfBlocks, true, 1, "gno.land/r/a")},
	}}
	a, err := computeActivity(context.Background(), f, nil)
	if err != nil {
		t.Fatalf("computeActivity: %v", err)
	}
	got := a.Recent["gno.land/r/a"]
	if len(got) != 3 || got[0].Height != tip-10 || got[2].Height != tip-weekOfBlocks {
		t.Fatalf("recent = %+v, want the three calls newest first", got)
	}
	// Blocks are one second apart in the fake: the estimate is exact here.
	if want := f.t0.Add(time.Duration(tip-10) * time.Second); !a.timeOf(tip - 10).Equal(want) {
		t.Errorf("timeOf = %v, want %v", a.timeOf(tip-10), want)
	}
}

// The section says why it has no rows: still counting, unreadable, or no
// call in the window, which a partial window cannot claim.
func TestCallsSectionStates(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_000_000, 0)
	a := &Activity{From: 1, To: 100, Since: now.Add(-100 * time.Second), Until: now, Recent: map[string][]Call{
		"gno.land/r/a": {{Func: "Approve", Caller: "g1x", Height: 40, Gas: 154_000_000, Others: 1}},
	}}
	for _, c := range []struct {
		name  string
		a     *Activity
		err   error
		check func(*components.CallsSection) bool
	}{
		{"pending", nil, ErrPending, func(s *components.CallsSection) bool { return s.Pending() }},
		{"failed", nil, errors.New("down"), func(s *components.CallsSection) bool { return s.Unavailable() }},
		{"quiet", a, nil, func(s *components.CallsSection) bool { return len(s.Rows) == 0 && s.Empty() }},
		{"quiet but partial", &Activity{Partial: true}, nil, func(s *components.CallsSection) bool { return !s.Empty() && s.Partial }},
	} {
		s := callsSection(c.a, c.err, "gno.land/r/quiet", now, "https://indexer.test")
		if !c.check(s) {
			t.Errorf("%s: section = %+v", c.name, s)
		}
	}
	s := callsSection(a, nil, "gno.land/r/a", now, "https://indexer.test")
	if len(s.Rows) != 1 || s.Rows[0].Gas != "154 M" || s.Rows[0].Batch != "tx of 2 calls" || s.Rows[0].Ago != "1 min ago" {
		t.Errorf("row = %+v", s.Rows)
	}
}

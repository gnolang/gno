package chainmap

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

func msg(typ, pkg, fn string) indexer.Message {
	var m indexer.Message
	m.Value.Type, m.Value.PkgPath, m.Value.Func, m.Value.Caller = typ, pkg, fn, "g1c"
	return m
}

func callTx(h int, ok bool, gas int, pkgs ...string) indexer.Tx {
	tx := indexer.Tx{Height: h, Success: ok, GasUsed: gas}
	for _, p := range pkgs {
		tx.Messages = append(tx.Messages, msg("MsgCall", p, "F"+strconv.Itoa(h)))
	}
	return tx
}

// A transaction is one row per realm: two calls into the same realm read
// "Approve, Deposit", and the row counts the calls the transaction held (a
// run is not a call).
func TestRecordCallsOneRowPerTransaction(t *testing.T) {
	t.Parallel()

	both := indexer.Tx{Height: 9, Success: true, GasUsed: 100, Messages: []indexer.Message{
		msg("MsgCall", "gno.land/r/a", "Approve"), msg("MsgCall", "gno.land/r/a", "Deposit"), msg("MsgCall", "gno.land/r/b", "Swap"),
	}}
	withRun := indexer.Tx{Height: 8, Success: true, GasUsed: 100, Messages: []indexer.Message{msg("MsgCall", "gno.land/r/a", "Withdraw"), msg("MsgRun", "", "")}}
	b := countBand([]indexer.Tx{both, withRun}, true)

	got := b.recent["gno.land/r/a"]
	if len(got) != 2 || got[0].Func != "Approve, Deposit" || got[0].Calls != 3 || got[1].Calls != 1 {
		t.Fatalf("rows = %+v, want one row per transaction: 3 calls, then 1", got)
	}
}

// Failed calls are kept only for a realm that also has a successful one, at
// most maxFailedKept per realm: anyone can send failed calls cheaply, to any
// path, and must neither grow the aggregate nor push real calls out.
func TestRecordCallsBoundsFailedCalls(t *testing.T) {
	t.Parallel()

	var txs []indexer.Tx
	for h := 1; h <= 10; h++ {
		txs = append(txs, callTx(100+h, false, 1, "gno.land/r/a"), callTx(100+h, false, 1, "gno.land/r/ghost"+strconv.Itoa(h)))
	}
	txs = append(txs, callTx(5, true, 1, "gno.land/r/a"), callTx(4, true, 1, "gno.land/r/a"))
	b := countBand(txs, true)

	if len(b.recent) != 1 {
		t.Errorf("recent holds %d realms, want only r/a: a realm with only failed calls is not kept", len(b.recent))
	}
	failed := 0
	for _, c := range b.recent["gno.land/r/a"] {
		if c.Failed {
			failed++
		}
	}
	if failed != maxFailedKept || len(b.recent["gno.land/r/a"]) != maxFailedKept+2 {
		t.Errorf("r/a keeps %d failed of %d, want %d failed and both successful ones", failed, len(b.recent["gno.land/r/a"]), maxFailedKept)
	}
}

// A function name is a Go identifier of any length: kept short.
func TestRecordCallsClipsFunctionNames(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("X", 5000)
	b := countBand([]indexer.Tx{{Height: 1, Success: true, Messages: []indexer.Message{msg("MsgCall", "gno.land/r/a", long)}}}, true)
	if f := b.recent["gno.land/r/a"][0].Func; len(f) > maxFuncKept+len("…") {
		t.Errorf("func kept at %d bytes", len(f))
	}
}

// newest keeps no more than it returns: trimmed calls must not stay alive in
// the slice's backing array for the week.
func TestNewestDropsTheRest(t *testing.T) {
	t.Parallel()

	calls := make([]Call, 200)
	for i := range calls {
		calls[i].Height = i
	}
	if got := newest(calls); cap(got) > recentCallsKept {
		t.Errorf("cap = %d, want at most %d", cap(got), recentCallsKept)
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
		"gno.land/r/a": {{Func: "Approve", Caller: "g1x", Height: 40, Gas: 154_000_000, Calls: 2}},
	}}
	if s := callsSection(nil, ErrPending, "gno.land/r/a", now, ""); !s.Pending {
		t.Error("pending")
	}
	if s := callsSection(nil, errors.New("down"), "gno.land/r/a", now, ""); !s.Unavailable {
		t.Error("unavailable")
	}
	if s := callsSection(a, nil, "gno.land/r/quiet", now, ""); !s.Empty() {
		t.Error("quiet realm must be empty")
	}
	if s := callsSection(&Activity{Partial: true}, nil, "gno.land/r/quiet", now, ""); s.Empty() || !s.Partial {
		t.Error("a partial window must not claim no calls")
	}
	s := callsSection(a, nil, "gno.land/r/a", now, "")
	if len(s.Rows) != 1 || s.Rows[0].Gas != "154 M" || s.Rows[0].Batch != "tx of 2 calls" || s.Rows[0].Ago != "1 minute ago" || s.Rows[0].Height != 40 {
		t.Errorf("row = %+v", s.Rows)
	}
}

// Before the aggregate first exists an overview does not wait for it: the
// first scan takes seconds and a page would only wait to say "pending".
func TestCallsSectionDoesNotWaitForAColdScan(t *testing.T) {
	t.Parallel()

	f := &fakeIndexer{tip: weekOfBlocks + 50_000, t0: time.Unix(0, 0), bandDelay: 5 * time.Second}
	h := New(Deps{Indexer: f, Imports: fakeImports{}, Domain: "gno.land"})
	start := time.Now()
	s := h.CallsSection(context.Background(), "gno.land/r/a", time.Now())
	if d := time.Since(start); d > 300*time.Millisecond || !s.Pending {
		t.Errorf("cold section took %v (pending %v), want an immediate pending", d, s.Pending)
	}
}

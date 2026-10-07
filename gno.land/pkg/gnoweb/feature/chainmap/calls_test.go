package chainmap

import (
	"context"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

// A package's recent calls are read once per TTL whatever the traffic, keep
// only calls into the package, newest first, and stop at the page's length.
func TestRecentCallsAreCachedAndFiltered(t *testing.T) {
	t.Parallel()

	call := func(h int, pkg string) indexer.Tx {
		tx := indexer.Tx{Height: h, Success: true, GasUsed: 1_000_000, Messages: make([]indexer.Message, 1)}
		tx.Messages[0].Value.Type, tx.Messages[0].Value.PkgPath, tx.Messages[0].Value.Func = "MsgCall", pkg, "F"
		return tx
	}
	f := &fakeIndexer{tip: 1000, t0: time.Now().Add(-1000 * time.Second)}
	for h := 1; h <= recentCallsShown+5; h++ {
		f.recent = append(f.recent, call(h, "gno.land/r/a"))
	}
	f.recent = append(f.recent, call(999, "gno.land/r/other"))
	h := New(Deps{Indexer: f, Imports: fakeImports{}, Domain: "gno.land"})

	got, err := h.RecentCalls(context.Background(), "gno.land/r/a")
	if err != nil {
		t.Fatalf("RecentCalls: %v", err)
	}
	if len(got.Calls) != recentCallsShown || got.Calls[0].Height != recentCallsShown+5 {
		t.Fatalf("got %d calls starting at %d, want %d newest first", len(got.Calls), got.Calls[0].Height, recentCallsShown)
	}
	if _, err := h.RecentCalls(context.Background(), "gno.land/r/a"); err != nil || f.recentCalls != 1 {
		t.Errorf("a second read queried the indexer again (%d queries)", f.recentCalls)
	}
}

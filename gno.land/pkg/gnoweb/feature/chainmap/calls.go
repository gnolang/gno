package chainmap

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

const (
	// recentCallsShown is how many calls the overview lists.
	recentCallsShown = 8
	// recentCallsRead is how many transactions are asked for: the indexer's
	// answer also holds the package's deploys and rows it could not decode,
	// which are not calls.
	recentCallsRead = 24
	// recentCallsTTL is how long a package's answer is served: the list is a
	// glance at activity, not a live feed.
	recentCallsTTL = 2 * time.Minute
	// recentCallsTimeout bounds one lookup: two indexer queries.
	recentCallsTimeout = 8 * time.Second
	// recentCallsWait is how long an overview waits before rendering
	// without them. The lookup carries on for the next load.
	recentCallsWait = 1500 * time.Millisecond
	// maxRecentCallsEntries bounds the cache, one entry per package.
	maxRecentCallsEntries = 512
)

// Call is one call into a package, as the indexer recorded it.
type Call struct {
	Func, Caller string
	Height, Gas  int
	Failed       bool
	// Others counts the other calls and runs in the same transaction.
	Others int
	// Time is the block's time; zero when the indexer did not give it.
	Time time.Time
}

// RecentCalls are a package's last calls, newest first.
type RecentCalls struct {
	Calls []Call
	// AsOf is the indexer's last block when they were read.
	AsOf int
}

func newRecentCallsFlight() *flight[*RecentCalls] {
	return &flight[*RecentCalls]{ttl: recentCallsTTL, errTTL: failureTTL, timeout: recentCallsTimeout, max: maxRecentCallsEntries}
}

// RecentCalls returns the last calls into pkgPath, a fully qualified package
// path, read at most once per recentCallsTTL whatever the traffic.
func (h *Handler) RecentCalls(ctx context.Context, pkgPath string) (*RecentCalls, error) {
	if !h.HasIndexer() {
		return nil, errors.New("no indexer configured")
	}
	return h.recentCalls.get(ctx, pkgPath, func(ctx context.Context) (*RecentCalls, error) {
		return readRecentCalls(ctx, h.deps.Indexer, pkgPath)
	})
}

func readRecentCalls(ctx context.Context, idx Indexer, pkgPath string) (*RecentCalls, error) {
	asOf, err := idx.LatestBlockHeight(ctx)
	if err != nil {
		return nil, err
	}
	txs, err := idx.RecentByPackage(ctx, pkgPath, recentCallsRead)
	if err != nil {
		return nil, err
	}
	rc := &RecentCalls{AsOf: asOf}
	for _, tx := range txs {
		// The filter matched the package's deploys too, and rows the indexer
		// could not decode: only a MsgCall into the package is a call here.
		i := slices.IndexFunc(tx.Messages, func(m indexer.Message) bool {
			return m.Type() == "MsgCall" && m.Path() == pkgPath
		})
		if i < 0 {
			continue
		}
		others := 0
		for j, m := range tx.Messages {
			if j != i && (m.Type() == "MsgCall" || m.Type() == "MsgRun") {
				others++
			}
		}
		m := tx.Messages[i].Value
		rc.Calls = append(rc.Calls, Call{Func: m.Func, Caller: m.Caller, Height: tx.Height, Gas: tx.GasUsed, Failed: !tx.Success, Others: others})
	}
	slices.SortStableFunc(rc.Calls, func(a, b Call) int { return b.Height - a.Height })
	rc.Calls = rc.Calls[:min(len(rc.Calls), recentCallsShown)]

	heights := make([]int, 0, len(rc.Calls))
	for _, c := range rc.Calls {
		heights = append(heights, c.Height)
	}
	times, err := idx.BlockTimes(ctx, slices.Compact(heights))
	if err != nil {
		// The calls stand without their age: each row then shows its block.
		return rc, nil
	}
	for i := range rc.Calls {
		rc.Calls[i].Time = times[rc.Calls[i].Height]
	}
	return rc, nil
}

// CallsSection is the overview's Recent calls for pkgPath, a fully qualified
// path, or nil without an indexer. It waits recentCallsWait at most.
func (h *Handler) CallsSection(ctx context.Context, pkgPath string, now time.Time) *components.CallsSection {
	if !h.HasIndexer() {
		return nil
	}
	wctx, cancel := context.WithTimeout(ctx, recentCallsWait)
	defer cancel()
	rc, err := h.RecentCalls(wctx, pkgPath)
	switch {
	case errors.Is(err, ErrPending):
		return &components.CallsSection{State: components.CallsPending}
	case err != nil:
		h.deps.Logger.Warn("overview: recent calls unavailable", "path", pkgPath, "error", err)
		return &components.CallsSection{State: components.CallsUnavailable}
	}
	s := &components.CallsSection{Indexer: &components.IndexerStatus{URL: h.deps.Indexer.URL(), LastBlock: rc.AsOf}}
	for _, c := range rc.Calls {
		s.Rows = append(s.Rows, components.CallRow{
			Func: c.Func, Caller: c.Caller, Failed: c.Failed, Others: c.Others,
			Gas: MetricGas.format(int((int64(c.Gas) + 999_999) / 1_000_000)),
			Ago: ago(c, now),
		})
	}
	return s
}

// ago is how long before now c's block was, or its height when the indexer
// gave no time.
func ago(c Call, now time.Time) string {
	if c.Time.IsZero() {
		return fmt.Sprintf("block %d", c.Height)
	}
	switch d := now.Sub(c.Time); {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d min ago", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%d h ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%d d ago", int(d.Hours()/24))
	}
}

package chainmap

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

// The overview's Recent calls come from the 7-day activity scan the map
// already makes: each band keeps every realm's newest calls, so listing them
// costs the indexer nothing more, whatever the traffic, and a path nobody
// calls costs nothing at all.
const (
	// recentCallsKept is how many calls a realm keeps, per band and overall.
	recentCallsKept = 8
	// recentCallsWait is how long an overview waits for the aggregate before
	// rendering without it; once computed it is served at once.
	recentCallsWait = 1500 * time.Millisecond
)

// Call is one call into a realm, as the indexer recorded it.
type Call struct {
	Func, Caller string
	Height, Gas  int
	Failed       bool
	// Others counts the other calls, runs and deploys in the transaction,
	// which Gas, the transaction's, also paid for.
	Others int
}

// recordCalls adds tx's calls to the newest calls kept per realm.
func recordCalls(recent map[string][]Call, tx indexer.Tx) {
	workers := 0
	for _, m := range tx.Messages {
		switch m.Type() {
		case "MsgCall", "MsgRun", "MsgAddPackage":
			workers++
		}
	}
	for _, m := range tx.Messages {
		p := m.Path()
		if m.Type() != "MsgCall" || p == "" {
			continue
		}
		recent[p] = newest(append(recent[p], Call{
			Func: m.Value.Func, Caller: m.Value.Caller, Height: tx.Height, Gas: tx.GasUsed,
			Failed: !tx.Success, Others: workers - 1,
		}))
	}
}

// newest sorts calls newest first and keeps recentCallsKept of them.
func newest(calls []Call) []Call {
	slices.SortStableFunc(calls, func(a, b Call) int { return cmp.Compare(b.Height, a.Height) })
	return calls[:min(len(calls), recentCallsKept)]
}

// timeOf places a height of the window in time, between the times of its
// first and last block: the indexer has no time per transaction, and blocks
// come at a steady rate.
func (a *Activity) timeOf(height int) time.Time {
	if a.To <= a.From {
		return a.Until
	}
	f := float64(height-a.From) / float64(a.To-a.From)
	return a.Since.Add(time.Duration(f * float64(a.Until.Sub(a.Since))))
}

// CallsSection is the overview's Recent calls for pkgPath, a fully qualified
// realm path, or nil without an indexer. The caller decides it applies: a
// realm, at the latest height.
func (h *Handler) CallsSection(ctx context.Context, pkgPath string, now time.Time) *components.CallsSection {
	if !h.HasIndexer() {
		return nil
	}
	wctx, cancel := context.WithTimeout(ctx, recentCallsWait)
	defer cancel()
	a, err := h.activity.get(wctx, "activity", func(ctx context.Context) (*Activity, error) {
		return computeActivity(ctx, h.deps.Indexer, h.closedBands)
	})
	if err != nil && !errors.Is(err, ErrPending) {
		h.deps.Logger.Warn("overview: recent calls unavailable", "error", err)
	}
	return callsSection(a, err, pkgPath, now, h.deps.Indexer.URL())
}

// callsSection writes the aggregate's calls into pkgPath for display, or the
// state that stands for them.
func callsSection(a *Activity, err error, pkgPath string, now time.Time, indexerURL string) *components.CallsSection {
	switch {
	case errors.Is(err, ErrPending):
		return &components.CallsSection{State: components.CallsPending}
	case err != nil || a == nil:
		return &components.CallsSection{State: components.CallsUnavailable}
	}
	s := &components.CallsSection{Partial: a.Partial, Indexer: &components.IndexerStatus{URL: indexerURL, LastBlock: a.To}}
	for _, c := range a.Recent[pkgPath] {
		row := components.CallRow{
			Func: c.Func, Caller: c.Caller, Failed: c.Failed, Height: c.Height,
			Gas: FormatGas(int64(c.Gas)),
		}
		if c.Others > 0 {
			row.Batch = "tx of " + strconv.Itoa(c.Others+1) + " calls"
		}
		if !a.Until.IsZero() {
			t := a.timeOf(c.Height)
			row.Time, row.Ago = t.UTC().Format(time.RFC3339), ago(now.Sub(t))
		}
		s.Rows = append(s.Rows, row)
	}
	return s
}

// ago writes a duration as how long ago something was.
func ago(d time.Duration) string {
	switch {
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

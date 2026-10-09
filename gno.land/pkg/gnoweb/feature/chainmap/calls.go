package chainmap

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"strconv"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

// The overview's Recent calls come from the 7-day activity scan the map
// already makes: each band keeps every realm's newest calls, so listing them
// adds no indexer query per page, and a path nobody calls costs nothing.
const (
	// recentCallsKept is how many rows a realm keeps, per band and overall.
	recentCallsKept = 8
	// maxFailedKept bounds the failed calls among them: anyone can send
	// failed calls cheaply, and must not push the real ones out.
	maxFailedKept = 2
	// maxFuncKept bounds a function name, a Go identifier of any length.
	maxFuncKept = 64
	// recentCallsWait is how long an overview waits for a refresh of the
	// aggregate; once one exists, it is served at once.
	recentCallsWait = 1500 * time.Millisecond
)

// Call is one transaction's calls into a realm: one row on the overview.
type Call struct {
	// Func names the functions called, in order ("Approve, Deposit").
	Func, Caller string
	Height       int
	// Gas is the transaction's, which every call in it shared.
	Gas    int64
	Failed bool
	// Calls counts all the calls in the transaction, into any realm.
	Calls int
}

// recordCalls adds tx as one row to each realm it calls.
func recordCalls(recent map[string][]Call, tx indexer.Tx) {
	rows := make(map[string]*Call)
	var order []string
	total := 0
	for _, m := range tx.Messages {
		p := m.Path()
		if m.Type() != "MsgCall" || p == "" {
			continue
		}
		total++
		fn := m.Value.Func
		if len(fn) > maxFuncKept {
			fn = fn[:maxFuncKept] + "…"
		}
		if r, ok := rows[p]; ok {
			r.Func += ", " + fn
			continue
		}
		rows[p] = &Call{Func: fn, Caller: m.Value.Caller, Height: tx.Height, Gas: int64(tx.GasUsed), Failed: !tx.Success}
		order = append(order, p)
	}
	for _, p := range order {
		r := rows[p]
		r.Calls = total
		recent[p] = newest(append(recent[p], *r))
	}
}

// newest sorts calls newest first and keeps recentCallsKept of them, at most
// maxFailedKept failed, in a slice of their own: trimmed calls must not stay
// alive in the backing array.
func newest(calls []Call) []Call {
	slices.SortStableFunc(calls, func(a, b Call) int { return cmp.Compare(b.Height, a.Height) })
	out := make([]Call, 0, min(len(calls), recentCallsKept))
	failed := 0
	for _, c := range calls {
		if len(out) == recentCallsKept {
			break
		}
		if c.Failed {
			if failed == maxFailedKept {
				continue
			}
			failed++
		}
		out = append(out, c)
	}
	return out
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
	a, err := h.loadActivity(ctx, recentCallsWait)
	if err != nil && !errors.Is(err, ErrPending) {
		h.deps.Logger.Debug("overview: recent calls unavailable", "error", err)
	}
	return callsSection(a, err, pkgPath, now, h.deps.Indexer.URL())
}

// callsSection writes the aggregate's calls into pkgPath for display, or the
// state that stands for them.
func callsSection(a *Activity, err error, pkgPath string, now time.Time, indexerURL string) *components.CallsSection {
	switch {
	case errors.Is(err, ErrPending):
		return &components.CallsSection{Pending: true}
	case err != nil || a == nil:
		return &components.CallsSection{Unavailable: true}
	}
	s := &components.CallsSection{Partial: a.Partial, Indexer: &components.IndexerStatus{URL: indexerURL, LastBlock: a.To}}
	for _, c := range a.Recent[pkgPath] {
		row := components.CallRow{
			Func: c.Func, Caller: c.Caller, Failed: c.Failed, Height: c.Height,
			Gas: components.FormatGas(c.Gas),
		}
		if c.Calls > 1 {
			row.Batch = "tx of " + strconv.Itoa(c.Calls) + " calls"
		}
		if !a.Until.IsZero() {
			t := a.timeOf(c.Height)
			row.Time, row.Ago = t.UTC().Format(time.RFC3339), components.FormatRelativeTime(now, t)
		}
		s.Rows = append(s.Rows, row)
	}
	return s
}

package components

// Recent-calls states the section tells apart, so an indexer that could not
// answer never reads as a realm nobody calls.
const (
	CallsPending     = "pending"
	CallsUnavailable = "unavailable"
)

// CallsSection is the overview's Recent calls: the last calls into the realm
// over the activity window, from an indexer. Nil when no indexer is
// configured; the overview then has no such section at all.
type CallsSection struct {
	Rows []CallRow
	// State is CallsPending or CallsUnavailable when there is nothing to
	// trust; empty otherwise.
	State string
	// Partial is set when part of the window could not be read: no rows then
	// says nothing.
	Partial bool
	// Indexer is the provenance footer, set when the indexer answered.
	Indexer *IndexerStatus
}

// Pending, Unavailable and Empty name the section's states for the template.
func (s *CallsSection) Pending() bool     { return s.State == CallsPending }
func (s *CallsSection) Unavailable() bool { return s.State == CallsUnavailable }
func (s *CallsSection) Empty() bool       { return s.State == "" && !s.Partial && len(s.Rows) == 0 }

// CallRow is one call, written for display.
type CallRow struct {
	// Func is the function called, Caller the full address that signed.
	Func, Caller string
	// Gas is the transaction's gas; Batch says how many calls that
	// transaction held when it held more than this one.
	Gas, Batch string
	// Time is the block's estimated time (RFC 3339), Ago the same relative
	// to now; both empty when the indexer gave no times.
	Time, Ago string
	Height    int
	Failed    bool
}

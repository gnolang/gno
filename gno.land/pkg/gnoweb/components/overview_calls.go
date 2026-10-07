package components

// Recent-calls states the section tells apart, so an indexer that could not
// answer never reads as a realm nobody calls.
const (
	CallsPending     = "pending"
	CallsUnavailable = "unavailable"
)

// CallsSection is the overview's Recent calls: the last calls into the realm,
// from an indexer. Nil when no indexer is configured; the overview then has
// no such section at all.
type CallsSection struct {
	Rows []CallRow
	// State is CallsPending or CallsUnavailable when there are no rows to
	// trust; empty otherwise.
	State string
	// Indexer is the provenance footer, set when the indexer answered.
	Indexer *IndexerStatus
}

// CallRow is one call, written for display.
type CallRow struct {
	// Func is the function called, Caller the full address that signed.
	Func, Caller string
	// Gas is the whole transaction's gas, Ago how long ago its block was.
	Gas, Ago string
	Failed   bool
	// Others counts the other calls batched in the same transaction.
	Others int
}

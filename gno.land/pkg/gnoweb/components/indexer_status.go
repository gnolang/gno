package components

// IndexerStatus is the provenance footer of a page showing indexer data
// (ui/indexer_status). Indexer data is not consensus data, so the reader gets
// how far it has indexed.
type IndexerStatus struct {
	LastBlock int
}

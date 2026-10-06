package components

// IndexerStatus is the provenance footer of a page showing indexer data
// (ui/indexer_status). Indexer data is not consensus data, so the reader gets
// the endpoint and how far it has indexed.
type IndexerStatus struct {
	URL       string
	LastBlock int
}

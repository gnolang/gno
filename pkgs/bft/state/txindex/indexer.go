package txindex

// TxIndexer interface defines methods to index and search transactions.
type TxIndexer interface {
	/*
		// AddBatch analyzes, indexes and stores a batch of transactions.
		AddBatch(b *Batch) error

		// Index analyzes, indexes and stores a single transaction.
		Index(result *types.TxResult) error

		// Get returns the transaction specified by hash or nil if the transaction is not indexed
		// or stored.
		Get(hash []byte) (*types.TxResult, error)

		// Search allows you to query for transactions.
		Search(q *query.Query) ([]*types.TxResult, error)
	*/
}

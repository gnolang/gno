package null

import (
	"github.com/gnolang/gno/tm2/pkg/bft/state/txindex"
)

var _ txindex.TxIndexer = (*TxIndex)(nil)

// TxIndex acts as a /dev/null.
type TxIndex struct{}

/*
// Get on a TxIndex is disabled and panics when invoked.
func (txi *TxIndex) Get(hash []byte) (*types.TxResult, error) {
	return nil, errors.New(`Indexing is disabled (set 'tx_index = "kv"' in config)`)
}

// AddBatch is a noop and always returns nil.
func (txi *TxIndex) AddBatch(batch *txindex.Batch) error {
	return nil
}

// Index is a noop and always returns nil.
func (txi *TxIndex) Index(result *types.TxResult) error {
	return nil
}

func (txi *TxIndex) Search(q *query.Query) ([]*types.TxResult, error) {
	return []*types.TxResult{}, nil
}
*/

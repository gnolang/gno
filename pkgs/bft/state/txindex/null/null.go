package null

import (
	"github.com/gnolang/gno/pkgs/bft/state/txindex"
	"github.com/gnolang/gno/pkgs/bft/types"
)

var _ txindex.TxIndexer = (*TxIndexer)(nil)

const (
	IndexerType = "none"
)

// TxIndexer acts as a /dev/null
type TxIndexer struct{}

func NewNullIndexer() *TxIndexer {
	return &TxIndexer{}
}

func (t TxIndexer) Start() error {
	return nil
}

func (t TxIndexer) Stop() error {
	return nil
}

func (t TxIndexer) Index(_ types.TxResult) error {
	return nil
}

func (t TxIndexer) GetType() string {
	return IndexerType
}

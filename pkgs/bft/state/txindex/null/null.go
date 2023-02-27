package null

import (
	"github.com/gnolang/gno/pkgs/bft/state/txindex"
	"github.com/gnolang/gno/pkgs/bft/types"
)

var _ txindex.TxIndexer = (*TxIndex)(nil)

// TxIndex acts as a /dev/null.
type TxIndex struct{}

func (t TxIndex) Start() error {
	return nil
}

func (t TxIndex) Close() error {
	return nil
}

func (t TxIndex) Index(_ *types.TxResult) error {
	return nil
}

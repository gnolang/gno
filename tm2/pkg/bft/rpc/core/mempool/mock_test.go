package mempool

import (
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

type (
	checkTxDelegate    func(tx types.Tx, cb func(abci.Response)) error
	reapMaxTxsDelegate func(max int) []types.Tx
	sizeDelegate       func() int
	txsBytesDelegate   func() int64
)

type mockMempool struct {
	checkTxFn    checkTxDelegate
	reapMaxTxsFn reapMaxTxsDelegate
	sizeFn       sizeDelegate
	txsBytesFn   txsBytesDelegate
}

func (m *mockMempool) CheckTx(tx types.Tx, cb func(abci.Response)) error {
	if m.checkTxFn != nil {
		return m.checkTxFn(tx, cb)
	}

	return nil
}

func (m *mockMempool) ReapMaxTxs(max int) []types.Tx {
	if m.reapMaxTxsFn != nil {
		return m.reapMaxTxsFn(max)
	}

	return nil
}

func (m *mockMempool) Size() int {
	if m.sizeFn != nil {
		return m.sizeFn()
	}

	return 0
}

func (m *mockMempool) TxsBytes() int64 {
	if m.txsBytesFn != nil {
		return m.txsBytesFn()
	}

	return 0
}

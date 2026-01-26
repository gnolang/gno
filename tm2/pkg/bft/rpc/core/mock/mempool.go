package mock

import (
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

type (
	CheckTxDelegate    func(tx types.Tx, cb func(abci.Response)) error
	ReapMaxTxsDelegate func(max int) types.Txs
	SizeDelegate       func() int
	TxsBytesDelegate   func() int64
)

type Mempool struct {
	CheckTxFn    CheckTxDelegate
	ReapMaxTxsFn ReapMaxTxsDelegate
	SizeFn       SizeDelegate
	TxsBytesFn   TxsBytesDelegate
}

func (m *Mempool) CheckTx(tx types.Tx, cb func(abci.Response)) error {
	if m.CheckTxFn != nil {
		return m.CheckTxFn(tx, cb)
	}

	return nil
}

func (m *Mempool) ReapMaxTxs(max int) types.Txs {
	if m.ReapMaxTxsFn != nil {
		return m.ReapMaxTxsFn(max)
	}

	return nil
}

func (m *Mempool) Size() int {
	if m.SizeFn != nil {
		return m.SizeFn()
	}

	return 0
}

func (m *Mempool) TxsBytes() int64 {
	if m.TxsBytesFn != nil {
		return m.TxsBytesFn()
	}

	return 0
}

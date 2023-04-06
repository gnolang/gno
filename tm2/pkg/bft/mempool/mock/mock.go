package mock

import (
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	mempl "github.com/gnolang/gno/tm2/pkg/bft/mempool"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/clist"
)

// Mempool is an empty implementation of a Mempool, useful for testing.
type Mempool struct{}

var _ mempl.Mempool = Mempool{}

func (Mempool) Lock()     {}
func (Mempool) Unlock()   {}
func (Mempool) Size() int { return 0 }
func (Mempool) CheckTx(_ types.Tx, _ func(abci.Response)) error {
	return nil
}

func (Mempool) CheckTxWithInfo(_ types.Tx, _ func(abci.Response),
	_ mempl.TxInfo,
) error {
	return nil
}
func (Mempool) ReapMaxBytesMaxGas(_, _ int64) types.Txs { return types.Txs{} }
func (Mempool) ReapMaxTxs(n int) types.Txs              { return types.Txs{} }
func (Mempool) Update(
	_ int64,
	_ types.Txs,
	_ []abci.ResponseDeliverTx,
	_ mempl.PreCheckFunc,
	_ int64,
) error {
	return nil
}
func (Mempool) Flush()                        {}
func (Mempool) FlushAppConn() error           { return nil }
func (Mempool) TxsAvailable() <-chan struct{} { return make(chan struct{}) }
func (Mempool) EnableTxsAvailable()           {}
func (Mempool) MaxTxBytes() int64             { return 0 }
func (Mempool) TxsBytes() int64               { return 0 }

func (Mempool) TxsFront() *clist.CElement    { return nil }
func (Mempool) TxsWaitChan() <-chan struct{} { return nil }

func (Mempool) InitWAL()  {}
func (Mempool) CloseWAL() {}

package mempool

import (
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

// Mempool is the minimal mempool interface the RPC handler needs
type Mempool interface {
	// CheckTx submits a transaction to the mempool.
	// If cb is non-nil, it is called with the CheckTx ABCI response
	CheckTx(tx types.Tx, cb func(abci.Response)) error

	// ReapMaxTxs returns up to max pending transactions from the mempool
	ReapMaxTxs(maxTxs int) types.Txs

	// Size returns the number of transactions currently in the mempool
	Size() int

	// TxsBytes returns the total size (in bytes) of all transactions in the mempool
	TxsBytes() int64
}

package mempool

import (
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

// Mempool defines the mempool interface.
//
// Updates to the mempool need to be synchronized with committing a block so
// apps can reset their transient state on Commit.
type Mempool interface {
	// CheckTx executes a new transaction against the application to determine
	// its validity and whether it should be added to the mempool.
	CheckTx(tx types.Tx, callback func(abci.Response)) error

	// CheckTxWithInfo performs the same operation as CheckTx, but with extra
	// meta data about the tx.
	// Currently this metadata is the peer who sent it, used to prevent the tx
	// from being gossiped back to them.
	CheckTxWithInfo(tx types.Tx, callback func(abci.Response), txInfo TxInfo) error

	// ReapMaxBytesMaxGas reaps transactions from the mempool up to maxDataBytes
	// bytes total with the condition that the total gasWanted must be less than
	// maxGas.
	// If both maxes are negative, there is no cap on the size of all returned
	// transactions (~ all available transactions).
	ReapMaxBytesMaxGas(maxDataBytes, maxGas int64) types.Txs

	// ReapMaxTxs reaps up to max transactions from the mempool.
	// If max is negative, there is no cap on the size of all returned
	// transactions (~ all available transactions).
	ReapMaxTxs(maxVal int) types.Txs

	// Update informs the mempool that the given txs were committed and can be discarded.
	// NOTE: this should be called *after* block is committed by consensus.
	// NOTE: unsafe; Lock/Unlock must be managed by caller
	Update(blockHeight int64, blockTxs types.Txs, deliverTxResponses []abci.ResponseDeliverTx, newPreFn PreCheckFunc, maxTxBytes int64) error

	// FlushAppConn flushes the mempool connection to ensure async reqResCb calls are
	// done. E.g. from CheckTx.
	FlushAppConn() error

	// Flush removes all transactions from the mempool and cache
	Flush()

	// TxsAvailable returns a channel which fires once for every height,
	// and only when transactions are available in the mempool.
	// NOTE: the returned channel may be nil if EnableTxsAvailable was not called.
	TxsAvailable() <-chan struct{}

	// EnableTxsAvailable initializes the TxsAvailable channel, ensuring it will
	// trigger once every height when transactions are available.
	EnableTxsAvailable()

	// Size returns the number of transactions in the mempool.
	Size() int

	// TxsBytes returns the total size of all txs in the mempool.
	TxsBytes() int64

	// Maximum allowable tx size.
	MaxTxBytes() int64
}

// --------------------------------------------------------------------------------

// PreCheckFunc is an optional filter executed before CheckTx and rejects
// transaction if false is returned. An example would be to ensure that a
// transaction doesn't exceeded the block size.
//
// NOTE: there is no PostCheckFunc, for otherwise a checktx transaction
// that passes in the app's checktx state would increment sequence etc,
// causing an unexpected signature error until the next block.
type PreCheckFunc func(types.Tx) error

// TxInfo are parameters that get passed when attempting to add a tx to the
// mempool.
type TxInfo struct {
	// We don't use p2p.ID here because it's too big. The gain is to store max 2
	// bytes with each tx to identify the sender rather than 20 bytes.
	SenderID uint16
}

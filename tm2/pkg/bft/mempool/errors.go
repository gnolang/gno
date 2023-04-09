package mempool

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/errors"
)

// ErrTxInCache is returned to the client if we saw tx earlier
var ErrTxInCache = errors.New("Tx already exists in cache")

// TxTooLargeError means the tx is too big to be sent in a message to other peers
type TxTooLargeError struct {
	max    int64
	actual int64
}

func (e TxTooLargeError) Error() string {
	return fmt.Sprintf("Tx too large. Max size is %d, but got %d", e.max, e.actual)
}

// MempoolIsFullError means Tendermint & an application can't handle that much load
type MempoolIsFullError struct {
	numTxs int
	maxTxs int

	txsBytes    int64
	maxTxsBytes int64
}

func (e MempoolIsFullError) Error() string {
	return fmt.Sprintf(
		"mempool is full: number of txs %d (max: %d), total txs bytes %d (max: %d)",
		e.numTxs, e.maxTxs,
		e.txsBytes, e.maxTxsBytes)
}

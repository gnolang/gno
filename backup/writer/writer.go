package writer

import "github.com/gnolang/tx-archive/types"

// Writer defines the backup writer interface
type Writer interface {
	// WriteTxData outputs the given TX data
	// to some kind of storage
	WriteTxData(*types.TxData) error
}

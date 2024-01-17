package core

import (
	"fmt"

	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
)

// Tx allows you to query the transaction results. `nil` could mean the
// transaction is in the mempool, invalidated, or was not sent in the first
// place
func Tx(_ *rpctypes.Context, hash []byte) (*ctypes.ResultTx, error) {
	// Get the result from storage, if any
	result, err := sm.LoadTxResult(stateDB, hash)
	if err != nil {
		return nil, fmt.Errorf("tx (%X) not found", hash)
	}

	// Return the response
	return &ctypes.ResultTx{
		Hash:     hash,
		Height:   result.Height,
		Index:    result.Index,
		TxResult: result.Response,
		Tx:       result.Tx,
	}, nil
}

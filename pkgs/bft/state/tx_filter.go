package state

import (
	mempl "github.com/gnolang/gno/pkgs/bft/mempool"
	"github.com/gnolang/gno/pkgs/bft/types"
)

// TxPreCheck returns a function to filter transactions before processing.
// The function does nothing yet.
func TxPreCheck(state State) mempl.PreCheckFunc {
	return func(types.Tx) error { return nil }
}

// TxPostCheck returns a function to filter transactions after processing.
// The function limits the gas wanted by a transaction to the block's maximum total gas.
func TxPostCheck(state State) mempl.PostCheckFunc {
	return mempl.PostCheckMaxGas(state.ConsensusParams.Block.MaxGas)
}

package state

import (
	mempl "github.com/gnolang/gno/tm2/pkg/bft/mempool"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

// TxPreCheck returns a function to filter transactions before processing.
// The function does nothing yet.
func TxPreCheck(state State) mempl.PreCheckFunc {
	return func(types.Tx) error { return nil }
}

package auth

import (
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// EndBlocker is called in the EndBlock(), it calcuates the minimum gas price
// for the next gas price
func EndBlocker(ctx sdk.Context, gk GasPriceKeeperI) {
	gk.UpdateGasPrice(ctx)
}

// InitChainer is called in the InitChain(), it set the initial gas price in the
// GasPriceKeeper store
// for the next gas price
func InitChainer(ctx sdk.Context, gk GasPriceKeeperI, gp std.GasPrice) {
	gk.SetGasPrice(ctx, gp)
}

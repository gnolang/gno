package bank

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/sdk"
)

// EndBlocker is called in the EndBlock(), it checks invariants
func EndBlocker(ctx sdk.Context, bk BankKeeperI) {
	bankk := bk.(BankKeeper)
	invariant := BalanceChangeInvariant(bankk)
	msg, stop := invariant(ctx)
	if stop {
		panic(fmt.Errorf("invariant broken: %s", msg))
	}
}

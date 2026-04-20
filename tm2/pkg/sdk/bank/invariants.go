package bank

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// maxInvariantGas bounds the gas an invariant may consume. Invariants
// iterate broad state (e.g. all accounts) so they need a generous
// budget, but an unbounded meter would let a pathological state grow
// unbounded invariant runtime and silently break determinism if the
// operator ever enables invariants in production.
const maxInvariantGas = 3_000_000_000

// RegisterInvariants registers the bank module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, acck auth.AccountKeeper) {
	ir.RegisterRoute(ModuleName, "nonnegative-outstanding",
		NonnegativeBalanceInvariant(acck))
}

// NonnegativeBalanceInvariant checks that all accounts in the application have non-negative balances
func NonnegativeBalanceInvariant(acck auth.AccountKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var msg string
		var count int

		// Install a bounded meter so account iteration through the
		// cache.Store gas layer actually enforces (production ctx
		// arrives with NewInfiniteGasMeter).
		ctx = ctx.WithGasMeter(store.NewGasMeter(maxInvariantGas))
		accts := acck.GetAllAccounts(ctx)
		for _, acc := range accts {
			coins := acc.GetCoins()
			if coins.IsAnyNegative() {
				count++
				msg += fmt.Sprintf("\t%s has a negative denomination of %s\n",
					acc.GetAddress().String(),
					coins.String())
			}
		}
		broken := count != 0

		return sdk.FormatInvariant(ModuleName, "nonnegative-outstanding",
			fmt.Sprintf("amount of negative accounts found %d\n%s", count, msg)), broken
	}
}

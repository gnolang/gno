package vm

import (
	"github.com/gnolang/gno/pkgs/sdk"
)

// RegisterInvariants registers the vm module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, vmk VMKeeper) {
	//ir.RegisterRoute(ModuleName, "nonnegative-outstanding",
	//	NonnegativeBalanceInvariant(acck))
}

/* TODO write new invariants for vm.
// NonnegativeBalanceInvariant checks that all accounts in the application have non-negative balances
func NonnegativeBalanceInvariant(acck auth.AccountKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var msg string
		var count int

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
*/

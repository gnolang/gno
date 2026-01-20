package bank

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const (
	// BalanceIncKey is the key to stores the balances increase of each denom
	balanceIncKey = "balanceIncrease"

	// BalanceDecKey is the key to stores the balances decrease of each denom
	balanceDecKey = "balanceDecrease"
)

// RegisterInvariants registers the bank module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, acck auth.AccountKeeper) {
	ir.RegisterRoute(ModuleName, "nonnegative-outstanding",
		NonnegativeBalanceInvariant(acck))
}

// NonnegativeBalanceInvariant checks that all accounts in the application have non-negative balances
// We don't currently check this invariant, as iterating through all accounts is costly.
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

// This invariant should be only checked in genesis block, not on other blocks.
// Since it iterate through all accounts. It is impractical in a production
// environment due to execution costs
func TotalSupplyInvariant(auth auth.AccountKeeperI, bank BankKeeperI) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		if ctx.BlockHeight() != 0 {
			return "", false
		}
		supply := std.Coins{}
		accts := auth.GetAllAccounts(ctx)
		for _, acc := range accts {
			coins := acc.GetCoins()
			supply = supply.Add(coins)
		}
		totalSupply := bank.GetParams(ctx).TotalSupply
		if !supply.IsEqual(totalSupply) {
			return sdk.FormatInvariant(
				ModuleName,
				"total supply",
				fmt.Sprintf("sum of accounts %v != total supply %v", supply, totalSupply),
			), true
		}

		return "", false
	}
}

// We check the changes in each block to ensure that the details of account balance changes
// tally to zero. This is checked every block and combined to use togateher with
// TotalSupplyInvariant checked in genesis block to ensure, total supply holds invariant
// for every blocks
func BalanceChangeInvariant(bank BankKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.Store(bank.key)
		balanceInc := std.Coins{}
		balanceDec := std.Coins{}
		if bz := store.Get(storeKey(balanceIncKey)); bz != nil {
			amino.MustUnmarshal(bz, &balanceInc)
		}

		if bz := store.Get(storeKey(balanceDecKey)); bz != nil {
			amino.MustUnmarshal(bz, &balanceDec)
		}

		if balanceInc.Sub(balanceDec) != nil {
			return sdk.FormatInvariant(
				ModuleName,
				"balance changes",
				fmt.Sprintf("sum of balance increase %v != sum of balance increase %v in a block", balanceInc, balanceDec),
			), true
		}

		return "", false
	}
}

// XXX: We need to add invariant checks when implementing mint and burn functions.

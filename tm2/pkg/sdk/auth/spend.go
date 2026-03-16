package auth

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// DeductSessionSpend checks and deducts spending from a session's spend limit.
// If SpendLimit is empty, any non-zero amount is rejected — the session cannot
// spend coins at all (useful when another signer pays gas, or for zero-send calls).
// Mutates da.SpendUsed in memory. Caller must persist via SetSessionAccount.
func DeductSessionSpend(da std.DelegatedAccount, amount std.Coins, blockTime int64) error {
	if amount.IsZero() {
		return nil
	}

	// No spend limit set — no spending allowed.
	if len(da.GetSpendLimit()) == 0 {
		return std.ErrSessionNotAllowed("session has no spend limit")
	}

	// Reset period if expired.
	if da.GetSpendPeriod() > 0 && blockTime >= da.GetSpendReset()+da.GetSpendPeriod() {
		da.SetSpendUsed(nil)
		da.SetSpendReset(blockTime)
	}

	// Check limit.
	newUsed := da.GetSpendUsed().Add(amount)
	if !da.GetSpendLimit().IsAllGTE(newUsed) {
		return std.ErrSessionNotAllowed("session spend limit exceeded")
	}

	// Deduct in memory.
	da.SetSpendUsed(newUsed)
	return nil
}

// SessionAccountSetter can persist session accounts after spend deduction.
type SessionAccountSetter interface {
	SetSessionAccount(ctx sdk.Context, master crypto.Address, acc std.Account)
}

// CheckAndDeductSessionSpend looks up the session for a signer from context
// and checks/deducts spending. Persists the updated session account to store.
// Returns nil if not a session tx or signer isn't using a session.
func CheckAndDeductSessionSpend(ctx sdk.Context, ak SessionAccountSetter, signerAddr crypto.Address, amount std.Coins) error {
	sa := ctx.Value(std.SessionAccountsContextKey{})
	if sa == nil {
		return nil
	}
	sessions, ok := sa.(map[crypto.Address]std.DelegatedAccount)
	if !ok {
		return nil
	}
	da, ok := sessions[signerAddr]
	if !ok {
		return nil
	}

	if err := DeductSessionSpend(da, amount, ctx.BlockTime().Unix()); err != nil {
		return err
	}

	// Persist the updated spend to store.
	ak.SetSessionAccount(ctx, signerAddr, da.(std.Account))
	return nil
}

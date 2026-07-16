package auth

import (
	"fmt"

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
		return std.ErrSessionNotAllowed(fmt.Sprintf(
			"session has no spend limit (attempted %s)", amount))
	}

	// Reset period if expired.
	if da.GetSpendPeriod() > 0 && blockTime >= da.GetSpendReset()+da.GetSpendPeriod() {
		da.SetSpendUsed(nil)
		da.SetSpendReset(blockTime)
	}

	// Check limit.
	newUsed := da.GetSpendUsed().Add(amount)
	if !da.GetSpendLimit().IsAllGTE(newUsed) {
		return std.ErrSessionNotAllowed(fmt.Sprintf(
			"session spend limit exceeded: attempted=%s, used=%s, limit=%s",
			amount, da.GetSpendUsed(), da.GetSpendLimit()))
	}

	// Deduct in memory.
	da.SetSpendUsed(newUsed)
	return nil
}

// CheckSessionSpend verifies that `amount` could be deducted from the
// session's current budget without exceeding SpendLimit — WITHOUT
// mutating da. Uses the same semantics as DeductSessionSpend for
// empty-limit rejection and period-reset, but returns the same error
// without touching state.
//
// Used by the ante's session pre-check to fail fast on obviously-
// over-limit session-signed txs before any gas fee is deducted,
// which prevents the mempool-gas-bleed attack where repeated over-
// limit submissions would otherwise charge gas per attempt.
func CheckSessionSpend(da std.DelegatedAccount, amount std.Coins, blockTime int64) error {
	if amount.IsZero() {
		return nil
	}
	if len(da.GetSpendLimit()) == 0 {
		return std.ErrSessionNotAllowed(fmt.Sprintf(
			"session has no spend limit (attempted %s)", amount))
	}

	// Compute the effective SpendUsed after any pending period reset.
	// Does NOT mutate da — reset is only "conceptual" here.
	spendUsed := da.GetSpendUsed()
	if da.GetSpendPeriod() > 0 && blockTime >= da.GetSpendReset()+da.GetSpendPeriod() {
		spendUsed = nil
	}

	newUsed := spendUsed.Add(amount)
	if !da.GetSpendLimit().IsAllGTE(newUsed) {
		return std.ErrSessionNotAllowed(fmt.Sprintf(
			"session spend limit would be exceeded: attempted=%s, used=%s, limit=%s",
			amount, spendUsed, da.GetSpendLimit()))
	}
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
	//
	// The `da.(std.Account)` assertion is guaranteed safe: DelegatedAccount
	// embeds Account (see std/account.go DelegatedAccount interface), so any
	// concrete type satisfying DelegatedAccount also satisfies Account.
	// The assertion is here only because SetSessionAccount's signature takes
	// std.Account — Go can't auto-widen through the interface map value.
	ak.SetSessionAccount(ctx, signerAddr, da.(std.Account))
	return nil
}

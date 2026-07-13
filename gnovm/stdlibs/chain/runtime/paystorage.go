package runtime

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
	"github.com/gnolang/gno/tm2/pkg/overflow"
)

func PayStorage(m *gno.Machine, maxDeposit int64) {
	// 1. Validate maxDeposit > 0
	if maxDeposit <= 0 {
		m.Panic(typedString("PayStorage: maxDeposit must be positive"))
		return
	}

	// 2. Rule 1 — Only realms can call PayStorage
	_, currentPkgPath := execctx.CurrentRealm(m)
	if !gno.IsRealmPath(currentPkgPath) {
		m.Panic(typedString("PayStorage: caller is not a realm"))
		return
	}

	// 3. Rule 2 — Function creator must match payer
	callerFrame := m.MustPeekCallFrame(2)
	if callerFrame.Func == nil {
		m.Panic(typedString("PayStorage: cannot be called from anonymous function"))
		return
	}
	if callerFrame.Func.PkgPath != currentPkgPath {
		m.Panic(typedString("PayStorage: function creator realm does not match paying realm"))
		return
	}

	// 4. Get context (returned by value; PayStorageInfo is a shared pointer so mutations propagate)
	ctx := execctx.GetContext(m)

	// 5. Check PayStorage not already called
	if ctx.PayStorageInfo == nil {
		m.Panic(typedString("PayStorage: feature not available in this context"))
		return
	}
	// PayStorage only applies to 0-fee sponsored txs — mirror of PayGas. In a
	// normal fee-paying tx the signer pays their own storage deposit, so
	// PayStorage is a no-op here; otherwise a realm exposing an unconditional
	// PayStorage would silently pay the deposit for any caller's writes on an
	// ordinary tx (a self-drain), contradicting the "only in sponsored txs" rule.
	if !ctx.PayStorageInfo.Eligible {
		return
	}
	if ctx.PayStorageInfo.MaxDeposit > 0 {
		m.Panic(typedString("PayStorage: already called in this transaction"))
		return
	}

	// 7. Check gas price denom is set (needed for balance lookup)
	if ctx.GasPrice.Price.Denom == "" {
		m.Panic(typedString("PayStorage: gas price not set"))
		return
	}

	// 8. Check realm balance. Fold an existing PayGas commitment into the required
	// balance ONLY when this SAME realm made it (then one realm covers storage +
	// gas together). A different realm sponsoring gas pays from its own balance
	// (settlement charges each realm separately), so its commitment must not
	// inflate this realm's storage-affordability pre-check. (Two-realm sponsorship
	// is supported: realm A may pay gas while realm B pays storage.)
	realmAddr := gno.DerivePkgBech32Addr(currentPkgPath)
	coins := ctx.Banker.GetCoins(realmAddr)
	ugnotBalance := int64(0)
	for _, c := range coins {
		if c.Denom == ctx.GasPrice.Price.Denom {
			ugnotBalance = c.Amount
			break
		}
	}
	totalRequired := maxDeposit
	if ctx.PayGasInfo != nil && ctx.PayGasInfo.MaxFee > 0 &&
		ctx.PayGasInfo.RealmPkgPath == currentPkgPath {
		sum, ok := overflow.Add(maxDeposit, ctx.PayGasInfo.MaxFee)
		if !ok {
			m.Panic(typedString("PayStorage: total commitment overflow"))
			return
		}
		totalRequired = sum
	}
	if ugnotBalance < totalRequired {
		m.Panic(typedString("PayStorage: insufficient realm balance for storage + gas"))
		return
	}

	// 9. Set PayStorageInfo
	ctx.PayStorageInfo.RealmPkgPath = currentPkgPath
	ctx.PayStorageInfo.RealmAddr = gno.DerivePkgCryptoAddr(currentPkgPath)
	ctx.PayStorageInfo.MaxDeposit = maxDeposit
}

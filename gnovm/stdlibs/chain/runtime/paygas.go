package runtime

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
	"github.com/gnolang/gno/tm2/pkg/overflow"
)

func PayGas(m *gno.Machine, maxFee int64) {
	// 1. Validate maxFee > 0
	if maxFee <= 0 {
		m.Panic(typedString("PayGas: maxFee must be positive"))
		return
	}

	// 2. Rule 1 — Only realms can call PayGas
	_, currentPkgPath := execctx.CurrentRealm(m)
	if !gno.IsRealmPath(currentPkgPath) {
		m.Panic(typedString("PayGas: caller is not a realm"))
		return
	}

	// 3. Rule 2 — Function creator must match payer
	callerFrame := m.MustPeekCallFrame(2)
	if callerFrame.Func == nil {
		m.Panic(typedString("PayGas: cannot be called from anonymous function"))
		return
	}
	if callerFrame.Func.PkgPath != currentPkgPath {
		m.Panic(typedString("PayGas: function creator realm does not match paying realm"))
		return
	}

	// 4. Get context (returned by value; PayGasInfo is a shared pointer so mutations propagate)
	ctx := execctx.GetContext(m)

	// 5. Rule 3 — Only once per tx
	if ctx.PayGasInfo == nil {
		m.Panic(typedString("PayGas: feature not available in this context"))
		return
	}
	// PayGas only applies to 0-fee sponsored txs. In a normal fee-paying tx the
	// signer already pays for gas, so calling PayGas is a no-op here — this
	// prevents charging both the signer (ante fee) and the realm (settlement),
	// and prevents shrinking the user's gas limit below their GasWanted.
	if !ctx.PayGasInfo.Eligible {
		return
	}
	if ctx.PayGasInfo.MaxFee > 0 {
		m.Panic(typedString("PayGas: already called in this transaction"))
		return
	}

	// 6. Check gas price is set
	if ctx.GasPrice.Gas <= 0 || ctx.GasPrice.Price.Amount <= 0 || ctx.GasPrice.Price.Denom == "" {
		m.Panic(typedString("PayGas: gas price not set"))
		return
	}

	// 7. Check realm balance
	realmAddr := gno.DerivePkgBech32Addr(currentPkgPath)
	coins := ctx.Banker.GetCoins(realmAddr)
	ugnotBalance := int64(0)
	for _, c := range coins {
		if c.Denom == ctx.GasPrice.Price.Denom {
			ugnotBalance = c.Amount
			break
		}
	}
	// Fold an existing PayStorage commitment into the required balance ONLY when
	// this SAME realm made it — then one realm must cover gas + storage together.
	// A different realm sponsoring storage pays from its own balance (settlement
	// charges each realm separately), so its commitment must not inflate this
	// realm's gas-affordability pre-check. (Two-realm sponsorship is supported.)
	// The storage deposit is denominated in StorageDepositDenom, which need not
	// be the gas-price denom this balance was read in; only fold when they match,
	// since adding amounts of two different tokens is meaningless.
	totalRequired := maxFee
	if ctx.PayStorageInfo != nil && ctx.PayStorageInfo.MaxDeposit > 0 &&
		ctx.PayStorageInfo.RealmPkgPath == currentPkgPath &&
		ctx.StorageDepositDenom == ctx.GasPrice.Price.Denom {
		// Fold only the commitment still OUTSTANDING. On the per-message path
		// (SponsorStorage=false) earlier messages already debited the realm and
		// recorded it in SpentDeposit, and the balance read above reflects that
		// — so folding the full MaxDeposit would demand it twice and reject an
		// otherwise affordable tx. This mirrors how the keeper derives the
		// remaining budget (MaxDeposit - SpentDeposit).
		outstanding := ctx.PayStorageInfo.MaxDeposit - ctx.PayStorageInfo.SpentDeposit
		if outstanding > 0 {
			sum, ok := overflow.Add(maxFee, outstanding)
			if !ok {
				m.Panic(typedString("PayGas: total commitment overflow"))
				return
			}
			totalRequired = sum
		}
	}
	if ugnotBalance < totalRequired {
		m.Panic(typedString("PayGas: insufficient realm balance"))
		return
	}

	// 8. Derive gas limit with overflow-safe math:
	// derivedLimit = maxFee * Gas / Price.Amount
	product, ok := overflow.Mul(maxFee, ctx.GasPrice.Gas)
	if !ok {
		m.Panic(typedString("PayGas: maxFee * gas price overflows"))
		return
	}
	derivedLimit := product / ctx.GasPrice.Price.Amount

	// Never raise the gas limit above the credit window the ante handler granted
	// for this tx; PayGas may only tighten it to what maxFee affords. This keeps a
	// single sponsored tx bounded by MaxGasCreditPerTx (and thus by Block.MaxGas),
	// so a large maxFee cannot let one tx monopolize a whole block's compute.
	//
	// An INFINITE meter reports Limit() == 0 (see infiniteGasMeter.Limit). That
	// sentinel means "no cap", not "zero budget" — the ante installs one for
	// source-gas hardfork replay, where the source chain's outcome must be
	// reproduced regardless of current metering.
	//
	// Both the clamp and the budget check below must be skipped for it. Clamping
	// would drive every derived limit to 0; and an infinite meter still
	// accumulates GasConsumed(), so enforcing the budget against it could refuse
	// to replay a historically successful sponsored tx purely because today's gas
	// costs or gas price differ — which would stop a hardfork node from booting.
	// The commitment itself is still recorded below, so settlement is unchanged
	// (SetLimit is a no-op on an infinite meter).
	unmetered := m.GasMeter.Limit() <= 0

	if curLimit := m.GasMeter.Limit(); !unmetered && derivedLimit > curLimit {
		derivedLimit = curLimit
	}

	// 9. Check budget not already exceeded
	if !unmetered {
		consumed := m.GasMeter.GasConsumed()
		if derivedLimit < consumed {
			m.Panic(typedString("PayGas: maxFee budget exceeded at current gas price"))
			return
		}

		// 10. Update gas meter limit
		m.GasMeter.SetLimit(derivedLimit)
	}

	// 11. Set PayGasInfo — shared pointer, visible to baseapp via SDK context
	ctx.PayGasInfo.RealmPkgPath = currentPkgPath
	ctx.PayGasInfo.RealmAddr = gno.DerivePkgCryptoAddr(currentPkgPath)
	ctx.PayGasInfo.MaxFee = maxFee
}

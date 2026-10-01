package runtime

// PayGas commits rlm to pay this transaction's gas, up to maxFee ugnot.
//
// rlm is the paying realm's capability, as for banker.NewBanker: pass the cur
// of your crossing function (or a value threaded down from it). It must be the
// current realm, and a top-level realm rather than a sub-realm token. It comes
// last because a realm-typed first parameter would make PayGas a crossing
// function.
//
// It only takes effect in a 0-fee sponsored transaction, and is a no-op
// otherwise. There it may be called once per transaction, and it covers the
// gas of the whole transaction, other messages included. The gas limit becomes
// what maxFee buys at the current gas price, capped by the credit window; the
// realm is charged for the gas actually used, and only if the transaction
// succeeds.
func PayGas(maxFee int64, rlm realm) {
	if !rlm.IsCurrent() {
		panic("PayGas: rlm must be the current realm")
	}
	payGas(rlm.PkgPath(), maxFee)
}

func payGas(pkgPath string, maxFee int64)

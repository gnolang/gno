// PKGPATH: gno.land/r/totalcoindenom
package totalcoindenom

import (
	"chain/banker"
)

// The TotalCoin half of zrealm_banker_getcoin_denom.gno, and separate from it
// because this branch implements TotalCoin at all — it was
// panic("not yet implemented") — so what is pinned here is that a realm can reach
// a working, denom-rejecting supply read through the interpreter.
//
// The rejection is not attributable to TotalCoin's own check: the table is
// non-empty, so AmountOf validates inside the sum. That check matters only for an
// empty table, where the loop never runs, and is pinned in Go by
// TestTotalCoinRejectsAMalformedDenomWithNothingHeld.
func main(cur realm) {
	bnk := banker.NewReadonlyBanker()

	// A well-formed denom nobody holds has zero supply, not an error. Asserted so
	// this cannot pass with a TotalCoin that rejects everything.
	println(bnk.TotalCoin("ugnot"))

	bnk.TotalCoin("UPPERCASE")
}

// Output:
// 0

// Error:
// invalid denom: UPPERCASE

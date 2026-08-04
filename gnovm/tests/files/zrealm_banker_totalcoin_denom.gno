// PKGPATH: gno.land/r/totalcoindenom
package totalcoindenom

import (
	"chain/banker"
)

// The TotalCoin half of zrealm_banker_getcoin_denom.gno. Both are needed: the two
// bankers behind the natives validate the denom in separate functions, so removing
// either validation leaves the other's filetest passing. This branch implements
// TotalCoin — it was panic("not yet implemented") — and the supply ADR claims that
// gno test and production agree on it, which is what this pins.
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
// invalid denom: invalid denom: UPPERCASE

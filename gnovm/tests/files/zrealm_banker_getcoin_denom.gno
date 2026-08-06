// PKGPATH: gno.land/r/getcoindenom
package getcoindenom

import (
	"chain/banker"
)

// GetCoin must reject a malformed denom under gno test as it does on chain, or a
// realm whose test suite passed would panic in production. This pins the whole
// path — realm to .gno wrapper to native to banker — which the Go-level tests do
// not exercise; it does not pin any one guard, and TestBanker deliberately has
// none, because Coins.AmountOf validates before it reads. What breaks this is
// GetCoin reaching a balance without going through Coins, which is exactly what
// the chain side does: SDKBanker.GetCoin therefore checks explicitly, pinned by
// TestSDKBankerGetCoinValidatesDenom.
func main(cur realm) {
	bnk := banker.NewReadonlyBanker()
	addr := cur.Address()

	// A well-formed denom nobody holds is zero, not an error. Asserted so this
	// cannot pass with a GetCoin that rejects everything.
	println(bnk.GetCoin(addr, "ugnot"))

	bnk.GetCoin(addr, "UPPERCASE")
}

// Output:
// 0

// Error:
// invalid denom: UPPERCASE

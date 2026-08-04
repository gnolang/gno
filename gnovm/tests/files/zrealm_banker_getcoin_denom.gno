// PKGPATH: gno.land/r/getcoindenom
package getcoindenom

import (
	"chain/banker"
)

// GetCoin takes a realm-supplied string straight to a store key, so both bankers
// behind the native — SDKBanker on chain, TestBanker under gno test — validate the
// denom and panic. They have to agree: a realm whose gno test accepted a malformed
// denom would panic in production instead. The chain half is pinned by
// TestSDKBankerGetCoinValidatesDenom; this is the gno test half.
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
// invalid denom: invalid denom: UPPERCASE

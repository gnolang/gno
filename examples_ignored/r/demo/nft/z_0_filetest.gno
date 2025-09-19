// PKGPATH: gno.land/r/demo/nft_test
package nft_test

import (
	"gno.land/p/demo/testutils"
	"gno.land/r/demo/nft"
)

func main() {
	addr1 := testutils.TestAddress("addr1")
	// addr2 := testutils.TestAddress("addr2")
	grc721 := nft.GetToken()
	tid := grc721.Mint(addr1, "NFT#1")
	println(grc721.OwnerOf(tid))
	println(addr1)
}

// Output:
// g1v9jxgu33ta047h6lta047h6lta047h6l43dqc5
// g1v9jxgu33ta047h6lta047h6lta047h6l43dqc5

// Realm:
// finalizerealm["gno.land/r/demo/nft_test"]

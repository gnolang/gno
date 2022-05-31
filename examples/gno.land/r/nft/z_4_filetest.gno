// PKGPATH: gno.land/r/nft_test
package nft_test

import (
	"std"

	"gno.land/p/testutils"
	"gno.land/r/nft"
)

func main() {
	caller := std.GetCallerAt(1)
	addr1 := testutils.TestAddress("addr1")
	addr2 := testutils.TestAddress("addr2")
	grc721 := nft.GetGRC721()
	tid := grc721.Mint(caller, "NFT#1")
	println(grc721.OwnerOf(tid))
	println(addr1)
	grc721.Approve(caller, tid) // approve self.
	grc721.TransferFrom(caller, addr1, tid)
	grc721.Approve("", tid) // approve addr1.
	grc721.TransferFrom(addr1, addr2, tid)
}

// Error:
// unauthorized

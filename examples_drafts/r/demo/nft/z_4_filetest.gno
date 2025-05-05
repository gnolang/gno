// PKGPATH: gno.land/r/demo/nft_test
package nft_test

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/nft"
)

func main() {
	caller := std.CallerAt(1)
	addr1 := testutils.TestAddress("addr1")
	addr2 := testutils.TestAddress("addr2")
	grc721 := nft.GetToken()
	tid := grc721.Mint(caller, "NFT#1")
	println(grc721.OwnerOf(tid))
	println(addr1)
	grc721.Approve(caller, tid) // approve self.
	grc721.TransferFrom(caller, addr1, tid)
	grc721.Approve("", tid) // approve addr1.
	grc721.TransferFrom(addr1, addr2, tid)
}

// Error:
// gno.land/r/demo/nft_test/z_4.gno:12:12: name CallerAt not declared

// TypeCheckError:
// gno.land/r/demo/nft_test/z_4.gno:12:16: undefined: std.CallerAt

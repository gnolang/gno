package foo721

import (
	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/ufmt/v0"
)

var (
	admin address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh" // govdao t1 multisig
	foo   *grc721.BasicNFT
)

func init(cur realm) {
	foo = grc721.NewBasicNFT(0, cur, "FooNFT", "FNFT")
	mintNNFT(admin, 10)                                     // @administrator (10)
	mintNNFT("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh", 5) // govdao t1 multisig (5)
}

func mintNNFT(owner address, n int64) {
	count := foo.TokenCount()
	for i := count; i < count+n; i++ {
		tid := grc721.TokenID(ufmt.Sprintf("%d", i))
		foo.Mint(owner, tid)
	}
}

// Getters

func BalanceOf(user address) int64 {
	balance, err := foo.BalanceOf(user)
	if err != nil {
		panic(err)
	}

	return balance
}

func OwnerOf(tid grc721.TokenID) address {
	owner, err := foo.OwnerOf(tid)
	if err != nil {
		panic(err)
	}

	return owner
}

func IsApprovedForAll(owner, user address) bool {
	return foo.IsApprovedForAll(owner, user)
}

func GetApproved(tid grc721.TokenID) address {
	addr, err := foo.GetApproved(tid)
	if err != nil {
		panic(err)
	}

	return addr
}

// Setters

// The Approve / SetApprovalForAll / TransferFrom wrappers are the
// owning realm's trusted boundary: they derive `caller` from
// `cur.Previous().Address()` (unforgeable per-frame identity) and
// hand the address-typed parameter to the concrete writer methods on
// the unexported `foo *basicNFT`. The basicNFT writer methods take
// `caller address` only — no rlm crosses the call, so even if the
// underlying type were ever exposed as an interface, no cur could
// leak through. See `p/demo/tokens/grc721/igrc721.gno` for the
// Reader/Writer split doctrine.
func Approve(cur realm, user address, tid grc721.TokenID) {
	caller := cur.Previous().Address()
	err := foo.Approve(caller, user, tid)
	if err != nil {
		panic(err)
	}
}

func SetApprovalForAll(cur realm, user address, approved bool) {
	caller := cur.Previous().Address()
	err := foo.SetApprovalForAll(caller, user, approved)
	if err != nil {
		panic(err)
	}
}

func TransferFrom(cur realm, from, to address, tid grc721.TokenID) {
	caller := cur.Previous().Address()
	err := foo.TransferFrom(caller, from, to, tid)
	if err != nil {
		panic(err)
	}
}

// Admin

func Mint(cur realm, to address, tid grc721.TokenID) {
	caller := cur.Previous().Address()
	assertIsAdmin(caller)
	err := foo.Mint(to, tid)
	if err != nil {
		panic(err)
	}
}

func Burn(cur realm, tid grc721.TokenID) {
	caller := cur.Previous().Address()
	assertIsAdmin(caller)
	err := foo.Burn(tid)
	if err != nil {
		panic(err)
	}
}

// Render

func Render(path string) string {
	switch {
	case path == "":
		return foo.RenderHome()
	default:
		return "404\n"
	}
}

// Util

func assertIsAdmin(address_XXX address) {
	if address_XXX != admin {
		panic("restricted access")
	}
}

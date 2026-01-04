package foo721

import (
	"chain/runtime"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/ufmt"
)

var (
	admin address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"
	foo           = grc721.NewBasicNFT("FooNFT", "FNFT")
)

func init() {
	mintNNFT(admin, 10)                                     // @administrator (10)
	mintNNFT("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm", 5) // @hariom (5)
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

func Approve(_ realm, user address, tid grc721.TokenID) {
	err := foo.Approve(user, tid)
	if err != nil {
		panic(err)
	}
}

func SetApprovalForAll(_ realm, user address, approved bool) {
	err := foo.SetApprovalForAll(user, approved)
	if err != nil {
		panic(err)
	}
}

func TransferFrom(_ realm, from, to address, tid grc721.TokenID) {
	err := foo.TransferFrom(from, to, tid)
	if err != nil {
		panic(err)
	}
}

// Admin

func Mint(_ realm, to address, tid grc721.TokenID) {
	caller := runtime.PreviousRealm().Address()
	assertIsAdmin(caller)
	err := foo.Mint(to, tid)
	if err != nil {
		panic(err)
	}
}

func Burn(_ realm, tid grc721.TokenID) {
	caller := runtime.PreviousRealm().Address()
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

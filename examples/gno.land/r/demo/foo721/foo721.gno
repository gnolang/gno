package foo721

import (
	"std"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/users"
)

var (
	admin std.Address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"
	foo               = grc721.NewBasicNFT("FooNFT", "FNFT")
)

func init() {
	mintNNFT(admin, 10)                                     // @administrator (10)
	mintNNFT("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm", 5) // @hariom (5)
}

func mintNNFT(owner std.Address, n uint64) {
	count := foo.TokenCount()
	for i := count; i < count+n; i++ {
		tid := grc721.TokenID(ufmt.Sprintf("%d", i))
		foo.Mint(owner, tid)
	}
}

// Getters

func BalanceOf(user users.AddressOrName) uint64 {
	balance, err := foo.BalanceOf(user.Resolve())
	if err != nil {
		panic(err)
	}

	return balance
}

func OwnerOf(tid grc721.TokenID) std.Address {
	owner, err := foo.OwnerOf(tid)
	if err != nil {
		panic(err)
	}

	return owner
}

func IsApprovedForAll(owner, user users.AddressOrName) bool {
	return foo.IsApprovedForAll(owner.Resolve(), user.Resolve())
}

func GetApproved(tid grc721.TokenID) std.Address {
	addr, err := foo.GetApproved(tid)
	if err != nil {
		panic(err)
	}

	return addr
}

// Setters

func Approve(user users.AddressOrName, tid grc721.TokenID) {
	err := foo.Approve(user.Resolve(), tid)
	if err != nil {
		panic(err)
	}
}

func SetApprovalForAll(user users.AddressOrName, approved bool) {
	err := foo.SetApprovalForAll(user.Resolve(), approved)
	if err != nil {
		panic(err)
	}
}

func TransferFrom(from, to users.AddressOrName, tid grc721.TokenID) {
	err := foo.TransferFrom(from.Resolve(), to.Resolve(), tid)
	if err != nil {
		panic(err)
	}
}

// Admin

func Mint(to users.AddressOrName, tid grc721.TokenID) {
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller)
	err := foo.Mint(to.Resolve(), tid)
	if err != nil {
		panic(err)
	}
}

func Burn(tid grc721.TokenID) {
	caller := std.PrevRealm().Addr()
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

func assertIsAdmin(address std.Address) {
	if address != admin {
		panic("restricted access")
	}
}

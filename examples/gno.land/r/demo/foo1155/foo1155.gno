package foo1155

import (
	"std"

	"gno.land/p/demo/grc/grc1155"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/users"
)

var (
	dummyURI             = "ipfs://xyz"
	admin    std.Address = "g10x5phu0k6p64cwrhfpsc8tk43st9kug6wft530"
	foo                  = grc1155.NewBasicGRC1155Token(dummyURI)
)

func init() {
	mintGRC1155Token(admin) // @administrator (10)
}

func mintGRC1155Token(owner std.Address) {
	for i := 1; i <= 10; i++ {
		tid := grc1155.TokenID(ufmt.Sprintf("%d", i))
		foo.SafeMint(owner, tid, 100)
	}
}

// Getters

func BalanceOf(user users.AddressOrName, tid grc1155.TokenID) uint64 {
	balance, err := foo.BalanceOf(user.Resolve(), tid)
	if err != nil {
		panic(err)
	}

	return balance
}

func BalanceOfBatch(users []users.AddressOrName, batch []grc1155.TokenID) []uint64 {
	var usersResolved []std.Address

	for i := 0; i < len(users); i++ {
		usersResolved[i] = users[i].Resolve()
	}
	balanceBatch, err := foo.BalanceOfBatch(usersResolved, batch)
	if err != nil {
		panic(err)
	}

	return balanceBatch
}

func IsApprovedForAll(owner, user users.AddressOrName) bool {
	return foo.IsApprovedForAll(owner.Resolve(), user.Resolve())
}

// Setters

func SetApprovalForAll(user users.AddressOrName, approved bool) {
	err := foo.SetApprovalForAll(user.Resolve(), approved)
	if err != nil {
		panic(err)
	}
}

func TransferFrom(from, to users.AddressOrName, tid grc1155.TokenID, amount uint64) {
	err := foo.SafeTransferFrom(from.Resolve(), to.Resolve(), tid, amount)
	if err != nil {
		panic(err)
	}
}

func BatchTransferFrom(from, to users.AddressOrName, batch []grc1155.TokenID, amounts []uint64) {
	err := foo.SafeBatchTransferFrom(from.Resolve(), to.Resolve(), batch, amounts)
	if err != nil {
		panic(err)
	}
}

// Admin

func Mint(to users.AddressOrName, tid grc1155.TokenID, amount uint64) {
	caller := std.OrigCaller()
	assertIsAdmin(caller)
	err := foo.SafeMint(to.Resolve(), tid, amount)
	if err != nil {
		panic(err)
	}
}

func MintBatch(to users.AddressOrName, batch []grc1155.TokenID, amounts []uint64) {
	caller := std.OrigCaller()
	assertIsAdmin(caller)
	err := foo.SafeBatchMint(to.Resolve(), batch, amounts)
	if err != nil {
		panic(err)
	}
}

func Burn(from users.AddressOrName, tid grc1155.TokenID, amount uint64) {
	caller := std.OrigCaller()
	assertIsAdmin(caller)
	err := foo.Burn(from.Resolve(), tid, amount)
	if err != nil {
		panic(err)
	}
}

func BurnBatch(from users.AddressOrName, batch []grc1155.TokenID, amounts []uint64) {
	caller := std.OrigCaller()
	assertIsAdmin(caller)
	err := foo.BatchBurn(from.Resolve(), batch, amounts)
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

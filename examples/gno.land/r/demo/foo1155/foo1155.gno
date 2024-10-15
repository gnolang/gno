package foo1155

import (
	"std"

	"gno.land/p/demo/grc/grc1155"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/users"

	pusers "gno.land/p/demo/users"
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

func BalanceOf(user pusers.AddressOrName, tid grc1155.TokenID) uint64 {
	balance, err := foo.BalanceOf(users.Resolve(user), tid)
	if err != nil {
		panic(err)
	}

	return balance
}

func BalanceOfBatch(ul []pusers.AddressOrName, batch []grc1155.TokenID) []uint64 {
	var usersResolved []std.Address

	for i := 0; i < len(ul); i++ {
		usersResolved[i] = users.Resolve(ul[i])
	}
	balanceBatch, err := foo.BalanceOfBatch(usersResolved, batch)
	if err != nil {
		panic(err)
	}

	return balanceBatch
}

func IsApprovedForAll(owner, user pusers.AddressOrName) bool {
	return foo.IsApprovedForAll(users.Resolve(owner), users.Resolve(user))
}

// Setters

func SetApprovalForAll(user pusers.AddressOrName, approved bool) {
	err := foo.SetApprovalForAll(users.Resolve(user), approved)
	if err != nil {
		panic(err)
	}
}

func TransferFrom(from, to pusers.AddressOrName, tid grc1155.TokenID, amount uint64) {
	err := foo.SafeTransferFrom(users.Resolve(from), users.Resolve(to), tid, amount)
	if err != nil {
		panic(err)
	}
}

func BatchTransferFrom(from, to pusers.AddressOrName, batch []grc1155.TokenID, amounts []uint64) {
	err := foo.SafeBatchTransferFrom(users.Resolve(from), users.Resolve(to), batch, amounts)
	if err != nil {
		panic(err)
	}
}

// Admin

func Mint(to pusers.AddressOrName, tid grc1155.TokenID, amount uint64) {
	caller := std.GetOrigCaller()
	assertIsAdmin(caller)
	err := foo.SafeMint(users.Resolve(to), tid, amount)
	if err != nil {
		panic(err)
	}
}

func MintBatch(to pusers.AddressOrName, batch []grc1155.TokenID, amounts []uint64) {
	caller := std.GetOrigCaller()
	assertIsAdmin(caller)
	err := foo.SafeBatchMint(users.Resolve(to), batch, amounts)
	if err != nil {
		panic(err)
	}
}

func Burn(from pusers.AddressOrName, tid grc1155.TokenID, amount uint64) {
	caller := std.GetOrigCaller()
	assertIsAdmin(caller)
	err := foo.Burn(users.Resolve(from), tid, amount)
	if err != nil {
		panic(err)
	}
}

func BurnBatch(from pusers.AddressOrName, batch []grc1155.TokenID, amounts []uint64) {
	caller := std.GetOrigCaller()
	assertIsAdmin(caller)
	err := foo.BatchBurn(users.Resolve(from), batch, amounts)
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

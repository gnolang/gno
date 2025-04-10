package foo1155

import (
	"chain"
	"chain/runtime"

	"gno.land/p/demo/grc/grc1155"
	"gno.land/p/demo/ufmt"
)

var (
	dummyURI               = "ipfs://xyz"
	admin    chain.Address = "g10x5phu0k6p64cwrhfpsc8tk43st9kug6wft530"
	foo                    = grc1155.NewBasicGRC1155Token(dummyURI)
)

func init() {
	mintGRC1155Token(admin) // @administrator (10)
}

func mintGRC1155Token(owner chain.Address) {
	for i := 1; i <= 10; i++ {
		tid := grc1155.TokenID(ufmt.Sprintf("%d", i))
		foo.SafeMint(owner, tid, 100)
	}
}

// Getters

func BalanceOf(user chain.Address, tid grc1155.TokenID) uint64 {
	balance, err := foo.BalanceOf(user, tid)
	if err != nil {
		panic(err)
	}

	return balance
}

func BalanceOfBatch(ul []chain.Address, batch []grc1155.TokenID) []uint64 {
	balanceBatch, err := foo.BalanceOfBatch(ul, batch)
	if err != nil {
		panic(err)
	}

	return balanceBatch
}

func IsApprovedForAll(owner, user chain.Address) bool {
	return foo.IsApprovedForAll(owner, user)
}

// Setters

func SetApprovalForAll(user chain.Address, approved bool) {
	err := foo.SetApprovalForAll(user, approved)
	if err != nil {
		panic(err)
	}
}

func TransferFrom(from, to chain.Address, tid grc1155.TokenID, amount uint64) {
	err := foo.SafeTransferFrom(from, to, tid, amount)
	if err != nil {
		panic(err)
	}
}

func BatchTransferFrom(from, to chain.Address, batch []grc1155.TokenID, amounts []uint64) {
	err := foo.SafeBatchTransferFrom(from, to, batch, amounts)
	if err != nil {
		panic(err)
	}
}

// Admin

func Mint(to chain.Address, tid grc1155.TokenID, amount uint64) {
	caller := runtime.OriginCaller()
	assertIsAdmin(caller)
	err := foo.SafeMint(to, tid, amount)
	if err != nil {
		panic(err)
	}
}

func MintBatch(to chain.Address, batch []grc1155.TokenID, amounts []uint64) {
	caller := runtime.OriginCaller()
	assertIsAdmin(caller)
	err := foo.SafeBatchMint(to, batch, amounts)
	if err != nil {
		panic(err)
	}
}

func Burn(from chain.Address, tid grc1155.TokenID, amount uint64) {
	caller := runtime.OriginCaller()
	assertIsAdmin(caller)
	err := foo.Burn(from, tid, amount)
	if err != nil {
		panic(err)
	}
}

func BurnBatch(from chain.Address, batch []grc1155.TokenID, amounts []uint64) {
	caller := runtime.OriginCaller()
	assertIsAdmin(caller)
	err := foo.BatchBurn(from, batch, amounts)
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

func assertIsAdmin(address chain.Address) {
	if address != admin {
		panic("restricted access")
	}
}

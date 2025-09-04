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

func BalanceOf(user chain.Address, tid grc1155.TokenID) int64 {
	balance, err := foo.BalanceOf(user, tid)
	if err != nil {
		panic(err)
	}

	return balance
}

func BalanceOfBatch(ul []chain.Address, batch []grc1155.TokenID) []int64 {
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

func SetApprovalForAll(_ realm, user chain.Address, approved bool) {
	err := foo.SetApprovalForAll(user, approved)
	if err != nil {
		panic(err)
	}
}

func TransferFrom(_ realm, from, to chain.Address, tid grc1155.TokenID, amount int64) {
	err := foo.SafeTransferFrom(from, to, tid, amount)
	if err != nil {
		panic(err)
	}
}

func BatchTransferFrom(_ realm, from, to chain.Address, batch []grc1155.TokenID, amounts []int64) {
	err := foo.SafeBatchTransferFrom(from, to, batch, amounts)
	if err != nil {
		panic(err)
	}
}

// Admin

func Mint(_ realm, to chain.Address, tid grc1155.TokenID, amount int64) {
	caller := runtime.OriginCaller()
	assertIsAdmin(caller)
	err := foo.SafeMint(to, tid, amount)
	if err != nil {
		panic(err)
	}
}

func MintBatch(_ realm, to chain.Address, batch []grc1155.TokenID, amounts []int64) {
	caller := runtime.OriginCaller()
	assertIsAdmin(caller)
	err := foo.SafeBatchMint(to, batch, amounts)
	if err != nil {
		panic(err)
	}
}

func Burn(_ realm, from chain.Address, tid grc1155.TokenID, amount int64) {
	caller := runtime.OriginCaller()
	assertIsAdmin(caller)
	err := foo.Burn(from, tid, amount)
	if err != nil {
		panic(err)
	}
}

func BurnBatch(_ realm, from chain.Address, batch []grc1155.TokenID, amounts []int64) {
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

func assertIsAdmin(address_XXX chain.Address) {
	if address_XXX != admin {
		panic("restricted access")
	}
}

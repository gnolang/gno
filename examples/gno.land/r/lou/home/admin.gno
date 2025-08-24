package home

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/seqid"
)

var (
	admin   = std.Address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs")
	Ownable = ownable.NewWithAddress(admin)

	token  *grc20.Token
	ledger *grc20.PrivateLedger
)

func init() {
	CreateToken("LouTokens", "LTK", 2)
}

func CreateToken(name, symbol string, decimals uint) {
	token, ledger = grc20.NewToken(name, symbol, 2)
	ledger.Mint(admin, 100000)
}

func ResetRatings(_ realm) {
	Ownable.AssertOwnedByPrevious()

	ratings = avl.NewTree()
	ratingID = seqid.ID(0)
}

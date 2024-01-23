package mynonfungibletoken

import (
	"std"

	"gno.land/p/demo/grc/grc721"
)

var (
	admin std.Address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // set admin account
	// provision the token's name and symbol
	mynonfungibletoken = grc721.NewBasicNFT("mynonfungibletoken", "MNFT")
)

func init() {
	mintNNFT(admin, 10) // @administrator (supply = 10)
}

package gns

import (
	"std"
	"testing"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ownable"

	"gno.land/p/gnoswap/consts"
)

func resetObject(t *testing.T) {
	t.Helper()

	resetGnsTokenObject(t)

	height := std.ChainHeight()
	lastMintedHeight = height
}

func resetGnsTokenObject(t *testing.T) {
	t.Helper()

	token, privateLedger = grc20.NewToken("Gnoswap", "GNS", 6)
	UserTeller = token.CallerTeller()
	owner = ownable.NewWithAddress(consts.ADMIN)
	privateLedger.Mint(owner.Owner(), INITIAL_MINT_AMOUNT)
}

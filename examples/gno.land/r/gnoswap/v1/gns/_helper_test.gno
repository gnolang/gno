package gns

import (
	"std"
	"testing"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ownable"

	"gno.land/r/gnoswap/v1/access"
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

	adminAddr, _ := access.GetAddress(access.ROLE_ADMIN)
	owner = ownable.NewWithAddress(adminAddr)

	privateLedger.Mint(owner.Owner(), INITIAL_MINT_AMOUNT)
}

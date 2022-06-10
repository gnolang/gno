package faucet

import (
	"std"

	"gno.land/p/avl"
	"gno.land/p/ufmt"
)

var (
	gAdminAddr        std.Address
	gGiverAddr        std.Address = std.DerivePkgAddr("gno.land/r/faucet-community")
	gTotalTransferred std.Coins
	gTotalTransfers   = uint(0)
	gPerTransferSend  = std.Coins{std.Coin{"ugnot", 1000000}}
	gInPause          = false
	gMessage          = "# Community Faucet.\n\n"
)

func init() {
	// gAdminAddr = std.GetOrigCaller()
	// gGiverAddr = std.GetOrigCaller()
}

func Transfer(to std.Address) {
	assertIsGiver()

	// TODO: log history?
	// TODO: limit per dest?

	send := gPerTransferSend
	// gTotalTransferred = gTotalTransferred.Add(send)
	gTotalTransfers++

	banker := std.GetBanker(std.BankerTypeOrigSend)
	pkgaddr := std.GetOrigPkgAddr()
	banker.SendCoins(pkgaddr, to, send)
}

func Render(path string) string {
	banker := std.GetBanker(std.BankerTypeOrigSend)
	balance := banker.GetCoins(std.GetOrigPkgAddr())

	output := gMessage
	output += ufmt.Sprintf("Balance: %s\n", balance.String())
	output += ufmt.Sprintf("Rewarded: %s (in %d times)\n", gTotalTransferred.String(), gTotalTransfers)
	return output
}

func assertIsGiver() bool {
	// std.AssertOriginCall()
	caller := std.GetOrigCaller()
	return caller == gGiverAddr
}

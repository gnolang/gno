package faucet

import (
	"errors"
	"std"

	"gno.land/p/avl"
	"gno.land/p/ufmt"
)

var (
	gAdminAddr        std.Address = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq"
	gControllerAddr   std.Address = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq"
	gTotalTransferred std.Coins
	gTotalTransfers   = uint(0)
	gPerTransferSend  = std.Coins{std.Coin{"ugnot", 1000000}}
	gInPause          = false
	gMessage          = "# Community Faucet.\n\n"
)

func Transfer(to std.Address) error {
	if err := assertIsController(); err != nil {
		return err
	}

	// TODO: log history?
	// TODO: limit per dest?

	send := gPerTransferSend

	gTotalTransferred = gTotalTransferred.Add(send)
	gTotalTransfers++

	banker := std.GetBanker(std.BankerTypeOrigSend)
	pkgaddr := std.GetOrigPkgAddr()
	banker.SendCoins(pkgaddr, to, send)
	return nil
}

func Render(path string) string {
	banker := std.GetBanker(std.BankerTypeOrigSend)
	balance := banker.GetCoins(std.GetOrigPkgAddr())

	output := gMessage
	output += ufmt.Sprintf("Balance: %s\n", balance.String())
	output += ufmt.Sprintf("Rewarded: %s (in %d times)\n", gTotalTransferred.String(), gTotalTransfers)

	if path == "?debug" {
		output += ufmt.Sprintf("Admin: %s, Controller: %s\n", gAdminAddr.String(), gControllerAddr.String())
	}
	return output
}

func assertIsController() error {
	// std.AssertOriginCall()
	caller := std.GetOrigCaller()
	if caller != gControllerAddr {
		return errors.New("restricted for controller")
	}
	return nil
}

package faucet

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

var (
	// configurable by admin.
	gAdminAddr          std.Address = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	gControllers                    = avl.NewTree()
	gControllersMaxSize             = 10 // limit it to 10
	gInPause                        = false
	gMessage                        = "# Community Faucet.\n\n"

	// internal vars, for stats.
	gTotalTransferred std.Coins
	gTotalTransfers   = uint(0)

	// per request limit, 350 gnot
	gLimit std.Coin = std.Coin{"ugnot", 350000000}
)

func Transfer(to std.Address, send int64) string {
	if err := assertIsController(); err != nil {
		return err.Error()
	}

	if gInPause {
		return errors.New("faucet in pause").Error()
	}

	// limit the per request
	if send > gLimit.Amount {
		return errors.New("Per request limit " + gLimit.String() + " exceed").Error()
	}
	sendCoins := std.Coins{std.Coin{Denom: "ugnot", Amount: send}}

	gTotalTransferred = gTotalTransferred.Add(sendCoins)
	gTotalTransfers++

	banker := std.GetBanker(std.BankerTypeRealmSend)
	pkgaddr := std.CurrentRealm().Addr()
	banker.SendCoins(pkgaddr, to, sendCoins)
	return ""
}

func GetPerTransferLimit() int64 {
	return gLimit.Amount
}

func Render(_ string) string {
	banker := std.GetBanker(std.BankerTypeRealmSend)
	balance := banker.GetCoins(std.CurrentRealm().Addr())

	output := gMessage
	if gInPause {
		output += "Status: inactive.\n"
	} else {
		output += "Status: active.\n"
	}
	output += ufmt.Sprintf("Balance: %s.\n", balance.String())
	output += ufmt.Sprintf("Total transfers: %s (in %d times).\n\n", gTotalTransferred.String(), gTotalTransfers)

	output += "Package address: " + std.CurrentRealm().Addr().String() + "\n\n"
	output += ufmt.Sprintf("Admin: %s\n\n ", gAdminAddr.String())
	output += ufmt.Sprintf("Controllers:\n\n ")

	for i := 0; i < gControllers.Size(); i++ {
		_, v := gControllers.GetByIndex(i)
		output += ufmt.Sprintf("%s  ", v.(std.Address))
	}

	output += "\n\n"
	output += ufmt.Sprintf("Per request limit: %s\n\n", gLimit.String())

	return output
}

func assertIsController() error {
	caller := std.GetOrigCaller()

	ok := gControllers.Has(caller.String())
	if !ok {
		return errors.New(caller.String() + " is not on the controller list")
	}
	return nil
}

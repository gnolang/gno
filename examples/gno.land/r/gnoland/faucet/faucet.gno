package faucet

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"errors"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ufmt/v0"
)

var (
	// configurable by admin.
	gAdminAddr          address = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	gControllers                = avl.NewTree()
	gControllersMaxSize         = 10 // limit it to 10
	gInPause                    = false
	gMessage                    = "# Community Faucet.\n\n"

	// internal vars, for stats.
	gTotalTransferred chain.Coins
	gTotalTransfers   = uint(0)

	// per request limit, 350 gnot
	gLimit chain.Coin = chain.NewCoin("ugnot", 350_000_000)
)

func Transfer(cur realm, to address, send int64) string {
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
	sendCoins := chain.Coins{chain.NewCoin("ugnot", send)}

	gTotalTransferred = gTotalTransferred.Add(sendCoins)
	gTotalTransfers++

	banker_ := banker.NewBanker(banker.BankerTypeRealmSend)
	pkgaddr := runtime.CurrentRealm().Address()
	banker_.SendCoins(pkgaddr, to, sendCoins)
	return ""
}

func GetPerTransferLimit() int64 {
	return gLimit.Amount
}

func bankerAddr(cur realm) address {
	return runtime.CurrentRealm().Address()
}

func Render(_ string) string {
	banker_ := banker.NewBanker(banker.BankerTypeReadonly)
	balance := banker_.GetCoins(bankerAddr(cross))

	output := gMessage
	if gInPause {
		output += "Status: inactive.\n"
	} else {
		output += "Status: active.\n"
	}
	output += ufmt.Sprintf("Balance: %s.\n", balance.String())
	output += ufmt.Sprintf("Total transfers: %s (in %d times).\n\n", gTotalTransferred.String(), gTotalTransfers)

	output += "Package address: " + runtime.CurrentRealm().Address().String() + "\n\n"
	output += ufmt.Sprintf("Admin: %s\n\n ", gAdminAddr.String())
	output += ufmt.Sprintf("Controllers:\n\n ")

	for i := 0; i < gControllers.Size(); i++ {
		_, v := gControllers.GetByIndex(i)
		output += ufmt.Sprintf("%s  ", v.(address))
	}

	output += "\n\n"
	output += ufmt.Sprintf("Per request limit: %s\n\n", gLimit.String())

	return output
}

func assertIsController() error {
	caller := runtime.PreviousRealm().Address()
	ok := gControllers.Has(caller.String())
	if !ok {
		return errors.New(caller.String() + " is not on the controller list")
	}
	return nil
}

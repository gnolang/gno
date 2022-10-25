package faucet

import (
	"errors"
	"std"

	"gno.land/p/ufmt"
	FA "gno.land/r/faucet-admin" // faucet admin
)

var (
	// configurable by only admin
	// use getter func from 'gno.land/r/faucet-admin'
	gInPause = FA.GetInPause()

	// internal vars, for stats
	gTotalTransferred std.Coins
	gTotalTransfers   = uint(0)
)


func Transfer(to std.Address) error {
	if err := assertIsController(); err != nil {
		return err
	}

	if gInPause {
		return errors.New("faucet in pause")
	}

	send := FA.GetPerTransferSend(to)

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

	output := FA.GetMessage()
	if gInPause {
		output += "Status: inactive.\n"
	} else {
		output += "Status: active.\n"
	}
	output += ufmt.Sprintf("Balance: %s.\n", balance.String())
	output += ufmt.Sprintf("Faucet total transfered: %s (in %d times).\n\n", gTotalTransferred.String(), gTotalTransfers)

	output += ufmt.Sprintf("Admin: %", FA.GetAdminAddress())
	output += ufmt.Sprintf("Controller: \n\n")

	for _, v := range FA.GetControllerAddress() {
		output += ufmt.Sprintf("- %s \n\n", v)
	}

	output += ufmt.Sprintf("\n\n")

	output += ufmt.Sprintf("Facuet Rules\n\n")
	output += ufmt.Sprintf(" 1. you'll receive 10 GNOT for 1st faucet\n")
	output += ufmt.Sprintf(" 2. you'll receive 1 GNOT for next faucet\n\n")

	// if path == "?debug" {
	// 	output += ufmt.Sprintf("Admin: %s, Controller: %s\n", gAdminAddr.String(), gControllerAddr.String())
	// }
	return output
}

func assertIsController() error {
	caller := std.GetOrigCaller()
	if caller != FA.GetControllerAddress() {
		return errors.New("restricted for controller")
	}
	return nil
}
package main

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/r/gnoland/faucet"
)

// mints coints to current realm
func init() {
	faucetaddr := chain.PackageAddress("gno.land/r/gnoland/faucet")
	testing.IssueCoins(faucetaddr, chain.Coins{{"ugnot", 200_000_000}})
}

// assert render with 2 controllers and 2 transfers
func main() {
	var (
		adminaddr       = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
		controlleraddr1 = testutils.TestAddress("controller1")
		controlleraddr2 = testutils.TestAddress("controller2")
		testaddr1       = testutils.TestAddress("test1")
		_               = testutils.TestAddress("test2")
	)
	testing.SetRealm(testing.NewUserRealm(adminaddr))
	err := faucet.AdminAddController(cross, controlleraddr1)
	if err != "" {
		panic(err)
	}
	err = faucet.AdminAddController(cross, controlleraddr2)
	if err != "" {
		panic(err)
	}
	testing.SetRealm(testing.NewUserRealm(controlleraddr1))
	err = faucet.Transfer(cross, testaddr1, 1_000_000)
	if err != "" {
		panic(err)
	}
	testing.SetRealm(testing.NewUserRealm(controlleraddr2))
	err = faucet.Transfer(cross, testaddr1, 2_000_000)
	if err != "" {
		panic(err)
	}
	println(faucet.Render(""))
}

// Output:
// # Community Faucet.
//
// Status: active.
// Balance: 197000000ugnot.
// Total transfers: 3000000ugnot (in 2 times).
//
// Package address: g1vdhkuarjdakxcetjxf047h6lta047h6lnrev3v
//
// Admin: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
//
//  Controllers:
//
//  g1vdhkuarjdakxcetjx9047h6lta047h6lsdacav  g1vdhkuarjdakxcetjxf047h6lta047h6lnrev3v
//
// Per request limit: 350000000ugnot

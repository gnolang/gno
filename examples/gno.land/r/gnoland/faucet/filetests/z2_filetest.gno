package main

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/gnoland/faucet"
)

// mints ugnot to current realm
func init() {
	faucetaddr := chain.PackageAddress("gno.land/r/gnoland/faucet")
	testing.IssueCoins(faucetaddr, chain.Coins{{"ugnot", 200_000_000}})
}

// assert render with empty path and 2 controllers
func main() {
	var (
		adminaddr       = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
		controlleraddr1 = testutils.TestAddress("controller1")
		controlleraddr2 = testutils.TestAddress("controller2")
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
	println(faucet.Render(""))
}

// Output:
// # Community Faucet.
//
// Status: active.
// Balance: 200000000ugnot.
// Total transfers:  (in 0 times).
//
// Package address: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
//
// Admin: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
//
//  Controllers:
//
//  g1vdhkuarjdakxcetjx9047h6lta047h6lsdacav  g1vdhkuarjdakxcetjxf047h6lta047h6lnrev3v
//
// Per request limit: 350000000ugnot

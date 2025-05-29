package main

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/gnoland/faucet"
)

// mints ugnot to current realm
func init() {
	faucetaddr := std.DerivePkgAddr("gno.land/r/gnoland/faucet")
	testing.IssueCoins(faucetaddr, std.Coins{{"ugnot", 200_000_000}})
}

// assert render with empty path and 2 controllers
func main() {
	var (
		adminaddr       = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
		controlleraddr1 = testutils.TestAddress("controller1")
		controlleraddr2 = testutils.TestAddress("controller2")
	)
	testing.SetRealm(std.NewUserRealm(adminaddr))
	err := cross(faucet.AdminAddController)(controlleraddr1)
	if err != "" {
		panic(err)
	}
	err = cross(faucet.AdminAddController)(controlleraddr2)
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

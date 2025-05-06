package main

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/gnoland/faucet"
)

// mints coints to current realm
func init() {
	faucetaddr := std.DerivePkgAddr("gno.land/r/gnoland/faucet")
	testing.IssueCoins(faucetaddr, std.Coins{{"ugnot", 200_000_000}})
}

// assert render with 2 controllers and 2 transfers
func main() {
	var (
		adminaddr       = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
		controlleraddr1 = testutils.TestAddress("controller1")
		controlleraddr2 = testutils.TestAddress("controller2")
		testaddr1       = testutils.TestAddress("test1")
		testaddr2       = testutils.TestAddress("test2")
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
	testing.SetRealm(std.NewUserRealm(controlleraddr1))
	err = cross(faucet.Transfer)(testaddr1, 1_000_000)
	if err != "" {
		panic(err)
	}
	testing.SetRealm(std.NewUserRealm(controlleraddr2))
	err = cross(faucet.Transfer)(testaddr1, 2_000_000)
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

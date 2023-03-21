package main

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/gnoland/faucet"
)

// assert render with 2 controllers and 2 transfers
func main() {
	var (
		adminaddr       = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
		controlleraddr1 = testutils.TestAddress("controller1")
		controlleraddr2 = testutils.TestAddress("controller2")
		testaddr1       = testutils.TestAddress("test1")
		testaddr2       = testutils.TestAddress("test2")
	)
	std.TestSetOrigCaller(adminaddr)
	err := faucet.AdminAddController(controlleraddr1)
	if err != nil {
		panic(err)
	}
	err = faucet.AdminAddController(controlleraddr2)
	if err != nil {
		panic(err)
	}
	std.TestSetOrigCaller(controlleraddr1)
	err = faucet.Transfer(testaddr1, 1000000)
	if err != nil {
		panic(err)
	}
	std.TestSetOrigCaller(controlleraddr2)
	err = faucet.Transfer(testaddr1, 2000000)
	if err != nil {
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
// Package address: g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4
//
// Admin: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
//
//  Controllers:
//
//  (unhandled)  (unhandled)
//
// Per request limit: 350000000ugnot

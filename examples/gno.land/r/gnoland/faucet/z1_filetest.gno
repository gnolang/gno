package main

import (
	"std"
	"testing"

	"gno.land/r/gnoland/faucet"
)

// mints ugnot to current realm
func init() {
	faucetaddr := std.DerivePkgAddr("gno.land/r/gnoland/faucet")
	testing.IssueCoins(faucetaddr, std.Coins{{"ugnot", 200_000_000}})
}

// assert render with a path and no controllers
func main() {
	println(faucet.Render("path"))
}

// Output:
// # Community Faucet.
//
// Status: active.
// Balance: 200000000ugnot.
// Total transfers:  (in 0 times).
//
// Package address: g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4
//
// Admin: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
//
//  Controllers:
//
//
//
// Per request limit: 350000000ugnot

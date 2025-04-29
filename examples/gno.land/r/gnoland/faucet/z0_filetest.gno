package main

import (
	"std"
	"testing"

	"gno.land/r/gnoland/faucet"
)

// mints ugnot to current realm
func init() {
	facuetaddr := std.DerivePkgAddr("gno.land/r/gnoland/faucet")
	testing.IssueCoins(facuetaddr, std.Coins{{"ugnot", 200000000}})
}

// assert render with empty path and no controllers
func main() {
	println(faucet.Render(""))
}

// Output:
// # Community Faucet.
//
// Status: active.
// Balance: .
// Total transfers:  (in 0 times).
//
// Package address: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// Admin: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
//
//  Controllers:
//
//
//
// Per request limit: 350000000ugnot

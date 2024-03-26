package main

import (
	"std"

	"gno.land/r/gnoland/faucet"
)

// mints ugnot to current realm
func init() {
	facuetaddr := std.DerivePkgAddr("gno.land/r/gnoland/faucet")
	std.TestIssueCoins(facuetaddr, std.Coins{{"ugnot", 200000000}})
}

// assert render with empty path and no controllers
func main() {
	println(faucet.Render(""))
}

// Output:
// # Community Faucet.
//
// Status: active.
// Balance: 200000000ugnot.
// Total transfers:  (in 0 times).
//
// Package address: g1ttrq7mp4zy6dssnmgyyktnn4hcj3ys8xhju0n7
//
// Admin: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
//
//  Controllers:
//
//
//
// Per request limit: 350000000ugnot

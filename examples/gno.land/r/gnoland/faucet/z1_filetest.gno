package main

import (
	"gno.land/r/gnoland/faucet"
)

// assert render with a path and no controllers
func main() {
	println(faucet.Render("path"))
}

// Output:
// path# Community Faucet.
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

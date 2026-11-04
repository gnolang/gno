// PKGPATH: gno.land/r/gnoland/faucet/filetests/z1_filetest

package z1_filetest

import (
	"chain"
	"testing"

	"gno.land/r/gnoland/faucet"
)

// mints ugnot to current realm
func init() {
	faucetaddr := chain.PackageAddress("gno.land/r/gnoland/faucet")
	testing.IssueCoins(faucetaddr, chain.Coins{{"ugnot", 200_000_000}})
}

// assert render with a path and no controllers
func main(cur realm) {
	println(faucet.Render(cross(cur), "path"))
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
// Admin: g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx
//
//  Controllers:
//
//
//
// Per request limit: 350000000ugnot

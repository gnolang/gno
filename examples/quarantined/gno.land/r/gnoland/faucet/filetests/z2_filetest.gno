// PKGPATH: gno.land/r/gnoland/faucet/filetests/z2_filetest

package z2_filetest

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
func main(cur realm) {
	var (
		adminaddr       = address("g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf")
		controlleraddr1 = testutils.TestAddress("controller1")
		controlleraddr2 = testutils.TestAddress("controller2")
	)
	testing.SetRealm(testing.NewUserRealm(adminaddr))
	err := faucet.AdminAddController(cross(cur), controlleraddr1)
	if err != "" {
		panic(err)
	}
	err = faucet.AdminAddController(cross(cur), controlleraddr2)
	if err != "" {
		panic(err)
	}
	println(faucet.Render(cross(cur), ""))
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
// Admin: g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf
//
//  Controllers:
//
//  g1vdhkuarjdakxcetjx9047h6lta047h6lsdacav  g1vdhkuarjdakxcetjxf047h6lta047h6lnrev3v
//
// Per request limit: 350000000ugnot

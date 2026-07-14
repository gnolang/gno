// PKGPATH: gno.land/r/demo/grc20reg_filetest

package grc20reg_filetest

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/demo/defi/grc20reg"
	"gno.land/r/tests/vm/test20"
)

func main(cur realm) {
	const tokenKey = "gno.land/r/tests/vm/test20"
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	carol := testutils.TestAddress("carol")

	if err := test20.PrivateLedger.Mint(alice, 1_000); err != nil {
		panic(err)
	}

	testing.SetRealm(testing.NewUserRealm(alice))
	grc20reg.Transfer(cross(cur), tokenKey, bob, 100)
	grc20reg.Approve(cross(cur), tokenKey, carol, 300)

	testing.SetRealm(testing.NewUserRealm(carol))
	grc20reg.TransferFrom(cross(cur), tokenKey, alice, bob, 200)

	println("alice balance:", test20.Token.BalanceOf(alice))
	println("bob balance:", test20.Token.BalanceOf(bob))
	println("carol allowance:", test20.Token.Allowance(alice, carol))
}

// Output:
// alice balance: 700
// bob balance: 300
// carol allowance: 100

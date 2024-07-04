package bar20

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

func TestPackage(t *testing.T) {
	alice := testutils.TestAddress("alice")
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice) // XXX: should not need this

	balance, _ := Token.BalanceOf(alice)
	expected := uint64(0)
	uassert.Equal(t, expected, balance, "balance does not match")

	ret := Faucet()
	uassert.Equal(t, "OK", ret, "faucet should be OK")

	balance, _ = Token.BalanceOf(alice)
	expected = uint64(1_000_000)
	uassert.Equal(t, expected, balance, "balance does not match")
}

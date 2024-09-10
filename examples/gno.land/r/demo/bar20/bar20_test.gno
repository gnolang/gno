package bar20

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"
)

func TestPackage(t *testing.T) {
	alice := testutils.TestAddress("alice")
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice) // XXX: should not need this

	urequire.Equal(t, Token.BalanceOf(alice), uint64(0))
	urequire.Equal(t, Faucet(), "OK")
	urequire.Equal(t, Token.BalanceOf(alice), uint64(1_000_000))
}

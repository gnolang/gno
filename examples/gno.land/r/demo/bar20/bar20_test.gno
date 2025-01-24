package bar20

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"
)

func TestPackage(t *testing.T) {
	alice := testutils.TestAddress("alice")
	t.SetPreviousRealm(std.NewUserRealm(alice))

	urequire.Equal(t, UserTeller.BalanceOf(alice), uint64(0))
	urequire.Equal(t, Faucet(), "OK")
	urequire.Equal(t, UserTeller.BalanceOf(alice), uint64(1_000_000))
}

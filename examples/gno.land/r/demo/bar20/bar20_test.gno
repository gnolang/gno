package bar20

import (
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"
)

func TestPackage(t *testing.T) {
	alice := testutils.TestAddress("alice")
	testing.SetOriginCaller(alice)

	urequire.Equal(t, UserTeller.BalanceOf(alice), int64(0))
	urequire.Equal(t, cross(Faucet)(), "OK")
	urequire.Equal(t, UserTeller.BalanceOf(alice), int64(1_000_000))
}

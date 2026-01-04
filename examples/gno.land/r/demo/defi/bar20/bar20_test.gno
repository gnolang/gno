package bar20

import (
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/urequire"
)

func TestPackage(t *testing.T) {
	alice := testutils.TestAddress("alice")
	testing.SetOriginCaller(alice)

	urequire.Equal(t, UserTeller.BalanceOf(alice), int64(0))
	urequire.Equal(t, Faucet(cross), "OK")
	urequire.Equal(t, UserTeller.BalanceOf(alice), int64(1_000_000))
}

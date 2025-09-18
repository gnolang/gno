package v1

import (
	"std"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
	"gno.land/p/onbloc/uint256"
)

func TestToken(t *testing.T) {
	creator := testutils.TestAddress("creator")
	testing.SetRealm(std.NewUserRealm(creator))
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	tokenID := Create(cross, "Test Token", "TEST", alice, uint256.NewUint(100))

	uassert.Equal(t, "Test Token", Name(cross, tokenID))
	uassert.Equal(t, "TEST", Symbol(cross, tokenID))
	uassert.Equal(t, uint256.NewUint(100).String(), BalanceOf(cross, tokenID, alice).String())

	testing.SetRealm(std.NewUserRealm(alice))
	Transfer(cross, tokenID, bob, uint256.NewUint(27))
	uassert.Equal(t, uint256.NewUint(73).String(), BalanceOf(cross, tokenID, alice).String())
	uassert.Equal(t, uint256.NewUint(27).String(), BalanceOf(cross, tokenID, bob).String())
}

package names

import (
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
	"gno.land/r/sys/users"
)

var alice = testutils.TestAddress("alice")

func TestIsAuthorizedAddressForNamespace(t *testing.T) {
	// username + addr mismatch
	uassert.False(t, IsAuthorizedAddressForNamespace(alice, "notregistered"))
	// PA namespace check
	uassert.True(t, IsAuthorizedAddressForNamespace(alice, alice.String()))

	// Empty name/address
	uassert.False(t, IsAuthorizedAddressForNamespace(address(""), ""))

	// Register proper username
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/users/v1")) // authorized write
	urequire.NoError(t, users.RegisterUser(cross, "alice", alice))

	// Proper namespace
	uassert.True(t, IsAuthorizedAddressForNamespace(alice, "alice"))

	// Wrong address for namespace
	uassert.False(t, IsAuthorizedAddressForNamespace(testutils.TestAddress("bob"), "alice"))
}

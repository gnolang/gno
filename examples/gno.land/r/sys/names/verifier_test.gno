package names

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"

	"gno.land/r/sys/users"
)

var alice = testutils.TestAddress("alice")

func TestDefaultVerifier(t *testing.T) {
	// Check enabled
	// username + addr mismatch
	uassert.False(t, defaultVerifier(alice, "notregistered"))
	// PA namespace check
	uassert.True(t, defaultVerifier(alice, alice.String()))

	// Empty name/address
	uassert.False(t, defaultVerifier(std.Address(""), ""))

	//Register proper username
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/gnoland/users/v1")) // authorized write
	std.TestSetOrigCaller(std.DerivePkgAddr("gno.land/r/gnoland/users/v1"))
	urequire.NoError(t, users.RegisterUser("alice", alice))

	// Proper namespace
	uassert.True(t, defaultVerifier(alice, "alice"))
}

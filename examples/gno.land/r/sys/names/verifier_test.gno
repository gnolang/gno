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

	// Check disabled, any case is true
	uassert.True(t, defaultVerifier(false, alice, alice.String()))
	uassert.True(t, defaultVerifier(false, "", alice.String()))
	uassert.True(t, defaultVerifier(false, adminAddr, "somerandomusername"))

	// Check enabled
	// username + addr mismatch
	uassert.False(t, defaultVerifier(true, alice, "notregistered"))
	// PA namespace check
	uassert.True(t, defaultVerifier(true, alice, alice.String()))

	// Empty name/address
	uassert.False(t, defaultVerifier(true, std.Address(""), ""))

	//Register proper username
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/gnoland/users")) // authorized write
	std.TestSetOrigCaller(std.DerivePkgAddr("gno.land/r/gnoland/users"))
	urequire.NoError(t, users.RegisterUser("alice", alice))

	// Proper namespace
	uassert.True(t, defaultVerifier(true, alice, "alice"))
}

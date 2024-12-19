package names

import (
	"std"
	"testing"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
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

	//Register proper username fixme: doesn't work
	//std.TestSetRealm(std.NewCodeRealm("gno.land/r/gnoland/users")) // authorized write
	//std.TestSetOrigCaller(std.DerivePkgAddr("gno.land/r/gnoland/users"))
	//urequire.NoError(t, users.WriteUser("alice", alice))
	//// Proper namespace
	//uassert.True(t, defaultVerifier(true, alice, "alice"))
}

func TestUpdateVerifyCall(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice)

	newFunc := func(bool, std.Address, string) bool {
		return true
	}

	uassert.Error(t, UpdateVerifyCall(newFunc), ownable.ErrUnauthorized.Error())

	std.TestSetRealm(std.NewUserRealm(adminAddr))
	std.TestSetOrigCaller(adminAddr)

	uassert.NoError(t, UpdateVerifyCall(newFunc))

	// These check would fail with the default verifier
	uassert.True(t, checkFunc(true, alice, "notregistered"))
	uassert.True(t, checkFunc(true, std.Address(""), ""))
}

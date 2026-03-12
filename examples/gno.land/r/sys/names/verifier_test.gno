package names

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

var alice = testutils.TestAddress("alice")

func TestDefaultVerifier(t *testing.T) {
	// Disabled: any case is true
	uassert.True(t, verifier(false, alice, alice.String()))
	uassert.True(t, verifier(false, "", alice.String()))
	uassert.True(t, verifier(false, alice, "somerandomusername"))

	// Enabled: PA namespace check
	uassert.True(t, verifier(true, alice, alice.String()))

	// Enabled: non-PA namespaces denied
	uassert.False(t, verifier(true, alice, "notregistered"))
	uassert.False(t, verifier(true, alice, "alice"))

	// Enabled: empty name/address
	uassert.False(t, verifier(true, address(""), ""))
	uassert.False(t, verifier(true, alice, ""))
	uassert.False(t, verifier(true, address(""), "something"))
}

func TestEnable(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("random")))
	uassert.AbortsWithMessage(t, "caller is not admin", func() {
		Enable(cross)
	})
	uassert.False(t, IsEnabled())

	testing.SetRealm(testing.NewUserRealm(admin))

	uassert.NotPanics(t, func() {
		Enable(cross)
	})
	uassert.True(t, IsEnabled())
	uassert.NotPanics(t, func() {
		Enable(cross)
	})
}

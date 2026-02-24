package names

import (
	"testing"

	"gno.land/p/nt/ownable"
	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
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
	testing.SetRealm(testing.NewUserRealm("g1manfred47kzduec920z88wfr64ylksmdcedlf5"))

	uassert.NotPanics(t, func() {
		Enable(cross)
	})

	// Confirm enable drops ownership
	uassert.Equal(t, Ownable.Owner().String(), "")
	uassert.AbortsWithMessage(t, ownable.ErrUnauthorized.Error(), func() {
		Enable(cross)
	})
}

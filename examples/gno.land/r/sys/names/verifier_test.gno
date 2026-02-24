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

func TestReservedNamespaces(t *testing.T) {
	// Reserved namespaces denied when enabled
	uassert.False(t, verifier(true, alice, "gnoland"))
	uassert.False(t, verifier(true, alice, "nt"))
	uassert.False(t, verifier(true, alice, "sys"))
	uassert.False(t, verifier(true, alice, "gov"))

	// Reserved namespaces allowed when disabled (pre-genesis)
	uassert.True(t, verifier(false, alice, "gnoland"))
	uassert.True(t, verifier(false, alice, "sys"))
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

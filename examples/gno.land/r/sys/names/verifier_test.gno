package names

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

var alice = testutils.TestAddress("alice")

func TestIsAuthorizedAddressForNamespace(t *testing.T) {
	// Address does not match namespace
	uassert.False(t, IsAuthorizedAddressForNamespace(alice, "notregistered"))

	// PA namespace: address matches its own string representation
	uassert.True(t, IsAuthorizedAddressForNamespace(alice, alice.String()))

	// Empty address and namespace
	uassert.False(t, IsAuthorizedAddressForNamespace(address(""), ""))

	// Different address for the same namespace
	uassert.False(t, IsAuthorizedAddressForNamespace(testutils.TestAddress("bob"), alice.String()))
}

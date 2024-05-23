package nonce

import (
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
)

// TestNonceManager tests the Manager functions
func TestNonceManager(t *testing.T) {
	nm := New()

	// Use a dummy address for testing
	address := testutils.TestAddress("dummy")

	// Test GetNonce method
	if nonce := nm.GetNonce(address); nonce != 0 {
		t.Errorf("expected nonce to be 0, got %d", nonce)
	}

	// Test UseNonce method
	nm.UseNonce(address)
	if newNonce := nm.GetNonce(address); newNonce != 1 {
		t.Errorf("expected new nonce to be 1, got %d", newNonce)
	}

	nm.UseNonce(address)
	if finalNonce := nm.GetNonce(address); finalNonce != 2 {
		t.Errorf("expected final nonce to be 2, got %d", finalNonce)
	}
}

func TestMultipleAddresses(t *testing.T) {
	nm := New()

	// Use dummy addresses for testing
	address1 := testutils.TestAddress("bob")
	address2 := testutils.TestAddress("alice")

	// Test UseNonce with multiple addresses

	if nonce := nm.GetNonce(address1); nonce != 0 {
		t.Errorf("expected final nonce to be 0, got %d", finalNonce)
	}

	if nonce := nm.GetNonce(address1); nonce != 0 {
		t.Errorf("expected final nonce to be 0, got %d", finalNonce)
	}

	nm.UseNonce(address1)
	if nonce := nm.UseNonce(address1); nonce != 1 {
		t.Errorf("expected used nonce to be 1 for address1, got %d", nonce)
	}

	nm.UseNonce(address2)
	if nonce := nm.UseNonce(address2); nonce != 1 {
		t.Errorf("expected used nonce to be 1 for address2, got %d", nonce)
	}
}

func TestConcurrentUse(t *testing.T) {
	nm := New()
	address := testutils.TestAddress("dummy")

	for i := 0; i < 10; i++ {
		t.Parallel()

		nm.UseNonce(address)
		if nonce := nm.GetNonce(address); nonce != uint64(i) {
			t.Errorf("expected used nonce to be %d, got %d", i, nonce)
		}
	}
}

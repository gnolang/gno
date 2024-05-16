package nonces

import (
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
)

// TestNonceManager tests the NonceManager functions
func TestNonceManager(t *testing.T) {
	nm := New()

	// Use a dummy address for testing
	address := testutils.TestAddress("dummy")

	// Initialize the address in the AVL tree
	nm.SetNonce(address, 0)

	// Test GetNonce method
	if nonce := nm.GetNonce(address); nonce != 0 {
		t.Errorf("expected nonce to be 0, got %d", nonce)
	}

	// Test UseNonce method
	if nonce := nm.UseNonce(address); nonce != 0 {
		t.Errorf("expected used nonce to be 0, got %d", nonce)
	}

	if newNonce := nm.GetNonce(address); newNonce != 1 {
		t.Errorf("expected new nonce to be 1, got %d", newNonce)
	}

	// Test UseCheckedNonce method with correct nonce
	if err := nm.UseCheckedNonce(address, 1); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if finalNonce := nm.GetNonce(address); finalNonce != 2 {
		t.Errorf("expected final nonce to be 2, got %d", finalNonce)
	}

	// Test UseCheckedNonce method with incorrect nonce
	err := nm.UseCheckedNonce(address, 1)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	expectedErr := ufmt.Sprintf("InvalidAccountNonce: expected 2, got 1 for account %s", address)
	if err.Error() != expectedErr {
		t.Errorf("expected error message '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestMultipleAddresses(t *testing.T) {
	nm := New()

	// Use dummy addresses for testing
	address1 := testutils.TestAddress("bob")
	address2 := testutils.TestAddress("alice")

	// Initialize the addresses in the AVL tree
	nm.SetNonce(address1, 0)
	nm.SetNonce(address2, 0)

	// Test UseNonce with multiple addresses
	if nonce := nm.UseNonce(address1); nonce != 0 {
		t.Errorf("expected used nonce to be 0 for address1, got %d", nonce)
	}

	if nonce := nm.UseNonce(address2); nonce != 0 {
		t.Errorf("expected used nonce to be 0 for address2, got %d", nonce)
	}

	if nonce := nm.UseNonce(address1); nonce != 1 {
		t.Errorf("expected used nonce to be 1 for address1, got %d", nonce)
	}

	if nonce := nm.UseNonce(address2); nonce != 1 {
		t.Errorf("expected used nonce to be 1 for address2, got %d", nonce)
	}
}

func TestConcurrentUse(t *testing.T) {
	nm := New()
	address := testutils.TestAddress("dummy")

	// Initialize the address in the AVL tree
	nm.SetNonce(address, 0)

	for i := 0; i < 10; i++ {
		t.Parallel()

		if nonce := nm.UseNonce(address); nonce != uint64(i) {
			t.Errorf("expected used nonce to be %d, got %d", i, nonce)
		}
	}
}

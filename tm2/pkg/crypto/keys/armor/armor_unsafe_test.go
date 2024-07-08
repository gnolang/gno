package armor

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
)

// TestArmorUnarmor_PrivKey_Unencrypted verifies that an unencrypted private key
// can be correctly armored and unarmored
func TestArmorUnarmor_PrivKey_Unencrypted(t *testing.T) {
	t.Parallel()

	// Generate a random private key
	randomPrivateKey := secp256k1.GenPrivKey()

	// Armor it, then unarmor it
	unarmoredPrivateKey, err := UnarmorPrivateKey(ArmorPrivateKey(randomPrivateKey))
	if err != nil {
		t.Fatalf("unable to unarmor private key, %v", err)
	}

	// Make sure the keys match
	assert.True(t, randomPrivateKey.Equals(unarmoredPrivateKey))
}

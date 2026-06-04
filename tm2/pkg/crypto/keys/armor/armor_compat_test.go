package armor

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto/armor"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/stretchr/testify/require"
)

func TestArmorUnarmor_PrivKey_EmptyPassBackward(t *testing.T) {
	t.Parallel()

	// This is a backward compatibility test for the old armor format.
	// Previously, empty passphrases would still nonetheless be encrypted with
	// bcrypt, which is slow and expensive. With this test, we ensure that
	// passwordless keys with the old system continue working.

	priv := secp256k1.GenPrivKey()
	saltBytes, encBytes := encryptPrivKey(priv, "")
	header := map[string]string{
		"kdf":  "bcrypt",
		"salt": fmt.Sprintf("%X", saltBytes),
	}
	armorStr := armor.EncodeArmor(blockTypePrivKey, header, encBytes)

	_, err := UnarmorDecryptPrivKey(armorStr, "wrongpassphrase")
	require.Error(t, err)
	decrypted, err := UnarmorDecryptPrivKey(armorStr, "")
	require.NoError(t, err)
	require.True(t, priv.Equals(decrypted))
}

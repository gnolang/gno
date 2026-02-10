package armor_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/armor"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

func TestArmorUnarmor_PrivKey_Encrypted(t *testing.T) {
	t.Parallel()

	priv := secp256k1.GenPrivKey()
	astr := armor.EncryptArmorPrivKey(priv, "passphrase")
	_, err := armor.UnarmorDecryptPrivKey(astr, "wrongpassphrase")
	require.Error(t, err)
	decrypted, err := armor.UnarmorDecryptPrivKey(astr, "passphrase")
	require.NoError(t, err)
	require.True(t, priv.Equals(decrypted))
}

func TestArmorUnarmor_PrivKey_EmptyPass(t *testing.T) {
	t.Parallel()

	priv := secp256k1.GenPrivKey()
	astr := armor.EncryptArmorPrivKey(priv, "")
	_, err := armor.UnarmorDecryptPrivKey(astr, "wrongpassphrase")
	require.Error(t, err)
	decrypted, err := armor.UnarmorDecryptPrivKey(astr, "")
	require.NoError(t, err)
	require.True(t, priv.Equals(decrypted))
}

func TestArmorUnarmor_PubKey(t *testing.T) {
	t.Parallel()

	// Select the encryption and storage for your cryptostore
	cstore := keys.NewInMemory()

	// Add keys and see they return in alphabetical order
	mn1 := `lounge napkin all odor tilt dove win inject sleep jazz uncover traffic hint require cargo arm rocket round scan bread report squirrel step lake`
	bip39Passphrase := ""
	info, err := cstore.CreateAccount("Bob", mn1, bip39Passphrase, "passphrase", 0, 0)
	require.NoError(t, err)
	astr := armor.ArmorPubKeyBytes(info.GetPubKey().Bytes())
	pubBytes, err := armor.UnarmorPubKeyBytes(astr)
	require.NoError(t, err)
	pub, err := crypto.PubKeyFromBytes(pubBytes)
	require.NoError(t, err)
	require.True(t, pub.Equals(info.GetPubKey()))
}

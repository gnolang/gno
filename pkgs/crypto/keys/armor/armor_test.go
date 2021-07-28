package armor_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/crypto/keys/armor"
	"github.com/gnolang/gno/pkgs/crypto/secp256k1"
)

func TestArmorUnarmorPrivKey(t *testing.T) {
	priv := secp256k1.GenPrivKey()
	astr := armor.EncryptArmorPrivKey(priv, "passphrase")
	_, err := armor.UnarmorDecryptPrivKey(astr, "wrongpassphrase")
	require.Error(t, err)
	decrypted, err := armor.UnarmorDecryptPrivKey(astr, "passphrase")
	require.NoError(t, err)
	require.True(t, priv.Equals(decrypted))
}

func TestArmorUnarmorPubKey(t *testing.T) {
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

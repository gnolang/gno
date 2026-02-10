package armor

import (
	"encoding/hex"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/armor"
	"github.com/gnolang/gno/tm2/pkg/crypto/bcrypt"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/keyerror"
	"github.com/gnolang/gno/tm2/pkg/crypto/xsalsa20symmetric"
	"github.com/gnolang/gno/tm2/pkg/os"
)

const (
	blockTypePrivKey        = "TENDERMINT PRIVATE KEY"
	blockTypeKeyInfo        = "TENDERMINT KEY INFO"
	blockTypePubKey         = "TENDERMINT PUBLIC KEY"
	bcryptSecurityParameter = 12
)

// -----------------------------------------------------------------
// add armor

// Armor the InfoBytes
func ArmorInfoBytes(bz []byte) string {
	return armorBytes(bz, blockTypeKeyInfo)
}

// Armor the PubKeyBytes
func ArmorPubKeyBytes(bz []byte) string {
	return armorBytes(bz, blockTypePubKey)
}

func armorBytes(bz []byte, blockType string) string {
	header := map[string]string{
		"type":    "Info",
		"version": "0.0.0",
	}
	return armor.EncodeArmor(blockType, header, bz)
}

// -----------------------------------------------------------------
// remove armor

// Unarmor the InfoBytes
func UnarmorInfoBytes(armorStr string) (bz []byte, err error) {
	return unarmorBytes(armorStr, blockTypeKeyInfo)
}

// Unarmor the PubKeyBytes
func UnarmorPubKeyBytes(armorStr string) (bz []byte, err error) {
	return unarmorBytes(armorStr, blockTypePubKey)
}

func unarmorBytes(armorStr, blockType string) (bz []byte, err error) {
	bType, header, bz, err := armor.DecodeArmor(armorStr)
	if err != nil {
		return
	}
	if bType != blockType {
		err = fmt.Errorf("unrecognized armor type %q, expected: %q", bType, blockType)
		return
	}
	if header["version"] != "0.0.0" {
		err = fmt.Errorf("unrecognized version: %v", header["version"])
		return
	}
	return
}

// -----------------------------------------------------------------
// encrypt/decrypt with armor

// Encrypt and armor the private key.
func EncryptArmorPrivKey(privKey crypto.PrivKey, passphrase string) string {
	if passphrase == "" {
		return ArmorPrivateKey(privKey)
	}
	saltBytes, encBytes := encryptPrivKey(privKey, passphrase)
	header := map[string]string{
		"kdf":  "bcrypt",
		"salt": fmt.Sprintf("%X", saltBytes),
	}
	armorStr := armor.EncodeArmor(blockTypePrivKey, header, encBytes)
	return armorStr
}

// encrypt the given privKey with the passphrase using a randomly
// generated salt and the xsalsa20 cipher. returns the salt and the
// encrypted priv key.
func encryptPrivKey(privKey crypto.PrivKey, passphrase string) (saltBytes []byte, encBytes []byte) {
	saltBytes = crypto.CRandBytes(16)
	key, err := bcrypt.GenerateFromPassword(saltBytes, []byte(passphrase), bcryptSecurityParameter)
	if err != nil {
		os.Exit("Error generating bcrypt key from passphrase: " + err.Error())
	}
	key = crypto.Sha256(key) // get 32 bytes
	privKeyBytes := privKey.Bytes()
	return saltBytes, xsalsa20symmetric.EncryptSymmetric(privKeyBytes, key)
}

// Unarmor and decrypt the private key.
func UnarmorDecryptPrivKey(armorStr string, passphrase string) (crypto.PrivKey, error) {
	var privKey crypto.PrivKey
	blockType, header, encBytes, err := armor.DecodeArmor(armorStr)
	if err != nil {
		return privKey, err
	}
	if blockType != blockTypePrivKey {
		return privKey, fmt.Errorf("unrecognized armor type: %v", blockType)
	}
	// Same as UnarmorPrivateKey
	if len(header) == 0 && passphrase == "" {
		return crypto.PrivKeyFromBytes(encBytes)
	}
	if header["kdf"] != "bcrypt" {
		return privKey, fmt.Errorf("unrecognized KDF type: %v", header["KDF"])
	}
	if header["salt"] == "" {
		return privKey, fmt.Errorf("missing salt bytes")
	}
	saltBytes, err := hex.DecodeString(header["salt"])
	if err != nil {
		return privKey, fmt.Errorf("error decoding salt: %w", err)
	}
	privKey, err = decryptPrivKey(saltBytes, encBytes, passphrase)
	return privKey, err
}

func decryptPrivKey(saltBytes []byte, encBytes []byte, passphrase string) (privKey crypto.PrivKey, err error) {
	key, err := bcrypt.GenerateFromPassword(saltBytes, []byte(passphrase), bcryptSecurityParameter)
	if err != nil {
		os.Exit("Error generating bcrypt key from passphrase: " + err.Error())
	}
	key = crypto.Sha256(key) // Get 32 bytes
	privKeyBytes, err := xsalsa20symmetric.DecryptSymmetric(encBytes, key)
	if err != nil && err.Error() == "ciphertext decryption failed" {
		return privKey, keyerror.NewErrWrongPassword()
	} else if err != nil {
		return privKey, err
	}
	privKey, err = crypto.PrivKeyFromBytes(privKeyBytes)
	return privKey, err
}

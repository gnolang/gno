package armor

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/armor"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

var emptyArmorHeader = map[string]string{}

// ArmorPrivateKey generates unencrypted armor for the
// given private key
func ArmorPrivateKey(privKey crypto.PrivKey) string {
	return armor.EncodeArmor(blockTypePrivKey, emptyArmorHeader, privKey.Bytes())
}

// UnarmorPrivateKey extracts the private key from the raw
// unencrypted armor
func UnarmorPrivateKey(armorStr string) (crypto.PrivKey, error) {
	// Decode the raw armor
	blockType, header, privKeyBytes, err := armor.DecodeArmor(armorStr)
	if err != nil {
		return nil, err
	}

	// Make sure it's a private key block
	if blockType != blockTypePrivKey {
		return nil, fmt.Errorf("unrecognized armor type: %v", blockType)
	}

	// Make sure the header is empty
	if len(header) > 0 {
		return nil, errors.New("non-empty private key header")
	}

	return crypto.PrivKeyFromBytes(privKeyBytes)
}

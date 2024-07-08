package crypto

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
)

func PrivKeyFromBytes(privKeyBytes []byte) (privKey PrivKey, err error) {
	err = amino.Unmarshal(privKeyBytes, &privKey)
	return
}

func PubKeyFromBytes(pubKeyBytes []byte) (pubKey PubKey, err error) {
	err = amino.Unmarshal(pubKeyBytes, &pubKey)
	return
}

package keys

import (
	"github.com/gnolang/gno/pkgs/crypto"
)

// KeyOutput defines a structure wrapping around an Info object used for output
// functionality.
type KeyOutput struct {
	Name      string                 `json:"name" yaml:"name"`
	Type      string                 `json:"type" yaml:"type"`
	Address   string                 `json:"address" yaml:"address"`
	PubKey    string                 `json:"pubkey" yaml:"pubkey"`
	Mnemonic  string                 `json:"mnemonic,omitempty" yaml:"mnemonic"`
	Threshold uint                   `json:"threshold,omitempty" yaml:"threshold"`
	PubKeys   []multisigPubKeyOutput `json:"pubkeys,omitempty" yaml:"pubkeys"`
}

// NewKeyOutput creates a default KeyOutput instance without Mnemonic, Threshold and PubKeys
func NewKeyOutput(name, keyType, address, pubkey string) KeyOutput {
	return KeyOutput{
		Name:    name,
		Type:    keyType,
		Address: address,
		PubKey:  pubkey,
	}
}

type multisigPubKeyOutput struct {
	Address string `json:"address" yaml:"address"`
	PubKey  string `json:"pubkey" yaml:"pubkey"`
	Weight  uint   `json:"weight" yaml:"weight"`
}

// Bech32KeysOutput returns a slice of KeyOutput objects, each with a Bech32
// prefix, given a slice of Info objects. It returns an error if any call to
// Bech32KeyOutput fails.
func Bech32KeysOutput(infos []Info) ([]KeyOutput, error) {
	kos := make([]KeyOutput, len(infos))
	for i, info := range infos {
		ko, err := Bech32KeyOutput(info)
		if err != nil {
			return nil, err
		}
		kos[i] = ko
	}

	return kos, nil
}

// Bech32KeyOutput create a KeyOutput with a Bech32 prefix. If the
// public key is a multisig public key, then the threshold and constituent
// public keys will be added.
func Bech32KeyOutput(keyInfo Info) (KeyOutput, error) {
	addr := keyInfo.GetPubKey().Address()
	pubs := crypto.PubKeyToBech32(keyInfo.GetPubKey())
	ko := NewKeyOutput(keyInfo.GetName(), keyInfo.GetType().String(), addr.String(), pubs)

	if mInfo, ok := keyInfo.(*multiInfo); ok {
		pubKeys := make([]multisigPubKeyOutput, len(mInfo.PubKeys))

		for i, pk := range mInfo.PubKeys {
			addr := pk.PubKey.Address()
			pubs := crypto.PubKeyToBech32(pk.PubKey)
			pubKeys[i] = multisigPubKeyOutput{addr.String(), pubs, pk.Weight}
		}

		ko.Threshold = mInfo.Threshold
		ko.PubKeys = pubKeys
	}

	return ko, nil
}

package crypto

import (
	"errors"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bech32"
)

func AddressToBech32(addr Address) string {
	bech32Addr, err := bech32.Encode(Bech32AddrPrefix(), addr[:])
	if err != nil {
		panic(err)
	}
	return bech32Addr
}

func AddressFromBech32(bech32str string) (Address, error) {
	bz, err := GetFromBech32(bech32str, Bech32AddrPrefix())
	if err != nil {
		return Address{}, err
	} else {
		return AddressFromBytes(bz), nil
	}
}

func PubKeyToBech32(pub PubKey) string {
	bech32PubKey, err := bech32.Encode(Bech32PubKeyPrefix(), pub.Bytes())
	if err != nil {
		panic(err)
	}
	return bech32PubKey
}

func PubKeyFromBech32(bech32str string) (pubKey PubKey, err error) {
	bz, err := GetFromBech32(bech32str, Bech32PubKeyPrefix())
	if err != nil {
		return PubKey(nil), err
	} else {
		err = amino.Unmarshal(bz, &pubKey)
		return
	}
}

// GetFromBech32 decodes a bytestring from a Bech32 encoded string.
func GetFromBech32(bech32str, prefix string) ([]byte, error) {
	if len(bech32str) == 0 {
		return nil, errors.New("decoding Bech32 failed: must provide a valid bech32 string")
	}

	hrp, bz, err := bech32.DecodeAndConvert(bech32str)
	if err != nil {
		return nil, err
	}

	if hrp != prefix {
		return nil, fmt.Errorf("invalid Bech32 prefix; expected %s, got %s", prefix, hrp)
	}

	return bz, nil
}

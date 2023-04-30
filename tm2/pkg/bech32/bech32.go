package bech32

import (
	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

// ConvertAndEncode encodes []byte to bech32.
// DEPRECATED use Encode
func ConvertAndEncode(hrp string, data []byte) (string, error) {
	converted, err := bech32.ConvertBits(data, 8, 5, true)
	if err != nil {
		return "", errors.Wrap(err, "encoding bech32 failed")
	}
	return bech32.Encode(hrp, converted)
}

func Encode(hrp string, data []byte) (string, error) {
	return ConvertAndEncode(hrp, data)
}

// DecodeAndConvert decodes bech32 to []byte.
// DEPRECATED use Decode
func DecodeAndConvert(bech string) (string, []byte, error) {
	hrp, data, err := bech32.DecodeNoLimit(bech)
	if err != nil {
		return "", nil, errors.Wrap(err, "decoding bech32 failed")
	}
	converted, err := bech32.ConvertBits(data, 5, 8, false)
	if err != nil {
		return "", nil, errors.Wrap(err, "decoding bech32 failed")
	}
	return hrp, converted, nil
}

func Decode(bech string) (string, []byte, error) {
	return DecodeAndConvert(bech)
}

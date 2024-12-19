package std

import "crypto/bech32"

type Address string // NOTE: bech32

func (a Address) String() string {
	return string(a)
}

// IsValid checks if the address is valid bech32 encoded string
func (a Address) IsValid() bool {
	_, _, ok := bech32.DecodeBech32(a)
	return ok
}

const RawAddressSize = 20

type RawAddress [RawAddressSize]byte

func EncodeBech32(prefix string, bz [20]byte) Address {
	b32, err := bech32.ConvertAndEncode(prefix, bytes[:])
	if err != nil {
		panic(err) // should not happen
	}
	return b32
}

func DecodeBech32(addr Address) (prefix string, bz [20]byte, ok bool) {
	prefix, bz, err := bech32.Decode(addr)
	if err != nil || len(bz) != 20 {
		return "", [20]byte{}, false
	}
	return prefix, [20]byte(bz), true
}

func convertAndEncode(hrp string, data []byte) (string, error) {
	converted, err := bech32.ConvertBits(data, 8, 5, true)
	if err != nil {
		return "", errors.Wrap(err, "encoding bech32 failed")
	}
	return bech32.Encode(hrp, converted)
}

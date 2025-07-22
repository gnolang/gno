package std

import (
	"crypto/bech32"
	"errors"
)

type Address = address // NOTE: bech32

// bech32AddrPrefix defines the Bech32 prefix of an address
const bech32AddrPrefix = "g"

const RawAddressSize = 20

type RawAddress [RawAddressSize]byte

func EncodeBech32(prefix string, bz [20]byte) Address {
	b32, err := convertAndEncode(prefix, bz[:])
	if err != nil {
		panic(err) // should not happen
	}
	return Address(b32)
}

func DecodeBech32(addr Address) (string, [20]byte, bool) {
	prefix, bz, err := decodeAndConvert(string(addr))
	if err != nil || len(bz) != 20 {
		return "", [20]byte{}, false
	}
	return prefix, convertTo20Byte(bz), true
}

func convertAndEncode(hrp string, data []byte) (string, error) {
	converted, err := bech32.ConvertBits(data, 8, 5, true)
	if err != nil {
		return "", errors.New("encoding bech32 failed: " + err.Error())
	}
	return bech32.Encode(hrp, converted)
}

func decodeAndConvert(bech string) (string, []byte, error) {
	hrp, data, err := bech32.DecodeNoLimit(bech)
	if err != nil {
		return "", nil, errors.New("decoding bech32 failed" + err.Error())
	}
	converted, err := bech32.ConvertBits(data, 5, 8, false)
	if err != nil {
		return "", nil, errors.New("decoding bech32 failed" + err.Error())
	}
	return hrp, converted, nil
}

func derivePkgAddr(string) string

// Returns a crypto hash derived pkgPath, unless pkgPath is a MsgRun run path,
// in which case the address is extracted from the path.
func DerivePkgAddr(pkgPath string) Address {
	addr := derivePkgAddr(pkgPath)
	return Address(addr)
}

func deriveStorageDepositAddr(string) string

// Returns a crypto hash derived pkgPath, unless pkgPath is a MsgRun run path,
// in which case the address is extracted from the path.
func DeriveStorageDepositAddr(pkgPath string) Address {
	addr := deriveStorageDepositAddr(pkgPath)
	return Address(addr)
}

func convertTo20Byte(in []byte) [20]byte {
	result := [20]byte{}
	copy(result[:], in)
	return result
}

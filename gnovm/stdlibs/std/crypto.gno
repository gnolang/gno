package std

import (
	"crypto/bech32"
	"errors"
)

type Address = address // NOTE: bech32

// Bech32AddrPrefix defines the Bech32 prefix of an address
const Bech32AddrPrefix = "g"

const RawAddressSize = 20

type RawAddress [RawAddressSize]byte

func EncodeBech32(prefix string, bz []byte) Address {
	b32, err := convertAndEncode(prefix, bz)
	if err != nil {
		panic(err) // should not happen
	}
	return Address(b32)
}

func convertAndEncode(hrp string, data []byte) (string, error) {
	converted, err := bech32.ConvertBits(data, 8, 5, true)
	if err != nil {
		return "", errors.New("encoding bech32 failed: " + err.Error())
	}
	return bech32.Encode(hrp, converted)
}

func DecodeBech32(bech string) (string, []byte, error) {
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

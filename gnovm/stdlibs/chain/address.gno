package chain

import (
	"crypto/bech32"
	"errors"
)

// bech32AddrPrefix defines the Bech32 prefix of an address
const bech32AddrPrefix = "g"

func EncodeBech32(prefix string, bz [20]byte) address {
	b32, err := convertAndEncode(prefix, bz[:])
	if err != nil {
		panic(err) // should not happen
	}
	return address(b32)
}

func DecodeBech32(addr address) (string, [20]byte, bool) {
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
func DerivePkgAddr(pkgPath string) address {
	addr := derivePkgAddr(pkgPath)
	return address(addr)
}

func deriveStorageDepositAddr(string) string

// Returns a crypto hash derived pkgPath, unless pkgPath is a MsgRun run path,
// in which case the address is extracted from the path.
func DeriveStorageDepositAddr(pkgPath string) address {
	addr := deriveStorageDepositAddr(pkgPath)
	return address(addr)
}

func convertTo20Byte(in []byte) [20]byte {
	result := [20]byte{}
	copy(result[:], in)
	return result
}

package std

import (
	"crypto/bech32"
	"crypto/ripemd160"
	"crypto/sha256"
	"errors"
	"strings"
)

type Address string // NOTE: bech32

// bech32AddrPrefix defines the Bech32 prefix of an address
const bech32AddrPrefix = "g"

func (a Address) String() string {
	return string(a)
}

// IsValid checks if the address is valid bech32 encoded string
func (a Address) IsValid() bool {
	_, _, ok := DecodeBech32(a)
	return ok
}

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

func PubkeyToAddress(publicKey string) (Address, error) {
	_, bz, err := decodeAndConvert(publicKey)
	if err != nil {
		return "", err
	}
	//XXX: the bytes we recover here are in the form:
	/// /tm.KEY_TYPE#\n!ACTUAL_KEY_BYES
	// so we need to split the string in order to get the actual key bytes
	// we also ensure that the type of the key is supported right now just PubKeySecp256k1
	parts := strings.Split(string(bz), "#\n!")
	keyType := parts[0]
	keyBytes := []byte(parts[1])

	if !strings.Contains(keyType, "/tm.PubKeySecp256k1") {
		return "", errors.New("invalid public key type")
	}

	sha := sha256.Sum256(keyBytes)
	hasherRIPEMD160 := ripemd160.New()
	hasherRIPEMD160.Write(sha[:])
	return EncodeBech32(bech32AddrPrefix, convertTo20Byte(hasherRIPEMD160.Sum(nil))), nil
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

func DerivePkgAddr(pkgPath string) Address {
	data := sumTruncated([]byte("pkgPath:" + pkgPath))
	return EncodeBech32(bech32AddrPrefix, data)
}

func sumTruncated(bz []byte) [20]byte {
	hash := sha256.Sum256(bz)
	return convertTo20Byte(hash[:RawAddressSize])
}

func convertTo20Byte(in []byte) [20]byte {
	/* For some reason [20]byte(bz) fails with: 'cannot convert []uint8 to ArrayKind'
	Maybe there is an issue to create
	*/
	result := [20]byte{}
	for index, b := range in {
		result[index] = b
	}

	return result
}

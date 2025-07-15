// Package keyderive provides utility functions for parsing Bech32-encoded public keys
// and deriving addresses from them. It offers a simple API to decode Bech32 representations
// of public keys into raw byte slices, and to compute corresponding addresses using
// standard hashing algorithms.
// Example usage:
//
//   pubKeyBytes, err := keyderive.ParseBech32PubKey(bech32Str)
//   if err != nil {
//       // handle error
//   }
//   address := keyderive.DeriveAddress(pubKeyBytes)

package keyderive

import (
	"crypto/ripemd160"
	"crypto/sha256"
	"errors"
	"std"
	"strings"
)

// Pubkey2Address format gpub1.....
func Pubkey2Address(publicKey string) (std.Address, error) {
	_, bz, err := std.DecodeBech32(publicKey)
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
	return std.EncodeBech32(std.Bech32AddrPrefix, hasherRIPEMD160.Sum(nil)), nil
}

// AddressMatchesKey returns if the publicKey matches the address passed as parameter
func AddressMatchesKey(address, publicKey string) (bool, error) {
	computedAddress, err := Pubkey2Address(publicKey)
	if err != nil {
		return false, err
	}
	return computedAddress.String() == address, nil
}

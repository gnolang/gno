// Package keyderive provides utility functions for parsing Bech32-encoded public keys
// and deriving addresses from them. It offers a simple API to decode Bech32 representations
// of public keys into raw byte slices, and to compute corresponding addresses using
// standard hashing algorithms.
// Example usage:
//
//   matched, err := keyderive.AddressMatchesKey("g1...", "gpub....")
//   if err != nil {
//       // handle error
//   }
//   address := keyderive.Pubkey2Address(gpub....)

package keyderive

import (
	"crypto/ripemd160"
	"crypto/sha256"
	"std"
	"strings"
)

// Pubkey2Address format gpub1.....
// Currently supported key formats are:
// - PubKeySecp256k1
// - PubKeyEd25519
//
// the code to obtain bech32 address from the publicKey bytes can be seen here:
// https://github.com/gnolang/gno/blob/34ba1efdb4fe0e23183a3524ff306a0429de9928/tm2/pkg/crypto/secp256k1/secp256k1.go#L122-L130
// https://github.com/gnolang/gno/blob/34ba1efdb4fe0e23183a3524ff306a0429de9928/tm2/pkg/crypto/ed25519/ed25519.go#L121-L123
func Pubkey2Address(publicKey string) (std.Address, error) {
	_, bz, err := std.DecodeBech32(publicKey)
	if err != nil {
		return "", err
	}
	//XXX: the bytes we recover here are in the form:
	/// /tm.KEY_TYPE#\n!ACTUAL_KEY_BYES
	// so in order to obtain the useful bytes we need to take the last N bytes
	// depending on the key type
	// for PubKeySecp256k1 (key size is 33 bytes), so we take the last 33 bytes
	// PubKeyEd25519 (key size is 32 bytes), so we take the last 32 bytes

	switch {
	case strings.Contains(string(bz), "/tm.PubKeySecp256k1"):
		keyBytes := bz[len(bz)-33:]
		sha := sha256.Sum256(keyBytes)
		hasherRIPEMD160 := ripemd160.New()
		hasherRIPEMD160.Write(sha[:])

		return std.EncodeBech32(std.Bech32AddrPrefix, hasherRIPEMD160.Sum(nil)), nil
	case strings.Contains(string(bz), "/tm.PubKeyEd25519"):
		keyBytes := bz[len(bz)-32:]
		sha := sha256.Sum256(keyBytes)
		return std.EncodeBech32(std.Bech32AddrPrefix, sha[:20]), nil
	default:
		panic("unsupported public key type")
	}
}

// AddressMatchesKey returns if the publicKey matches the address passed as parameter
func AddressMatchesKey(address, publicKey string) (bool, error) {
	computedAddress, err := Pubkey2Address(publicKey)
	if err != nil {
		return false, err
	}
	return computedAddress.String() == address, nil
}

// Package ctg is a simple utility package with helpers
// for bech32 address conversions.
package ctg

import (
	"crypto/bech32"

	"gno.land/p/nt/ufmt"
)

// ConvertCosmosToGno takes a Bech32 Cosmos address (prefix "cosmos")
// and returns the same address re-encoded with the gno.land prefix "g".
func ConvertCosmosToGno(addr string) (address, error) {
	prefix, decoded, err := bech32.Decode(addr)
	if err != nil {
		return "", ufmt.Errorf("bech32 decode failed: %v", err)
	}

	if prefix != "cosmos" {
		return "", ufmt.Errorf("expected a cosmos address, got prefix %q", prefix)
	}

	return address(mustEncode("g", decoded)), nil
}

func mustEncode(hrp string, data []byte) string {
	enc, err := bech32.Encode(hrp, data)
	if err != nil {
		panic(err)
	}
	return enc
}

// ConvertAnyToGno converts *any* valid Bech32 address to its gno.land form
// by preserving the underlying payload but replacing the prefix with "g".
// No prefix check is performed; invalid Bech32 input still returns an error.
func ConvertAnyToGno(addr string) (address, error) {
	_, decoded, err := bech32.Decode(addr)
	if err != nil {
		return "", ufmt.Errorf("bech32 decode failed: %v", err)
	}
	return address(mustEncode("g", decoded)), nil
}

// ConvertGnoToAny converts a gno.land address (prefixed with "g") to another Bech32
// prefix given by prefix. The function ensures the source address really
// is a gno.land address before proceeding.
//
// Example:
//
//	cosmosAddr, _ := ConvertGnoToAny("cosmos", "g1k98jx9...")
//	fmt.Println(cosmosAddr) // → cosmos1....
func ConvertGnoToAny(prefix string, addr address) (string, error) {
	origPrefix, decoded, err := bech32.Decode(string(addr))
	if err != nil {
		return "", ufmt.Errorf("bech32 decode failed: %v", err)
	}
	if origPrefix != "g" {
		return "", ufmt.Errorf("expected a gno address but got prefix %q", origPrefix)
	}
	return mustEncode(prefix, decoded), nil
}

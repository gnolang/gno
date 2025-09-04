// Package ctg is a simple utility package with helpers
// for bech32 address conversions.
package ctg

import (
	"chain"

	"gno.land/p/demo/ufmt"
)

// ConvertCosmosToGno takes a Bech32 Cosmos address (prefix "cosmos")
// and returns the same address re-encoded with the gno.land prefix "g".
func ConvertCosmosToGno(addr string) (chain.Address, error) {
	prefix, decoded, ok := chain.DecodeBech32(chain.Address(addr))
	if !ok {
		return "", ufmt.Errorf("bech32 decode failed")
	}

	if prefix != "cosmos" {
		return "", ufmt.Errorf("expected a cosmos address, got prefix %q", prefix)
	}

	return chain.EncodeBech32("g", decoded), nil
}

// ConvertAnyToGno converts *any* valid Bech32 address to its gno.land form
// by preserving the underlying payload but replacing the prefix with "g".
// No prefix check is performed; invalid Bech32 input still returns an error.
func ConvertAnyToGno(addr string) (chain.Address, error) {
	_, decoded, ok := chain.DecodeBech32(chain.Address(addr))
	if !ok {
		return "", ufmt.Errorf("bech32 decode failed")
	}
	return chain.EncodeBech32("g", decoded), nil
}

// ConvertGnoToAny converts a gno.land address (prefixed with "g") to another Bech32
// prefix given by prefix. The function ensures the source address really
// is a gno.land address before proceeding.
//
// Example:
//
//	cosmosAddr, _ := ConvertGnoToAny("cosmos", "g1k98jx9...")
//	fmt.Println(cosmosAddr) // → cosmos1....
func ConvertGnoToAny(prefix string, addr chain.Address) (string, error) {
	origPrefix, decoded, ok := chain.DecodeBech32(addr)
	if !ok {
		return "", ufmt.Errorf("bech32 decode failed")
	}
	if origPrefix != "g" {
		return "", ufmt.Errorf("expected a gno address but got prefix %q", origPrefix)
	}
	return chain.EncodeBech32(prefix, decoded).String(), nil
}

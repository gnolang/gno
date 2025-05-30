// Package ctg is a simple utility package that helps convert
// cosmos-prefixed addresses to gno addresses.
package ctg

import (
	"std"

	"gno.land/p/demo/ufmt"
)

// ConvertCosmosToGno takes a Bech32 Cosmos address (prefix "cosmos")
// and returns the same address re-encoded with the gno.land prefix "g".
func ConvertCosmosToGno(addr string) (std.Address, error) {
	prefix, decoded, ok := std.DecodeBech32(std.Address(addr))
	if !ok {
		return "", ufmt.Errorf("bech32 decode failed")
	}
	
	if prefix != "cosmos" {
		return "", ufmt.Errorf("expected a cosmos address, got prefix %q", prefix)
	}

	return std.EncodeBech32("g", decoded), nil
}

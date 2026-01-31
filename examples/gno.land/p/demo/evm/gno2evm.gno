package evm

import (
	"errors"
	"std"

	"gno.land/p/onbloc/uint256"
)

// Transform an Gno adress into it's unique Ethereum adress as uint256
func GnoAddressToUint256(gnoddr std.Address) (*uint256.Uint, error) {
	prefix, bytes, ok := std.DecodeBech32(gnoddr)
	if !ok || prefix != "g" {
		return nil, errors.New("invalid gno address")
	}
	return uint256.Zero().SetBytes20(bytes[:]), nil
}

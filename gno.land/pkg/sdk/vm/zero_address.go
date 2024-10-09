package vm

import "github.com/gnolang/gno/tm2/pkg/crypto"

const (
	zAddressBech32 = string("g1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqluuxe")
)

func ZeroAddress() crypto.Address {
	ZAddress, err := crypto.AddressFromBech32(zAddressBech32)
	if err != nil {
		panic(err)
	}
	return ZAddress
}

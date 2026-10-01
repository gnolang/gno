package testutils

import "crypto/bech32"

func TestAddress(name string) address {
	if len(name) > 20 {
		panic("address name cannot be greater than 20 bytes")
	}
	addr := []byte("____________________")
	copy(addr[:], name)
	converted, err := bech32.ConvertBits(addr, 8, 5, true)
	if err != nil {
		panic(err)
	}
	enc, err := bech32.Encode("g", converted)
	if err != nil {
		panic(err)
	}
	return address(enc)
}

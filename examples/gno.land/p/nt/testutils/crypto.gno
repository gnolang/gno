package testutils

import "chain"

func TestAddress(name string) address {
	if len(name) > 20 {
		panic("address name cannot be greater than 20 bytes")
	}
	addr := [20]byte{}
	// TODO: use strings.RepeatString or similar.
	// NOTE: I miss python's "".Join().
	blanks := "____________________"
	copy(addr[:], []byte(blanks))
	copy(addr[:], []byte(name))
	return address(chain.EncodeBech32("g", addr))
}

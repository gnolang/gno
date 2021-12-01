package testutils

import "std"

func TestAddress(name string) std.Address {
	if len(name) > std.AddressSize {
		panic("address name cannot be greater than std.AddressSize bytes")
	}
	addr := std.Address{}
	// TODO: use strings.RepeatString or similar.
	// NOTE: I miss python's "".Join().
	blanks := "____________________"
	copy(addr[:], []byte(blanks))
	copy(addr[:], []byte(name))
	return addr
}

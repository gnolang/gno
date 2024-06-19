func TestAddress(name string) std.Address {
	if len(name) > std.RawAddressSize {
		panic("address name cannot be greater than std.AddressSize bytes")
	}
	addr := std.RawAddress{}
	// TODO: use strings.RepeatString or similar.
	// NOTE: I miss python's "".Join().
	blanks := "____________________"
	copy(addr[:], []byte(blanks))
	copy(addr[:], []byte(name))
	return std.Address(std.EncodeBech32("g", addr))
}

package std

type Address string // NOTE: bech32

func (a Address) String() string {
	return string(a)
}

// Valid checks if the address is of specific length. Doesn't check prefix or checksum for the address
func (a Address) Valid() bool {
	return len(a) == RawAddressSize*2 // hex length
}

const RawAddressSize = 20

type RawAddress [RawAddressSize]byte

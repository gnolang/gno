package std

type Address string // NOTE: bech32

func (a Address) String() string {
	return string(a)
}

// IsValid checks if the address is valid bech32 encoded string
func (a Address) IsValid() bool {
	_, _, ok := decodeBech32(a.String())
	return ok
}

const RawAddressSize = 20

type RawAddress [RawAddressSize]byte

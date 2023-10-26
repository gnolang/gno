package std

type Address string // NOTE: bech32

func (a Address) String() string {
	return string(a)
}

// Valid prevents most human errors. Doesn't check prefix or checksum for the address
func (a Address) Valid() bool {
	if len(a) != 40 || len(a) == 0 {
		return false
	}
	return true
}

const RawAddressSize = 20

type RawAddress [RawAddressSize]byte

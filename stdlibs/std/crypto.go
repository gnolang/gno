package std

type Address string // NOTE: bech32

func (a Address) String() string {
	return string(a)
}

const RawAddressSize = 20

type RawAddress [RawAddressSize]byte

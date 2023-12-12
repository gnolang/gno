package std

type Address string // NOTE: bech32

func (a Address) String() string {
	return string(a)
}

// IsValid checks if the address is of specific length. Doesn't check prefix or checksum for the address
func (a Address) IsValid() bool {
	return len(a) == RawAddressSize*2 // hex length
}

const RawAddressSize = 20

type RawAddress [RawAddressSize]byte

type Addressable interface {
	// XXX document requirements.
	GetAddress() std.Address
}

// TODO implement as a native function,
// see DerivePkgAddr.
func AddrFromPkgPath(pkgPath string) Address {
	panic("not yet implemented")
}

// Addresses can also be derived from a subpath.
// This way anything with an address also has a URI path.
// TODO implement as a native function.
func AddrFromAddressable(pkgPath, subpath string) Address {
	// hash(pkgPath + "/" + subpath)
	panic("not yet implemented")
}

// A standard structure for storing Subpath and Address.
type SubpathAddress struct {
	Subpath string
	Address std.Address
}

// A standard structure for storing all addr related fields.
type FullAddress struct {
	PkgPath string
	Subpath string
	Address std.Address
}

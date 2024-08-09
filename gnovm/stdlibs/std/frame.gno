package std

// A Realm represents an on-chain entity which has an address, can execute code
// and is capable of managing Coins associated with the given address.
// It boils down to one of two kinds of realms:
//
//   - User realms: an end-user of the blockchain, who has a private key from
//     which the public key and address are subsequently derived.
//   - Code realms: a package published on-chain, with storage through globlal
//     variables and access to the [Banker]. It is Gno's specific flavour of a
//     "smart contract".
//
// In many instances, when simply saying a "realm", the latter is intended;
// however, users are also considered "realms".
//
// When using the Realm type, use [Realm.Addr] to get its address. This is
// always set. If the Realm is a code realm, [Realm.PkgPath] will return its
// pkgpath and [Realm.IsUser] will return false.
type Realm struct {
	addr    Address
	pkgPath string
}

func (r Realm) Addr() Address {
	return r.addr
}

func (r Realm) PkgPath() string {
	return r.pkgPath
}

func (r Realm) IsUser() bool {
	return r.pkgPath == ""
}

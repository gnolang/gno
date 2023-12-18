package std

import "github.com/gnolang/gno/tm2/pkg/crypto"

type Realm struct {
	addr    crypto.Bech32Address
	pkgPath string
}

func (r Realm) Addr() crypto.Bech32Address {
	return r.addr
}

func (r Realm) PkgPath() string {
	return r.pkgPath
}

func (r Realm) IsUser() bool {
	return r.pkgPath == ""
}

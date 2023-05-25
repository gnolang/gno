package stdlibs

import "github.com/gnolang/gno/tm2/pkg/crypto"

type Realm struct {
	Addr    crypto.Bech32Address
	PkgPath string
}

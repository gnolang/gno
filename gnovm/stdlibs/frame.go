package stdlibs

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

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

// isOriginCall returns true if the
func isOriginCall(m *gno.Machine) bool {
	return prevRealm(m).addr == m.Context.(ExecContext).OrigCaller
}

// prevRealm loops on frames and returns the second realm found in the calling
// order. If no such realm was found, returns the tx signer.
func prevRealm(m *gno.Machine) Realm {
	var lastRealmPath string
	for i := m.NumFrames() - 1; i > 0; i-- {
		fr := m.Frames[i]
		if fr.LastPackage == nil || !fr.LastPackage.IsRealm() {
			// Ignore non-realm frame
			continue
		}
		realmPath := fr.LastPackage.PkgPath
		if lastRealmPath == "" {
			// Record the path of the first ecountered realm and continue
			lastRealmPath = realmPath
			continue
		}
		if lastRealmPath != realmPath {
			// Second realm detected, return it.
			return Realm{
				addr:    fr.LastPackage.GetPkgAddr().Bech32(),
				pkgPath: realmPath,
			}
		}
	}
	// No second realm found, return the tx signer.
	return Realm{
		addr:    m.Context.(ExecContext).OrigCaller,
		pkgPath: "", // empty for users
	}
}

package stdlibs

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

type Realm struct {
	addr crypto.Bech32Address
	path string
}

func (r Realm) Addr() crypto.Bech32Address {
	return r.addr
}

func (r Realm) Path() string {
	return r.path
}

func (r Realm) IsUser() bool {
	return r.path == ""
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
				addr: fr.LastPackage.GetPkgAddr().Bech32(),
				path: realmPath,
			}
		}
	}
	// No second realm found, return the tx signer.
	return Realm{
		addr: m.Context.(ExecContext).OrigCaller,
		path: "", // empty for users
	}
}

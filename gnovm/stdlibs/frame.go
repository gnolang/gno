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

func prevRealm(m *gno.Machine) (crypto.Bech32Address, string) {
	var (
		ctx = m.Context.(ExecContext)
		// Default lastCaller is OrigCaller, the signer of the tx
		lastCaller  = ctx.OrigCaller
		lastPkgPath = ""
	)

	for i := m.NumFrames() - 1; i > 0; i-- {
		fr := m.Frames[i]
		if fr.LastPackage == nil || !fr.LastPackage.IsRealm() {
			// Ignore non-realm frame
			continue
		}
		pkgPath := fr.LastPackage.PkgPath
		// The first realm we encounter will be the one calling
		// this function; to get the calling realm determine the first frame
		// where fr.LastPackage changes.
		if lastPkgPath == "" {
			lastPkgPath = pkgPath
		} else if lastPkgPath == pkgPath {
			continue
		} else {
			lastCaller = fr.LastPackage.GetPkgAddr().Bech32()
			lastPkgPath = pkgPath
			break
		}
	}

	// Empty the pkgPath if we return a user
	if ctx.OrigCaller == lastCaller {
		lastPkgPath = ""
	}
	return lastCaller, lastPkgPath
}

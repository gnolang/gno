package runtime

import "chain"

// NewUserRealm creates a new "user" realm with the given address.
func NewUserRealm(user address) Realm {
	return Realm{addr: user}
}

// NewCodeRealm creates a new realm for a published code realm with the given
// pkgPath.
func NewCodeRealm(pkgPath string) Realm {
	if !isRealm(pkgPath) {
		panic("should only be called for Realms")
	}
	return Realm{pkgPath: pkgPath, addr: chain.DerivePkgAddr(pkgPath)}
}

func AssertOriginCall() // injected

// native bindings
func getRealm(height int) (address string, pkgPath string)
func isRealm(pkgPath string) bool

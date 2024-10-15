package std

// NewUserRealm creates a new "user" realm with the given address.
func NewUserRealm(user Address) Realm {
	return Realm{addr: user}
}

// NewCodeRealm creates a new realm for a published code realm with the given
// pkgPath.
func NewCodeRealm(pkgPath string) Realm {
	return Realm{pkgPath: pkgPath, addr: DerivePkgAddr(pkgPath)}
}

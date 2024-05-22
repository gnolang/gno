package std

func NewUserRealm(user Address) Realm {
	return Realm{addr: user}
}

func NewCodeRealm(pkgPath string) Realm {
	return Realm{pkgPath: pkgPath, addr: DerivePkgAddr(pkgPath)}
}

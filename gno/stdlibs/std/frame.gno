package std

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

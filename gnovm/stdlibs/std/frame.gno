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

func (r Realm) Denom(denom string) string {
	return RealmDenom(r.pkgPath, denom)
}

func RealmDenom(pkgPath, denom string) string {
	// TODO: Possibly remove after https://github.com/gnolang/gno/issues/3164
	// Similar to ibc spec
	// ibc_denom := 'ibc/' + hash('path' + 'base_denom')
	// gno_realm_denom := '/' + 'pkg_path' + ':' + 'base_denom'
	return "/" + pkgPath + ":" + denom
}

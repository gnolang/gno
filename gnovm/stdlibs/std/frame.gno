package std

type Realm struct {
	addr    Address
	pkgPath string
}

func (r Realm) String() string {
	if r.pkgPath == "" {
		return "UserRealm{ " + r.addr.String() + " }"
	} else {
		return "CodeRealm{ " + r.addr.String() + ", " + r.pkgPath + " }"
	}
}

func (r Realm) Address() Address {
	return r.addr
}

func (r Realm) PkgPath() string {
	return r.pkgPath
}

func (r Realm) IsUser() bool {
	return r.pkgPath == ""
}

func (r Realm) CoinDenom(coinName string) string {
	return CoinDenom(r.pkgPath, coinName)
}

func CoinDenom(pkgPath, coinName string) string {
	// TODO: Possibly remove after https://github.com/gnolang/gno/issues/3164
	// Similar to ibc spec
	// ibc_denom := 'ibc/' + hash('path' + 'base_denom')
	// gno_qualified_denom := '/' + 'pkg_path' + ':' + 'base_denom'
	return "/" + pkgPath + ":" + coinName
}

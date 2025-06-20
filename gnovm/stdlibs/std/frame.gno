package std

import "regexp"

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

var reGnoRunPath = regexp.MustCompile(`^([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+\.[a-zA-Z]{2,}/r/(?P<addr>g1[a-z0-9]+)/run$`)

func (r Realm) IsUser() bool {
	// non-code realm.
	if r.pkgPath == "" {
		return true
	}

	// `maketx run`
	// TODO: replace with the upcoming /e/ ephemeral realms
	if reGnoRunPath.MatchString(r.pkgPath) {
		return true
	}

	return false
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

package root

import "std"

var (
	counter     int
	currentImpl = "gno.land/r/x/manfred_upgrade_patterns/upgrade_c/v1"
)

func Inc(nb int) int {
	assertIsCurrentImpl()
	counter += nb
	return counter
}

func Get() int {
	return counter
}

func SetCurrentImpl(pkgpath string) {
	assertIsAdmin()
	currentImpl = pkgpath
}

func assertIsCurrentImpl() {
	if std.PreviousRealm().PkgPath() != currentImpl {
		panic("unauthorized")
	}
}

func assertIsAdmin() {
	// TODO
}

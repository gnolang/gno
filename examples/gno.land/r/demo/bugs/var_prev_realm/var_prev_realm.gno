package var_prev_realm

import (
	"std"

	"gno.land/r/demo/tests"
)

var (
	varRealm  = tests.GetPrevRealm()
	initRealm std.Realm
)

func init() {
	initRealm = tests.GetPrevRealm()
}

func Render(path string) string {
	s := ""
	s += "Var block realm:\n\nPackage path: " + varRealm.PkgPath() + "\n\nAddress: " + varRealm.Addr().String() + "\n\n"
	s += "Init block realm:\n\nPackage path: " + initRealm.PkgPath() + "\n\nAddress: " + initRealm.Addr().String() + "\n\n"
	return s
}

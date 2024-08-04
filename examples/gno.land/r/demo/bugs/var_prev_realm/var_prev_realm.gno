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

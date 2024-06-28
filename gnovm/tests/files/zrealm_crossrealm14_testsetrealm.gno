// PKGPATH: gno.land/r/demo/tests_test

package tests_test

import (
	"std"

	"gno.land/r/demo/tests"
)

func main() {
	println(tests.GetPrevRealm())

	std.TestSetRealm(std.NewCodeRealm("gno.land/r/demo/users"))
	println(tests.GetPrevRealm())

	std.TestSetRealm(std.NewUserRealm("g1alice"))
	println(tests.GetPrevRealm())
}

// Output:
// PORCAMADONNA

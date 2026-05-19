// PKGPATH: gno.land/r/tests/vm0
package tests0

import (
	tests "gno.land/r/tests/vm"
	"gno.land/r/tests/vm/tests_foo"
)

func init(cur realm) {
	tests_foo.AddFooStringer(0, cur, "one")
	tests_foo.AddFooStringer(0, cur, "two")
}

func main(cur realm) {
	tests_foo.AddFooStringer(0, cur, "three")
	println(tests.Render(""))
	println("end")
}

// Error:
// expected package name [vm0] but got [tests0]

// TypeCheckError:
// gno.land/r/tests/vm0/.gnobuiltins.gno:1:1: package vm0; expected package tests0; gno.land/r/tests/vm0/zrealm_tests0.gno:9:21: undefined: cross; gno.land/r/tests/vm0/zrealm_tests0.gno:9:21: undefined: cross; gno.land/r/tests/vm0/zrealm_tests0.gno:14:21: undefined: cross; gno.land/r/tests/vm0/zrealm_tests0.gno:14:21: undefined: cross

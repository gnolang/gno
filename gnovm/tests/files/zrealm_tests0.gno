// PKGPATH: gno.land/r/tests/vm0
package tests0

import (
	tests "gno.land/r/tests/vm"
	"gno.land/r/tests/vm/tests_foo"
)

func init() {
	tests_foo.AddFooStringer("one")
	tests_foo.AddFooStringer("two")
}

func main(cur realm) {
	tests_foo.AddFooStringer("three")
	println(tests.Render(""))
	println("end")
}

// Error:
// invalid mempackage: package name "tests0" does not match path element "vm0"

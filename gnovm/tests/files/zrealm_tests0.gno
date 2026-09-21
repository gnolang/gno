// PKGPATH: gno.land/r/tests/tests0
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

// Output:
// 0: &FooStringer{one}
// 1: &FooStringer{two}
// 2: &FooStringer{three}
//
// end

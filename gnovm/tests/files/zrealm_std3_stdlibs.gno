// PKGPATH: gno.land/r/std_test
package std_test

import (
	"std"
)

func foo() {
	println("foo", std.CurrentRealm().PkgPath())
}

func main() {
	println("main", std.CurrentRealm().PkgPath())
	foo()
}

// Output:
// main gno.land/r/std_test
// foo gno.land/r/std_test

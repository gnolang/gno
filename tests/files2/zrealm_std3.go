// PKGPATH: gno.land/r/std_test
package std_test

import (
	"std"
)

func foo() {
	println("foo", std.CurrentRealmPath())
}

func main() {
	println("main", std.CurrentRealmPath())
	foo()
}

// Output:
// main gno.land/r/std_test
// foo gno.land/r/std_test

// PKGPATH: gno.land/r/std_test
package std_test

import (
	"gno.land/p/tests" // NOTE: not a realm path.
)

func main() {
	println("test1", tests.CurrentRealmPath())
	func() {
		println("test2", tests.CurrentRealmPath())
	}()
	x := tests.CurrentRealmPath
	println("test3", x())
}

// Output:
// test1 gno.land/r/std_test
// test2 gno.land/r/std_test
// test3 gno.land/r/std_test

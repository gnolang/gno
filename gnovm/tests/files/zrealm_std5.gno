// PKGPATH: gno.land/r/std_test
package std_test

import (
	"gno.land/p/demo/tests" // NOTE: not a realm path.
)

func main() {
	crossing()

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

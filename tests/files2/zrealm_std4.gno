// PKGPATH: gno.land/r/std_test
package std_test

import (
	"gno.land/r/tests"
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
// test1 gno.land/r/tests
// test2 gno.land/r/tests
// test3 gno.land/r/tests

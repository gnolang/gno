// PKGPATH: gno.land/r/std_test
package std_test

import (
	"gno.land/r/demo/tests"
)

func main() {
	crossing()

	println("test1", cross(tests.CurrentRealmPath)())
	func() {
		println("test2", cross(tests.CurrentRealmPath)())
	}()
	x := tests.CurrentRealmPath
	println("test3", cross(x)())
}

// Output:
// test1 gno.land/r/demo/tests
// test2 gno.land/r/demo/tests
// test3 gno.land/r/demo/tests

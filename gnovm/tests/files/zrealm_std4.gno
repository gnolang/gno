// PKGPATH: gno.land/r/std_test
package std_test

import (
	"gno.land/r/demo/tests"
)

func main() {
	switchrealm()

	println("test1", withswitch(tests.CurrentRealmPath)())
	func() {
		println("test2", withswitch(tests.CurrentRealmPath)())
	}()
	x := tests.CurrentRealmPath
	println("test3", withswitch(x)())
}

// Output:
// test1 gno.land/r/demo/tests
// test2 gno.land/r/demo/tests
// test3 gno.land/r/demo/tests

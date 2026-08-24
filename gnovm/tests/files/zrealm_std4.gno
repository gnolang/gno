// PKGPATH: gno.land/r/std
package std

import (
	tests "gno.land/r/tests/vm"
)

func main(cur realm) {
	println("test1", tests.CurrentRealmPath(cross(cur)))
	func() {
		println("test2", tests.CurrentRealmPath(cross(cur)))
	}()
	x := tests.CurrentRealmPath
	println("test3", x(cross(cur)))
}

// Output:
// test1 gno.land/r/tests/vm
// test2 gno.land/r/tests/vm
// test3 gno.land/r/tests/vm

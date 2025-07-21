// PKGPATH: gno.land/r/std
package std

import (
	"gno.land/p/demo/tests" // NOTE: not a realm path.
)

func main(cur realm) {
	println("test1", tests.CurrentRealmPath())
	func() {
		println("test2", tests.CurrentRealmPath())
	}()
	x := tests.CurrentRealmPath
	println("test3", x())
}

// Output:
// test1 gno.land/r/std
// test2 gno.land/r/std
// test3 gno.land/r/std

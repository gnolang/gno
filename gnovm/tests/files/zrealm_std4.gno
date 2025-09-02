// PKGPATH: gno.land/r/std
package std

import (
	"gno.land/r/tests"
)

func main(cur realm) {
	println("test1", tests.CurrentRealmPath(cross))
	func() {
		println("test2", tests.CurrentRealmPath(cross))
	}()
	x := tests.CurrentRealmPath
	println("test3", x(cross))
}

// Output:
// test1 gno.land/r/tests
// test2 gno.land/r/tests
// test3 gno.land/r/tests

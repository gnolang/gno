// PKGPATH: gno.land/r/std
package std

import (
	"std"
)

func foo() {
	println("foo", std.CurrentRealm().PkgPath())
}

func main(cur realm) {
	println("main", std.CurrentRealm().PkgPath())
	foo()
}

// Output:
// main gno.land/r/std
// foo gno.land/r/std

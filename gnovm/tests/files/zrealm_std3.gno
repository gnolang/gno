// PKGPATH: gno.land/r/std
package std

import (
	"chain/runtime/unsafe"
)

func foo() {
	println("foo", unsafe.CurrentRealm().PkgPath())
}

func main(cur realm) {
	println("main", unsafe.CurrentRealm().PkgPath())
	foo()
}

// Output:
// main gno.land/r/std
// foo gno.land/r/std

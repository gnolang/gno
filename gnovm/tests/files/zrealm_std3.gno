// PKGPATH: gno.land/r/std
package std

import "chain/runtime"

func foo() {
	println("foo", runtime.CurrentRealm().PkgPath())
}

func main(cur realm) {
	println("main", runtime.CurrentRealm().PkgPath())
	foo()
}

// Output:
// main gno.land/r/std
// foo gno.land/r/std

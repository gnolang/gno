// PKGPATH: gno.land/r/std_test
package std_test

import "chain/runtime"

func foo() {
	println("foo", runtime.CurrentRealm().PkgPath())
}

func main() {
	println("main", runtime.CurrentRealm().PkgPath())
	foo()
}

// Output:
// main gno.land/r/std_test
// foo gno.land/r/std_test

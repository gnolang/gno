// PKGPATH: gno.land/r/type_alias_test
package type_alias_test

import "gno.land/p/demo/uassert"

type TestingT = uassert.TestingT

func main() {
	crossing()

	println("ok")
}

// Output:
// ok

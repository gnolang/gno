// PKGPATH: gno.land/r/testutils_test
package testutils_test

import (
	"gno.land/p/demo/testutils"
)

func main() {
	crossing()

	println(testutils.TestAddress("test1"))
	println(testutils.TestAddress("test2"))
}

// Output:
// g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7
// g1w3jhxapjta047h6lta047h6lta047h6laqcyu4

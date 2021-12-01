// PKGPATH: gno.land/r/testutils_test
package testutils_test

import (
	"gno.land/p/testutils"
)

func main() {
	println(testutils.TestAddress("test1"))
	println(testutils.TestAddress("test2"))
}

// Output:
// array[0x74657374315F5F5F5F5F5F5F5F5F5F5F5F5F5F5F]
// array[0x74657374325F5F5F5F5F5F5F5F5F5F5F5F5F5F5F]

// PKGPATH: gno.land/r/access_test
package access_test

import (
	"gno.land/p/testutils"
)

func main() {
	testutils.TestVar1 += 1
	println(testutils.TestVar1)
}

// Error:
// cannot modify external-realm or non-realm object
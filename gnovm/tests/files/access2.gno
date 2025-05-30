// PKGPATH: gno.land/r/access_test
package access_test

import (
	"gno.land/p/demo/testutils"
)

func main() {
	crossing()

	testutils.TestVar1 += 1
	println(testutils.TestVar1)
}

// Error:
// cannot directly modify readonly tainted object (w/o method): (const (ref(gno.land/p/demo/testutils) package{})).TestVar1

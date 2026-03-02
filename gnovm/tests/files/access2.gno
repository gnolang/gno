// PKGPATH: gno.land/r/access
package access

import (
	"gno.land/p/nt/testutils/v0"
)

func main(cur realm) {
	testutils.TestVar1 += 1
	println(testutils.TestVar1)
}

// Error:
// gno.land/r/access/access2.gno:9:2-20: cannot directly mutate gno.land/p/nt/testutils/v0.TestVar1 from gno.land/r/access

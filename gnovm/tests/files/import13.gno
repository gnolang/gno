// PKGPATH: gno.land/p/demo

package demo

import (
	"filetests/extern/foo/internal/aa"
)

func main() {
	println(aa.A)
}

// Error:
// gno.land/p/demo/import13.gno:6:2-36: internal/ packages can only be imported by packages rooted at the parent of "internal"

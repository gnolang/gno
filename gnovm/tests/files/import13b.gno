// PKGPATH: gno.land/p/demo

package demo

import (
	"filetests/extern/foo/internal"
)

func main() {
	println(internal.A)
}

// Error:
// gno.land/p/demo/import13b.gno:6:2-33: internal/ packages can only be imported by packages rooted at the parent of "internal"

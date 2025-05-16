// PKGPATH: gno.land/p/demo

package demo

import (
	"github.com/gnolang/gno/_test/foo/internal/aa"
)

func main() {
	println(aa.A)
}

// Error:
// gno.land/p/demo/files/import13.gno:6:2: internal/ packages can only be imported by packages rooted at the parent of "internal"

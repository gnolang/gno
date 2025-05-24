// PKGPATH: gno.land/p/demo

package demo

import (
	"github.com/gnolang/gno/_test/foo/internal"
)

func main() {
	println(internal.A)
}

// Error:
// gno.land/p/demo/import13b.gno:6:2-45: internal/ packages can only be imported by packages rooted at the parent of "internal"

// PKGPATH: gno.land/p/demo/bar
package bar

import (
	_ "gno.land/r/tests/vm"
)

func main() {
	println("dontcare")
}

// Error:
// gno.land/p/demo/bar/import11.gno:5:2-25: pure package path "gno.land/p/demo/bar" cannot import realm path "gno.land/r/tests/vm"

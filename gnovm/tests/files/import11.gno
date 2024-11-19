// PKGPATH: gno.land/p/demo/bar
package bar

import (
	"gno.land/r/demo/tests"
)

func main() {
	println(tests.Counter())
}

// Error:
// gno.land/p/demo/bar/files/import11.gno:5:2: should not happen p/ cannot import r/: package path gno.land/p/demo/bar imports realm path gno.land/r/demo/tests

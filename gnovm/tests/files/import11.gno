// PKGPATH: gno.land/p/demo/bar
package bar

import (
	"gno.land/r/tests/tests"
)

func main() {
	println(tests.Counter(cross))
}

// Error:
// gno.land/p/demo/bar/import11.gno:5:2-26: pure package path "gno.land/p/demo/bar" cannot import realm path "gno.land/r/tests/tests"

// PKGPATH: gno.land/p/demo/bar
package bar

import (
	"gno.land/r/demo/tests"
)

func main() {
	println(tests.Counter(cross))
}

// Error:
// gno.land/p/demo/bar/import11.gno:5:2-25: pure package path "gno.land/p/demo/bar" cannot import realm path "gno.land/r/demo/tests"

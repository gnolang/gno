// PKGPATH: gno.land/p/demo/bar
package bar

import (
	tests "gno.land/r/tests/vm"
)

func main() {
	println(tests.Counter(cross))
}

// Error:
// gno.land/p/demo/bar/import11.gno:5:2-29: pure package path "gno.land/p/demo/bar" cannot import realm path "gno.land/r/tests/vm"

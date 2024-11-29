// PKGPATH: gno.land/r/demo/boards2_test
package boards2_test

import "gno.land/r/demo/boards2"

func main() {
	boards2.CreateBoard("foo")
}

// Error:
// invalid non-user call

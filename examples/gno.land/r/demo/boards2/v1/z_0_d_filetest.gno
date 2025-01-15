// PKGPATH: gno.land/r/demo/boards2/v1_test
package v1_test

import (
	boards2 "gno.land/r/demo/boards2/v1"
)

func main() {
	boards2.CreateBoard("foo")
}

// Error:
// invalid non-user call

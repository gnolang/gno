// PKGPATH: gno.land/r/nt/boards2/v1_test
package v1_test

import (
	boards2 "gno.land/r/nt/boards2/v1"
)

func main() {
	boards2.CreateBoard("foo123")
}

// Error:
// invalid non-user call

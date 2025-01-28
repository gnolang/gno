// PKGPATH: gno.land/r/nt/boards2_test
package boards2_test

import "gno.land/r/nt/boards2"

func main() {
	boards2.CreateBoard("foo123")
}

// Error:
// invalid non-user call

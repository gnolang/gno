// PKGPATH: gno.land/r/std_test
package std_test

import (
	"gno.land/r/demo/tests"
)

func main() {
	defer tests.PrintCurrentRealmPath()
	tests.PrintCurrentRealmPath()
}

// Output:
// gno.land/r/demo/tests
// gno.land/r/demo/tests

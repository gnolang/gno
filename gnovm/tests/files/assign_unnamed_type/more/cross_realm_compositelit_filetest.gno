// PKGPATH: gno.land/r/declaredtype_test
package declaredtype_test

import (
	"gno.land/r/demo/tests"
)

func main() {
	crossing()

	println(tests.GetZeroType())
}

// Output:
// (readonly(slice[(0 gno.land/r/demo/tests.Word)]) gno.land/r/demo/tests.nat)

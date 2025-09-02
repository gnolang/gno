// PKGPATH: gno.land/r/declaredtype
package declaredtype

import (
	"gno.land/r/tests/tests"
)

func main() {
	println(tests.GetZeroType())
}

// Output:
// (readonly(slice[(0 gno.land/r/tests/tests.Word)]) gno.land/r/tests/tests.nat)

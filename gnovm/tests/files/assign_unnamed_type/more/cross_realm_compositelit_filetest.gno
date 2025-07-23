// PKGPATH: gno.land/r/declaredtype
package declaredtype

import (
	"gno.land/r/demo/tests"
)

func main() {
	println(tests.GetZeroType())
}

// Output:
// (readonly(slice[(0 gno.land/r/demo/tests.Word)]) gno.land/r/demo/tests.nat)

// PKGPATH: gno.land/r/declaredtype
package declaredtype

import (
	tests "gno.land/r/tests/vm"
)

func main() {
	println(tests.GetZeroType())
}

// Output:
// (readonly(slice[(0 gno.land/r/tests/vm.Word)]) gno.land/r/tests/vm.nat)

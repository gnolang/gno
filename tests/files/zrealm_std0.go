// PKGPATH: gno.land/r/std_test
package std_test

import (
	"std"
)

func main() {
	caller := std.GetCaller()
	println(caller)
}

// Output:
// array[0x7465737461646472746573746164647274657374]

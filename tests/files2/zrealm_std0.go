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
// array[0x74657374616464725F5F5F5F5F5F5F5F5F5F5F5F]

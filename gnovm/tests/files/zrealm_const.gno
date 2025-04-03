// PKGPATH: gno.land/r/const_test
package const_test

import "std"

func main() {
	switchrealm()

	println(std.RawAddressSize)
}

// Output:
// 20

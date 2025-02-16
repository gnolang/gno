// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm/func"

var f func() string

func main() {
	f = crossrealm.GetMethod2()
	println(f)
}

// Error:
// cannot attach a value of a type defined by another realm: struct{("struct_val" string)}

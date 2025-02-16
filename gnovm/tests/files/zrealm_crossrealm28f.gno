// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

type foo struct {
	name string
}

func (foo) Foo() { println("hello from crossrealm_test") }

var local_S []int

func init() {
	local_S = crossrealm.GetSlice10()
}

func main() {
	println("ok")
}

// Output:
// ok

// Realm:
// switchrealm["gno.land/r/crossrealm_test"]

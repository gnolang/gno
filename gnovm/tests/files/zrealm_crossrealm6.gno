// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/p/demo/tests"
)

var somevalue tests.TestRealmObject2

func init() {
	somevalue.Field = "test"
}

func main(cur realm) {
	// this is OK because the method is declared in a non-realm package.
	somevalue.Modify()
	println(somevalue)
}

// Output:
// (struct{("modified" string)} gno.land/p/demo/tests.TestRealmObject2)

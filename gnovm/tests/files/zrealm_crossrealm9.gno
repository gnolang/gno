// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/p/demo/tests/v0"
)

func main(cur realm) {
	tests.SomeValue2.Field = "modified"
	println("done")
}

// Error:
// cannot directly modify readonly tainted object (write it in a named function of the realm that owns it, during a crossing call into that realm; a /p/ package's object cannot be written after its init): (const (ref(gno.land/p/demo/tests/v0) package{})).SomeValue2.Field

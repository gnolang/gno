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
// cannot directly modify readonly tainted object (use a method or crossing function): (const (ref(gno.land/p/demo/tests/v0) package{})).SomeValue2.Field

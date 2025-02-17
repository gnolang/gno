// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/func"
)

var b = 1

type Local_S struct {
	name string
}

func main() {
	ls := &Local_S{name: "Local_S"}

	f := func() bool {
		c := b
		return true
	}
	crossrealm.SetCallback(f)

	ls.name = "modified"

	println("ok")
}

// Error:
// cannot attach a reference to an unreal object from an external realm: heapitem((struct{("Local_S" string)} gno.land/r/crossrealm_test.Local_S))

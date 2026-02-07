// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

type S struct {
	v *int
}

var v S

func main(cur realm) {
	println("main()")
	v.v = new(int)
	crossrealm_b.SetObject(cross, &v)
	println("done")
}

// Output:
// main()
// done

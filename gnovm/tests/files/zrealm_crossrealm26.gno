// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

type Struct struct {
	A int
}

func main(cur realm) {
	s := &Struct{A: 100}
	crossrealm_b.SetObject(cross, func() { println(&s.A) })
	s.A = 123
	println(s)
}

// Error:
// cannot directly modify readonly tainted object (w/o method): s<~VPBlock(1,1)>.A

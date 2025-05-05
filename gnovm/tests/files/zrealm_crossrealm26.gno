// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/r/demo/tests/crossrealm_b"
)

type Struct struct {
	A int
}

func main() {
	crossing()

	s := &Struct{A: 100}
	cross(crossrealm_b.SetObject)(func() { println(&s.A) })
	s.A = 123
	println(s)
}

// Error:
// cannot directly modify readonly tainted object (w/o method): s<~VPBlock(1,0)>.A

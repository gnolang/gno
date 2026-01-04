// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

type Struct struct {
	A int
}

func main(cur realm) {
	println("s := &Struct{A: 100}")
	s := &Struct{A: 100}
	println("crossrealm_b.SetObject(cross, &s.A)")
	crossrealm_b.SetObject(cross, &s.A)
	println("s.A = 123")
	s.A = 123
	println(s)
}

// Output:
// s := &Struct{A: 100}
// crossrealm_b.SetObject(cross, &s.A)
// s.A = 123

// Error:
// cannot directly modify readonly tainted object (w/o method): s<~VPBlock(1,1)>.A

// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

type Struct struct {
	A int
}

// var s *int
var s_g *Struct

func init() {
	println("s := &Struct{A: 100}")
	s := &Struct{A: 100}
	println("crossrealm_b.SetObject(cross, &s.A)")
	crossrealm_b.SetObject(cross, &s.A)
	println("s_g = s")
	s_g = s
}

func main(cur realm) {
	println("s_g.A = 123")
	s_g.A = 123
	println(s_g)
}

// Output:
// s := &Struct{A: 100}
// crossrealm_b.SetObject(cross, &s.A)
// s_g = s
// s_g.A = 123

// Error:
// cannot directly modify readonly tainted object (w/o method): s_g<~VPBlock(3,1)>.A

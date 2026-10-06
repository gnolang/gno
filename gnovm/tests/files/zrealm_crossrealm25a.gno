// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

type Struct struct {
	A int
}

var ourS *Struct

func init(cur realm) {
	println("s = &Struct{A: 100}")
	s := &Struct{A: 100}
	println("crossrealm_b.SetObject(cross, &s.A)")
	crossrealm_b.SetObject(cross(cur), &s.A)
	println("ourS = s")
	ourS = s
}

func main(cur realm) {
	println("crossrealm_b.IncrementObject(cross)")
	println(crossrealm_b.IncrementObject(cross(cur)))
	println(ourS.A)
}

// Output:
// s = &Struct{A: 100}
// crossrealm_b.SetObject(cross, &s.A)
// ourS = s
// crossrealm_b.IncrementObject(cross)

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): *(ptr<VPBlock(1,2)>)

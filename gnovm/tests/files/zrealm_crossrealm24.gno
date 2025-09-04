// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

type Struct struct {
	A int
}

func main(cur realm) {
	a := 1
	s := &Struct{A: 100}
	crossrealm_b.SetClosure(cross, func() {})
	crossrealm_b.SetObject(cross, &a)
	//cross(crossrealm_b.SetObject)(&s.A)
	//cross(crossrealm_b.SetObject)(func() { &s.A })
	s.A = 123
	println(s, a)
}

// Output:
// &(struct{(123 int)} gno.land/r/crossrealm.Struct) 1

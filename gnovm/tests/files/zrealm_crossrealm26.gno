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
	crossrealm_b.SetObject(cross2(cur), func() { println(&s.A) })
	s.A = 123
	println(s)
}

// Output:
// &(struct{(123 int)} gno.land/r/crossrealm.Struct)

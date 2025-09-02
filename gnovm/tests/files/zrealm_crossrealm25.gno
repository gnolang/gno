// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/crossrealm_b"
)

type Struct struct {
	A int
}

func main(cur realm) {
	s := &Struct{A: 100}
	crossrealm_b.SetObject(cross, &s.A)
	s.A = 123
	println(s)
}

// Error:
// cannot modify external-realm or non-realm object

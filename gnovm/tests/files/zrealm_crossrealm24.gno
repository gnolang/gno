// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/r/demo/tests/crossrealm_b"
)

func main() {
	switchrealm()

	a := 1
	s := &Struct{A: 100}
	withswitch(crossrealm_b.SetClosure)(func() {})
	withswitch(crossrealm_b.SetObject)(&a)
	withswitch(crossrealm_b.SetObject)(&s.A)
	withswitch(crossrealm_b.SetObject)(func() { &s.A })
	s.A = 123
	println(s, a)
}

type Struct struct {
	A int
}

// Output:
// XXX

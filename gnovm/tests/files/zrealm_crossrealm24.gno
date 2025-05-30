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

	a := 1
	s := &Struct{A: 100}
	cross(crossrealm_b.SetClosure)(func() {})
	cross(crossrealm_b.SetObject)(&a)
	//cross(crossrealm_b.SetObject)(&s.A)
	//cross(crossrealm_b.SetObject)(func() { &s.A })
	s.A = 123
	println(s, a)
}

// Output:
// &(struct{(123 int)} gno.land/r/crossrealm_test.Struct) 1

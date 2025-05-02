// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/r/demo/tests/crossrealm_b"
)

type Struct struct {
	A int
}

var root interface{}

func main() {
	crossing()

	s := &Struct{A: 100}
	cross(crossrealm_b.SetObject)(func() { println(&s.A) })

	cross(crossrealm_b.SetObject)(nil)

	root = s

	s.A = 123
	println(s)
}

// Output:
// &(struct{(123 int)} gno.land/r/crossrealm_test.Struct)

// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/r/demo/tests/crossrealm_b"
)

type Struct struct {
	A int
}

var s = &Struct{A: 100}

func main() {
	crossing()

	cross(crossrealm_b.SetObject)(func() { println(&s.A) })
	s.A = 123
	println(s)
}

// Output:
// &(struct{(123 int)} gno.land/r/crossrealm_test.Struct)

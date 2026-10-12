package tests_foo

import (
	tests "gno.land/r/tests/vm"
)

// for testing gno.land/r/tests/vm/interfaces.go

type FooStringer struct {
	FieldA string
}

func (fs *FooStringer) String() string {
	return "&FooStringer{" + fs.FieldA + "}"
}

// AddFooStringer is a non-crossing helper. Callers thread their own
// live cur as `rlm`; `cross(rlm)` forwards it into the (cur realm)-
// crossing tests.AddStringer.
func AddFooStringer(_ int, rlm realm, fa string) {
	tests.AddStringer(cross(rlm), &FooStringer{fa})
}

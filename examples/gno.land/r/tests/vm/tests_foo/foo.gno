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

func AddFooStringer(fa string) {
	tests.AddStringer(cross, &FooStringer{fa})
}

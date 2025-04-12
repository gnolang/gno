package tests_foo

import (
	"gno.land/r/demo/tests"
)

// for testing gno.land/r/demo/tests/interfaces.go

type FooStringer struct {
	FieldA string
}

func (fs *FooStringer) String() string {
	return "&FooStringer{" + fs.FieldA + "}"
}

func AddFooStringer(fa string) {
	cross(tests.AddStringer)(&FooStringer{fa})
}

package tests_foo

import (
	"gno.land/r/tests"
)

// for testing gno.land/r/tests/interfaces.go

type FooStringer struct {
	FieldA string
}

func (fs *FooStringer) String() string {
	return "&FooStringer{" + fs.FieldA + "}"
}

func AddFooStringer(fa string) {
	tests.AddStringer(cross, &FooStringer{fa})
}

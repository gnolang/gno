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

var f1 = &FooStringer{"one"}   // this make it real
var f2 = &FooStringer{"two"}   // this make it real
var f3 = &FooStringer{"three"} // this make it real

func AddFooStringer(fa string) {
	switch fa {
	case "one":
		tests.AddStringer(f1)
	case "two":
		tests.AddStringer(f2)
	case "three":
		tests.AddStringer(f3)
	}
}

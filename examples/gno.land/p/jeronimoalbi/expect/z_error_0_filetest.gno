package main

import (
	"errors"
	"strings"

	"gno.land/p/jeronimoalbi/expect"
)

var (
	output  strings.Builder
	t       = expect.MockTestingT(&output)
	testErr = errors.New("test")
)

func main() {
	expect.Func(t, func() error {
		return testErr
	}).ToFail().WithMessage("foo")

	expect.Func(t, func() error {
		return testErr
	}).ToFail().Not().WithMessage("test")

	println(output.String())
}

// Output:
// Expected error message to match
// Got: test
// Want: foo
// Expected error message to be different
// Got: test

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
	}).ToFail().WithError(errors.New("foo"))

	expect.Func(t, func() error {
		return testErr
	}).ToFail().Not().WithError(testErr)

	println(output.String())
}

// Output:
// Expected errors to match
// Got: test
// Want: foo
// Expected errors to be different
// Got: test

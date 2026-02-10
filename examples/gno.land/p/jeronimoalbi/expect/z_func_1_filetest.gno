package main

import (
	"errors"
	"strings"

	"gno.land/p/jeronimoalbi/expect"
)

var (
	output strings.Builder
	t      = expect.MockTestingT(&output)
)

func main() {
	msg := "Boom!"
	err := errors.New(msg)

	expect.Func(t, func() error { return err }).Not().ToFail()
	expect.Func(t, func() error { return err }).ToFail().Not().WithMessage(msg)
	expect.Func(t, func() error { return err }).ToFail().Not().WithError(err)

	expect.Func(t, func() (any, error) { return nil, err }).Not().ToFail()
	expect.Func(t, func() (any, error) { return nil, err }).ToFail().Not().WithMessage(msg)
	expect.Func(t, func() (any, error) { return nil, err }).ToFail().Not().WithError(err)

	println(output.String())
}

// Output:
// Func failed with error
// Got: Boom!
// Expected error message to be different
// Got: Boom!
// Expected errors to be different
// Got: Boom!
// Func failed with error
// Got: Boom!
// Expected error message to be different
// Got: Boom!
// Expected errors to be different
// Got: Boom!

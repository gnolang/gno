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
	gotMsg := "Boom!"
	gotErr := errors.New(gotMsg)
	wantMsg := "Tick Tock"
	wantErr := errors.New(wantMsg)

	expect.Func(t, func() error { return nil }).ToFail()
	expect.Func(t, func() error { return gotErr }).ToFail().WithMessage(wantMsg)
	expect.Func(t, func() error { return gotErr }).ToFail().WithError(wantErr)

	expect.Func(t, func() (any, error) { return nil, nil }).ToFail()
	expect.Func(t, func() (any, error) { return nil, gotErr }).ToFail().WithMessage(wantMsg)
	expect.Func(t, func() (any, error) { return nil, gotErr }).ToFail().WithError(wantErr)

	expect.Func(t, func() int { return 0 }).ToFail()

	println(output.String())
}

// Output:
// Expected func to return an error
// Expected error message to match
// Got: Boom!
// Want: Tick Tock
// Expected errors to match
// Got: Boom!
// Want: Tick Tock
// Expected func to return an error
// Expected error message to match
// Got: Boom!
// Want: Tick Tock
// Expected errors to match
// Got: Boom!
// Want: Tick Tock
// Unsupported error func type
// Got: unknown

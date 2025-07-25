package main

import (
	"strings"

	"gno.land/p/jeronimoalbi/expect"
)

var (
	output strings.Builder
	t      = expect.MockTestingT(&output)
)

func main() {
	gotMsg := "Boom!"
	wantMsg := "Tick Tock"

	expect.Func(t, func() {}).ToPanic()
	expect.Func(t, func() { panic(gotMsg) }).ToPanic().WithMessage(wantMsg)

	expect.Func(t, func() error { return nil }).ToPanic()
	expect.Func(t, func() error { panic(gotMsg) }).ToPanic().WithMessage(wantMsg)

	expect.Func(t, func() any { return nil }).ToPanic()
	expect.Func(t, func() any { panic(gotMsg) }).ToPanic().WithMessage(wantMsg)

	expect.Func(t, func() (any, error) { return nil, nil }).ToPanic()
	expect.Func(t, func() (any, error) { panic(gotMsg) }).ToPanic().WithMessage(wantMsg)

	expect.Func(t, func() int { return 0 }).ToPanic()

	println(output.String())
}

// Output:
// Expected function to panic
// Expected panic message to match
// Got: Boom!
// Want: Tick Tock
// Expected function to panic
// Expected panic message to match
// Got: Boom!
// Want: Tick Tock
// Expected function to panic
// Expected panic message to match
// Got: Boom!
// Want: Tick Tock
// Expected function to panic
// Expected panic message to match
// Got: Boom!
// Want: Tick Tock
// Unsupported func type
// Got: unknown

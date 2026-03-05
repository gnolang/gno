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
	msg := "Boom!"

	expect.Func(t, func() { panic(msg) }).Not().ToPanic()
	expect.Func(t, func() { panic(msg) }).ToPanic().Not().WithMessage(msg)

	expect.Func(t, func() error { panic(msg) }).Not().ToPanic()
	expect.Func(t, func() error { panic(msg) }).ToPanic().Not().WithMessage(msg)

	expect.Func(t, func() any { panic(msg) }).Not().ToPanic()
	expect.Func(t, func() any { panic(msg) }).ToPanic().Not().WithMessage(msg)

	expect.Func(t, func() (any, error) { panic(msg) }).Not().ToPanic()
	expect.Func(t, func() (any, error) { panic(msg) }).ToPanic().Not().WithMessage(msg)

	println(output.String())
}

// Output:
// Expected func not to panic
// Got: Boom!
// Expected panic message to be different
// Got: Boom!
// Expected func not to panic
// Got: Boom!
// Expected panic message to be different
// Got: Boom!
// Expected func not to panic
// Got: Boom!
// Expected panic message to be different
// Got: Boom!
// Expected func not to panic
// Got: Boom!
// Expected panic message to be different
// Got: Boom!

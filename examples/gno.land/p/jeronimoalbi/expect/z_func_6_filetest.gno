// PKGPATH: gno.land/r/demo/test
package test

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/jeronimoalbi/expect"
)

const (
	caller = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
	msg    = "Boom!"
)

var (
	output strings.Builder
	t      = expect.MockTestingT(&output)
)

func Fail() {
	crossing()

	panic(msg)
}

func Success() {
	crossing()

	// No panic
}

func main() {
	testing.SetRealm(std.NewUserRealm(caller))

	expect.Func(t, func() { cross(Fail)() }).ToCrossPanic()
	expect.Func(t, func() { cross(Fail)() }).ToCrossPanic().WithMessage(msg)

	expect.Func(t, func() error { cross(Fail)(); return nil }).ToCrossPanic()
	expect.Func(t, func() error { cross(Fail)(); return nil }).ToCrossPanic().WithMessage(msg)

	expect.Func(t, func() any { cross(Fail)(); return nil }).ToCrossPanic()
	expect.Func(t, func() any { cross(Fail)(); return nil }).ToCrossPanic().WithMessage(msg)

	expect.Func(t, func() (any, error) { cross(Fail)(); return nil, nil }).ToCrossPanic()
	expect.Func(t, func() (any, error) { cross(Fail)(); return nil, nil }).ToCrossPanic().WithMessage(msg)

	expect.Func(t, func() { cross(Success)() }).Not().ToCrossPanic()
	expect.Func(t, func() error { cross(Success)(); return nil }).Not().ToCrossPanic()
	expect.Func(t, func() any { cross(Success)(); return nil }).Not().ToCrossPanic()
	expect.Func(t, func() (any, error) { cross(Success)(); return nil, nil }).Not().ToCrossPanic()

	// None should fail, output should be empty
	print(output.String())
}

// Output:

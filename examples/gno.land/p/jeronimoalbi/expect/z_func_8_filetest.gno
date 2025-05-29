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

func main() {
	testing.SetRealm(std.NewUserRealm(caller))

	expect.Func(t, func() { cross(Fail)() }).Not().ToCrossPanic()
	expect.Func(t, func() { cross(Fail)() }).ToCrossPanic().Not().WithMessage(msg)

	expect.Func(t, func() error { cross(Fail)(); return nil }).Not().ToCrossPanic()
	expect.Func(t, func() error { cross(Fail)(); return nil }).ToCrossPanic().Not().WithMessage(msg)

	expect.Func(t, func() any { cross(Fail)(); return nil }).Not().ToCrossPanic()
	expect.Func(t, func() any { cross(Fail)(); return nil }).ToCrossPanic().Not().WithMessage(msg)

	expect.Func(t, func() (any, error) { cross(Fail)(); return nil, nil }).Not().ToCrossPanic()
	expect.Func(t, func() (any, error) { cross(Fail)(); return nil, nil }).ToCrossPanic().Not().WithMessage(msg)

	println(output.String())
}

// Output:
// Expected func not to cross panic
// Got: Boom!
// Expected cross panic message to be different
// Got: Boom!
// Expected func not to cross panic
// Got: Boom!
// Expected cross panic message to be different
// Got: Boom!
// Expected func not to cross panic
// Got: Boom!
// Expected cross panic message to be different
// Got: Boom!
// Expected func not to cross panic
// Got: Boom!
// Expected cross panic message to be different
// Got: Boom!
